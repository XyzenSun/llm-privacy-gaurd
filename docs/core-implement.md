# 核心实现

> 本仓库仅分享解决方案的设计思路，并非完整的生产级实现。前端页面仅供体验功能使用。

## 数据流

```
┌──────────┐   原文            ┌──────────────┐   占位符 prompt         ┌──────────┐
│  用户    │ ───────────────▶ │LLM Privacy Guard   │ ─────────────────────▶ │ 云端 LLM │
│          │ ◀─────────────── │               │ ◀─────────────────────│  (不可信) │
└──────────┘   还原后展示      └──────────────┘   占位符 response
                                      │
                                      ▼
                               ┌───────────────┐
                               │ 可信 AI (本地)│
                               └───────────────┘
```

核心原则：
- 云端只见语义化占位符（`${IP_ADDRESS_1}` / `${PASSWORD_2}`）
- 可信 AI 只做 NER，不参与业务对话，不接收完整历史
- 用户只见原文，占位符由本地映射表还原
- 映射表随 `session_id` 生命周期共存，持久化到数据库

---

## 可信 AI 调用

```
messages = [
    { role: "system",  content: <systemPrompt> },
    { role: "user",    content: <构建的 JSON 指令> }
]
response_format = { type: "json_object" }
temperature = 0
```

用户消息构造：
```
buildTrustedInstructions(knownMappings, recentContext, newMessage):
    return  "Return a JSON object with key 'entries'. "
            "known_mappings=" + JSON(knownMappings) + " "
            "new_message="     + JSON_escape(newMessage)
```

关键：`known_mappings` JSON 传入 → 可信 AI 知道哪些已映射，不重复标注，实现增量识别。

---

## 映射表

两张对称哈希表，持久化在 `mask_mappings` 表：

```
MaskMapping:
    session_id  TEXT
    original    TEXT  -- 原值
    placeholder TEXT  -- 占位符 ${TYPE_N}
    type        TEXT
    unique(session_id, original)    -- 反脱敏唯一
    unique(session_id, placeholder)  -- 脱敏唯一
```

运行时：
```
knownMappings : original → placeholder   # Mask
byPlaceholder : placeholder → original    # Unmask
```

### Mask

```
function applyKnownMappings(text, mappings):
    items = sort(mappings.items(), key = len(original), desc)  # 长度降序，防部分覆盖
    for (orig, ph) in items:
        text = text.replaceAll(orig, ph)
    return text
```

### Unmask

```
function unmask(text, byPlaceholder):
    items = sort(byPlaceholder.items(), key = len(placeholder), desc)
    for (ph, orig) in items:
        text = text.replaceAll(ph, orig)
    return text
```

### 一致性保证

可信 AI 对同一 original 给出不同 placeholder → 拒绝该轮返回，保持双射稳定：

```
function validateEntries(entries, knownMappings):
    for e in entries:
        if e.original in knownMappings:
            if knownMappings[e.original] != e.placeholder:
                return nil   # 冲突，整轮回滚
    return delta
```

---

## 增量调用

可信 AI 只看本轮新 prompt，历史用已有映射表告知。

```
function Mask(sessionID, content, knownMappings, turnID):
    preMasked = applyKnownMappings(content, knownMappings)

    if looksFullyCovered(preMasked):   # 正则快速判断
        return { maskedContent: preMasked }

    entries = trustedLLM.Detect(content, knownMappings)
    delta = validateEntries(entries, knownMappings)

    if delta is empty:
        return { maskedContent: preMasked }

    updated = merge(knownMappings, delta)
    finalMasked = applyKnownMappings(content, updated)
    mappingRepo.UpsertMany(delta)

    return { maskedContent: finalMasked, newEntries: delta }
```

快速启发式：
```
function looksFullyCovered(preMasked):
    patterns = [email, phone, idcard, bank, ipv4]
    for p in patterns:
        if matchPattern(preMasked, p): return false
    return true
```

---

## 四层MASK机制

```
Level 1: 原值 → 占位符 map 本地 replace     最快，纯 CPU
Level 2: looksFullyCovered 正则快判         跳过可信 AI
Level 3: Message.MaskedContent 列持久化     历史不再重算
Level 4: 可信 AI 真正被调用                 最贵
```

### 双列表消息持久化

```
Message:
    content         -- 展示给用户的原文
    masked_content  -- 发给云端的占位符文本
```

发送云端时：
```
function buildCloudMessages(history, newMaskedContent):
    out = []
    for m in history:
        out.append({ role: m.role, content: m.masked_content or m.content })
    out.append({ role: "user", content: newMaskedContent })
    return out
```

---

## 端到端流程

```
function Process(req):
    mappings = mappingRepo.ListBySession(session.id)
    history  = messageRepo.ListBySession(session.id)

    maskResult = Mask(session.id, req.message, mappings)
    cloudMessages = buildCloudMessages(history, maskResult.maskedContent)
    cloudResp = cloudLLM.Chat(cloudMessages)

    unmaskedResp = Unmask(cloudResp, byPlaceholder)

    messageRepo.Create(user, content=req.message, masked_content=maskResult.maskedContent)
    messageRepo.Create(assistant, content=unmaskedResp, masked_content=cloudResp)

    return { response: unmaskedResp }
```

---

## 失败处理

| 场景 | 兜底 |
|------|------|
| 可信 AI 返回非 JSON | 降级宽松解析，失败则报错 |
| 可信 AI 返回冲突 placeholder | 整轮不更新 map |
| 云端 LLM 失败 | 整轮 abort，下次重试等价 |
document.addEventListener('DOMContentLoaded', function() {
    var sessionId = null;
    var sessionList = document.getElementById('session-list');
    var chatWindow = document.getElementById('chat-window');
    var messageInput = document.getElementById('message-input');
    var sendBtn = document.getElementById('send-btn');
    var newSessionBtn = document.getElementById('new-session-btn');
    var logoutLink = document.getElementById('logout-link');

    var promptBefore = document.getElementById('prompt-before');
    var promptAfter = document.getElementById('prompt-after');
    var byPlaceholder = document.getElementById('by-placeholder');
    var cloudMasked = document.getElementById('cloud-masked');
    var cloudUnmasked = document.getElementById('cloud-unmasked');

    if (window.marked && typeof window.marked.setOptions === 'function') {
        window.marked.setOptions({
            gfm: true,
            breaks: true,
            headerIds: false,
            mangle: false
        });
    }

    loadSessions();

    newSessionBtn.addEventListener('click', createSession);
    sendBtn.addEventListener('click', sendMessage);
    logoutLink.addEventListener('click', logout);

    messageInput.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });

    function loadSessions() {
        fetch('/api/v1/sessions')
        .then(function(res) { return res.json(); })
        .then(function(data) {
            sessionList.innerHTML = '';
            (data.sessions || []).forEach(function(s) {
                var li = document.createElement('li');
                li.textContent = s.title || 'Session ' + s.id.substring(0, 8);
                li.dataset.id = s.id;
                li.addEventListener('click', function() {
                    selectSession(s.id, li);
                });
                sessionList.appendChild(li);
            });
        })
        .catch(function(err) { console.error('load sessions:', err); });
    }

    function createSession() {
        fetch('/api/v1/sessions', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({title: 'New Session'})
        })
        .then(function(res) { return res.json(); })
        .then(function(data) {
            if (data.id) {
                sessionId = data.id;
                chatWindow.innerHTML = '';
                loadSessions();
            }
        })
        .catch(function(err) { console.error('create session:', err); });
    }

    function selectSession(id, element) {
        sessionId = id;
        document.querySelectorAll('#session-list li').forEach(function(li) {
            li.classList.remove('active');
        });
        element.classList.add('active');
        loadMessages(id);
    }

    function loadMessages(sid) {
        fetch('/api/v1/sessions/' + sid + '/messages')
        .then(function(res) { return res.json(); })
        .then(function(data) {
            chatWindow.innerHTML = '';
            (data.messages || []).forEach(function(m) {
                appendMessage(m.role, m.content);
            });
            chatWindow.scrollTop = chatWindow.scrollHeight;
        })
        .catch(function(err) { console.error('load messages:', err); });
    }

    function appendMessage(role, content, opts) {
        var div = document.createElement('div');
        div.className = 'message ' + role;
        var asMarkdown = (opts && opts.markdown) || (role === 'assistant');
        if (asMarkdown && window.marked && window.DOMPurify) {
            var html = window.marked.parse(content || '');
            div.innerHTML = window.DOMPurify.sanitize(html);
            div.classList.add('markdown-body');
        } else {
            div.textContent = content;
        }
        chatWindow.appendChild(div);
        chatWindow.scrollTop = chatWindow.scrollHeight;
        return div;
    }

    function appendPendingAssistant() {
        var div = document.createElement('div');
        div.className = 'message assistant pending';
        div.innerHTML = '<span class="pending-label">思考中</span>' +
            '<span class="pending-dots"><i></i><i></i><i></i></span>';
        chatWindow.appendChild(div);
        chatWindow.scrollTop = chatWindow.scrollHeight;
        return div;
    }

    function sendMessage() {
        if (sendBtn.disabled) return;
        if (!sessionId) {
            alert('请先创建或选择一个会话');
            return;
        }
        var text = messageInput.value.trim();
        if (!text) return;

        appendMessage('user', text);
        messageInput.value = '';
        sendBtn.disabled = true;
        var pending = appendPendingAssistant();

        fetch('/api/v1/chat', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({sessionId: sessionId, message: text})
        })
        .then(function(res) { return res.json(); })
        .then(function(data) {
            if (pending && pending.parentNode) { pending.parentNode.removeChild(pending); }
            if (data.error) {
                appendMessage('assistant', 'Error: ' + data.error);
            } else {
                appendMessage('assistant', data.response || data.unmaskedResponse || '');
                updateDebug(data);
            }
            sendBtn.disabled = false;
        })
        .catch(function(err) {
            if (pending && pending.parentNode) { pending.parentNode.removeChild(pending); }
            appendMessage('assistant', 'Error: ' + err.message);
            sendBtn.disabled = false;
        });
    }

    function updateDebug(data) {
        promptBefore.textContent = data.promptBeforeMask || '';
        promptAfter.textContent = data.promptAfterMask || '';
        byPlaceholder.textContent = JSON.stringify(data.byPlaceholder || {}, null, 2);
        cloudMasked.textContent = data.cloudResponseMasked || '';
        cloudUnmasked.textContent = data.unmaskedResponse || data.response || '';
    }

    function logout(e) {
        e.preventDefault();
        fetch('/api/v1/auth/logout', {method: 'POST'})
        .then(function() { window.location.href = '/login.html'; })
        .catch(function() { window.location.href = '/login.html'; });
    }
});
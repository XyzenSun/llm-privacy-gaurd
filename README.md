English | [中文](README_zh-cn.md)

## Quick Start

The fastest way to experience LLM Privacy Guard is using Docker:

```bash
docker run -d \
  --name llm-privacy-guard \
  -p 19999:19999 \
  -e APP_PORT=19999 \
  -e APP_AUTHTOKEN=your-token-here \
  -e DB_TYPE=sqlite \
  -e DB_DSN=/app/data/llm_guard.db \
  -e APP_PROJECT_ROOT=/app \
  -v llm-privacy-data:/app/data \
  ghcr.io/xyzensun/llm-privacy-guard:latest
```

Or use Docker Compose:

```bash
# Download docker-compose.yml
curl -O https://raw.githubusercontent.com/XyzenSun/llm-privacy-gaurd-easy-try/master/docker-compose.yml

# Set your auth token
export APP_AUTHTOKEN=your-token-here

# Start the service
docker compose up -d
```

Then visit http://localhost:19999/login.html and login with your token.

**Note**: All configuration can be set via environment variables. No `.env` file required.

### Background and Motivation

When using AI services, many providers explicitly state that they store user data or use it for training (some providers claim they don't store user data, yet user chat logs have been found in leaked databases).

This means our conversation records with AI are visible to providers, introducing a **security concern**: we can easily leak private information during AI conversations.

*   **Example Scenario**: "Help me write an SSH login command, IP is `123.123.123.123`, password is `123456`."

While technically we could use jump servers or similar solutions, not everyone has this awareness, and not all AI users have a technical background. If providers extract our data or sell it to malicious actors, it could lead to personal information exposure or even financial theft.

Therefore, we need a way to ensure personal data security when using LLMs. This project takes a different approach from technical methods like data isolation or internal network segmentation — see below for the implementation details.

## LLM Privacy Guard - Core Design Philosophy

### Core Role Definitions

This project defines two roles:

1.  **Trusted LLM**: Locally deployed, or a provider you trust (e.g., Google Vertex AI's commitment that data is not stored and immediately deleted).
2.  **Cloud LLM**: The LLM API providers we commonly use, who explicitly state they use user data, or don't clearly state whether they use user data, or claim not to use user data but cannot be verified.

---

### Why Choose a Hybrid "Trusted LLM + Cloud LLM" Approach?

You might think: if cloud LLMs could leak data, why not just deploy locally? However, there are issues:

*   **Hardware Resource Constraints**: Deploying an AI model locally requires powerful hardware. For example, deploying a large model like GLM-5 requires multiple compute card arrays — most ordinary users don't have this capability.
*   **Closed-Source Models**: Many models are not open-source and cannot be deployed locally.

So, rather than giving up cloud LLMs entirely, we should combine them with trusted LLMs:

1.  In the **Trusted LLM**, mask private data in conversations.
2.  Send the masked conversation data to the **Cloud LLM**.
3.  After the Cloud LLM returns results, the **Trusted LLM** replaces and restores the private data, then returns it to the user.

This way, both user privacy and AI usability are maintained.

---

### Limitations of Traditional Data Masking Methods

Since the goal is to prevent privacy leaks, why not use word libraries (like ToolGood.Words) to remove sensitive information, or use reversible encryption algorithms to replace private information in prompts, then post-process to restore encrypted information in responses?

Although this approach has mature production-grade services, it has one drawback: **After encryption, can AI still understand your intent and achieve your purpose?**

Taking "generate SSH login command" as an example, compare the original prompt with the MD5-encrypted version:

#### Original Prompt Effect
*   **Prompt**: Help me write an SSH login command, IP is `123.123.123.123`, password is `123456`.
*   **LLM Response**:
    > On Linux or macOS terminals, the most common SSH login method is as follows. Note: Standard SSH commands don't support including passwords directly in the command for security reasons...
    > ```bash
    > ssh root@123.123.123.123
    > ```

#### Encrypted Prompt Effect
*   **Prompt**: Help me write an SSH login command, IP is `CCDD8D3D940A01B2FB3258C059924C0D`, password is `E10ADC3949BA59ABBE56E057F20F883E`.
*   **LLM Response**:
    > Regarding IP: `CCDD8D3D940A01B2FB3258C059924C0D` is a 32-bit hexadecimal string. A normal IP should be in a format like `192.168.1.1`...
    > If this is a special system environment... the command is:
    > ```bash
    > ssh root@CCDD8D3D940A01B2FB3258C059924C0D
    > ```

**Conclusion**: This difference occurs because encryption algorithms lose the original semantic information, affecting the LLM's generated results.

---

### Core Approach of This Project: Semantic Replacement

We let a trusted LLM understand the semantics of the original prompt and perform **semantic replacement**. For example:

*   Replace `My IP is 123.123.123.123` with `My IP is {IP_ADDR}`.

This way, the cloud LLM can understand our intent and provide correct responses.

#### Limitations and Countermeasures:

1.  **Comprehension Limitations**: The trusted LLM has limited understanding capability, and semantic replacement may not be accurate.
2.  **Cloud LLM Confusion About Semantic Variables**: The cloud LLM seeing `{IP_ADDR}` might think it's a variable and respond with "I see the IP address you provided is a variable."
    *   **Countermeasure**: Explain in the **System Prompt** sent to the cloud LLM: *"Values like `{$}` are variables. You don't need to prompt me to replace them. In your output, just output the corresponding variables without explanation."*

### Detailed Implementation
See [./docs/core-implement.md](./docs/core-implement.md)

## License
This project is licensed under the MIT License
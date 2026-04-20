package service

import "llm-privacy-gaurd/internal/llm"

type LiveTrustedLLMClient struct {
	runtime *LLMRuntimeService
}

type LiveCloudLLMClient struct {
	runtime *LLMRuntimeService
}

func NewLiveTrustedLLMClient(runtime *LLMRuntimeService) *LiveTrustedLLMClient {
	return &LiveTrustedLLMClient{runtime: runtime}
}

func NewLiveCloudLLMClient(runtime *LLMRuntimeService) *LiveCloudLLMClient {
	return &LiveCloudLLMClient{runtime: runtime}
}

func (c *LiveTrustedLLMClient) Detect(prompt string, knownMappings map[string]string, recentContext []llm.Message, newMessage string) ([]llm.MaskEntry, error) {
	client, err := c.runtime.TrustedClient()
	if err != nil {
		return nil, err
	}
	return client.Detect(prompt, knownMappings, recentContext, newMessage)
}

func (c *LiveTrustedLLMClient) Model() string {
	return c.runtime.TrustedModelName()
}

func (c *LiveCloudLLMClient) Chat(messages []llm.Message) (string, error) {
	client, err := c.runtime.CloudClient()
	if err != nil {
		return "", err
	}
	return client.Chat(messages)
}

func (c *LiveCloudLLMClient) Model() string {
	return c.runtime.CloudModelName()
}

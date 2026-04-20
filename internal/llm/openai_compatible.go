package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAICompatibleClient struct {
	client       *openai.Client
	model        string
	systemPrompt string
}

type openAIChatRequest struct {
	Model       string            `json:"model"`
	Messages    []Message         `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	ResponseFmt map[string]string `json:"response_format,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type detectEnvelope struct {
	Entries []MaskEntry `json:"entries"`
}

func NewOpenAICompatibleClient(baseURL, apiKey, model, systemPrompt string) *OpenAICompatibleClient {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(baseURL, "/")))
	}
	client := openai.NewClient(opts...)
	return &OpenAICompatibleClient{
		client:       &client,
		model:        model,
		systemPrompt: systemPrompt,
	}
}

func (c *OpenAICompatibleClient) Model() string {
	return c.model
}

func (c *OpenAICompatibleClient) Chat(messages []Message) (string, error) {
	body := openAIChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0,
	}
	resp, err := c.postChat(body)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty choices from openai-compatible chat")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAICompatibleClient) Detect(prompt string, knownMappings map[string]string, recentContext []Message, newMessage string) ([]MaskEntry, error) {
	instructions := buildTrustedInstructions(knownMappings, recentContext, newMessage)
	systemPrompt := c.systemPrompt
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = "You are a data masking engine. Return JSON array only."
	}
	body := openAIChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: instructions},
		},
		Temperature: 0,
		ResponseFmt: map[string]string{"type": "json_object"},
	}
	resp, err := c.postChat(body)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices from trusted llm detect")
	}
	return parseOpenAIDetectContent(resp.Choices[0].Message.Content)
}

func (c *OpenAICompatibleClient) postChat(body openAIChatRequest) (*openAIChatResponse, error) {
	ctx := context.Background()
	var out openAIChatResponse
	if err := c.client.Post(ctx, "/chat/completions", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseOpenAIDetectContent(content string) ([]MaskEntry, error) {
	var wrapped detectEnvelope
	if err := json.Unmarshal([]byte(content), &wrapped); err == nil && wrapped.Entries != nil {
		return wrapped.Entries, nil
	}
	return ParseMaskEntries(content)
}

func buildTrustedInstructions(knownMappings map[string]string, recentContext []Message, newMessage string) string {
	knownJSON, _ := json.Marshal(knownMappings)
	contextJSON, _ := json.Marshal(recentContext)
	return fmt.Sprintf("Return a JSON object with key 'entries'. Each entry must be {\"original\":\"...\",\"placeholder\":\"${TYPE_1}\",\"type\":\"TYPE\"}. Only detect NEW sensitive values from new_message. Do not repeat known mappings. known_mappings=%s recent_context=%s new_message=%q", string(knownJSON), string(contextJSON), newMessage)
}

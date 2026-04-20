package llm

import (
	"fmt"

	"llm-privacy-gaurd/internal/config"
)

type MetaAware interface {
	Model() string
}

func NewTrustedClient(cfg config.LLMConfig) (TrustedLLMClient, error) {
	switch cfg.Provider {
	case "openai", "openai-compatible":
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("trusted_llm base URL is required")
		}
		if cfg.Model == "" {
			return nil, fmt.Errorf("trusted_llm model is required")
		}
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("trusted_llm api key is required")
		}
		return NewOpenAICompatibleClient(cfg.BaseURL, cfg.APIKey, cfg.Model, cfg.SystemPrompt), nil
	case "":
		return nil, fmt.Errorf("trusted_llm provider is required")
	default:
		return nil, fmt.Errorf("unsupported trusted_llm provider: %s", cfg.Provider)
	}
}

func NewCloudClient(cfg config.LLMConfig) (CloudLLMClient, error) {
	switch cfg.Provider {
	case "openai", "openai-compatible":
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("cloud_llm base URL is required")
		}
		if cfg.Model == "" {
			return nil, fmt.Errorf("cloud_llm model is required")
		}
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("cloud_llm api key is required")
		}
		return NewOpenAICompatibleClient(cfg.BaseURL, cfg.APIKey, cfg.Model, ""), nil
	case "":
		return nil, fmt.Errorf("cloud_llm provider is required")
	default:
		return nil, fmt.Errorf("unsupported cloud_llm provider: %s", cfg.Provider)
	}
}

func ResolveModelName(client any, fallback string) string {
	if meta, ok := client.(MetaAware); ok {
		return meta.Model()
	}
	return fallback
}

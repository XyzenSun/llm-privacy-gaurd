package service

import (
	"llm-privacy-gaurd/internal/config"
	"llm-privacy-gaurd/internal/repository"
)

type SystemConfigService struct {
	repo *repository.SystemConfigRepository
}

func NewSystemConfigService(repo *repository.SystemConfigRepository) *SystemConfigService {
	return &SystemConfigService{repo: repo}
}

func (s *SystemConfigService) Resolve(cfg config.AppConfig) (config.AppConfig, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return cfg, err
	}

	values := map[string]string{}
	for _, item := range items {
		values[item.ConfigKey] = item.Value
	}

	resolved := cfg
	resolved.TrustedLLM.Provider = normalizeProvider(values["trusted_llm.provider"])
	resolved.TrustedLLM.Model = values["trusted_llm.model"]
	resolved.TrustedLLM.BaseURL = values["trusted_llm.base_url"]
	resolved.TrustedLLM.APIKey = values["trusted_llm.api_key"]
	resolved.TrustedLLM.SystemPrompt = values["trusted_llm.system_prompt"]
	resolved.CloudLLM.Provider = normalizeProvider(values["cloud_llm.provider"])
	resolved.CloudLLM.Model = values["cloud_llm.model"]
	resolved.CloudLLM.BaseURL = values["cloud_llm.base_url"]
	resolved.CloudLLM.APIKey = values["cloud_llm.api_key"]

	return resolved, nil
}

func (s *SystemConfigService) GetPublicConfig(cfg config.AppConfig) map[string]any {
	return map[string]any{
		"database": map[string]any{
			"type": cfg.Database.Type,
		},
		"trustedLlm": map[string]any{
			"provider":     cfg.TrustedLLM.Provider,
			"model":        cfg.TrustedLLM.Model,
			"baseUrl":      cfg.TrustedLLM.BaseURL,
			"apiKey":       cfg.TrustedLLM.APIKey,
			"apiKeySet":    cfg.TrustedLLM.APIKey != "",
			"systemPrompt": cfg.TrustedLLM.SystemPrompt,
		},
		"cloudLlm": map[string]any{
			"provider":  cfg.CloudLLM.Provider,
			"model":     cfg.CloudLLM.Model,
			"baseUrl":   cfg.CloudLLM.BaseURL,
			"apiKey":    cfg.CloudLLM.APIKey,
			"apiKeySet": cfg.CloudLLM.APIKey != "",
		},
	}
}

func (s *SystemConfigService) Update(values map[string]string) error {
	allowed := map[string]bool{
		"trusted_llm.provider":      true,
		"trusted_llm.model":         true,
		"trusted_llm.base_url":      true,
		"trusted_llm.api_key":       true,
		"trusted_llm.system_prompt": true,
		"cloud_llm.provider":        true,
		"cloud_llm.model":           true,
		"cloud_llm.base_url":        true,
		"cloud_llm.api_key":         true,
	}
	filtered := map[string]string{}
	for key, value := range values {
		if allowed[key] {
			filtered[key] = value
		}
	}
	return s.repo.SetMany(filtered)
}

func normalizeProvider(value string) string {
	if value == "mock" {
		return ""
	}
	return value
}

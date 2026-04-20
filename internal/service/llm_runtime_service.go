package service

import (
	"sync"

	"llm-privacy-gaurd/internal/config"
	"llm-privacy-gaurd/internal/llm"
)

type LLMRuntimeService struct {
	mu               sync.RWMutex
	cfg              *config.AppConfig
	trustedClient    llm.TrustedLLMClient
	trustedErr       error
	cloudClient      llm.CloudLLMClient
	cloudErr         error
	trustedModelName string
	cloudModelName   string
}

func NewLLMRuntimeService(cfg *config.AppConfig) *LLMRuntimeService {
	s := &LLMRuntimeService{cfg: cfg}
	s.Rebuild()
	return s
}

func (s *LLMRuntimeService) Rebuild() {
	s.mu.Lock()
	defer s.mu.Unlock()

	trustedClient, trustedErr := llm.NewTrustedClient(s.cfg.TrustedLLM)
	cloudClient, cloudErr := llm.NewCloudClient(s.cfg.CloudLLM)

	s.trustedClient = trustedClient
	s.trustedErr = trustedErr
	s.cloudClient = cloudClient
	s.cloudErr = cloudErr

	if trustedErr == nil && trustedClient != nil {
		s.trustedModelName = llm.ResolveModelName(trustedClient, s.cfg.TrustedLLM.Model)
	} else {
		s.trustedModelName = s.cfg.TrustedLLM.Model
	}
	if cloudErr == nil && cloudClient != nil {
		s.cloudModelName = llm.ResolveModelName(cloudClient, s.cfg.CloudLLM.Model)
	} else {
		s.cloudModelName = s.cfg.CloudLLM.Model
	}
}

func (s *LLMRuntimeService) TrustedClient() (llm.TrustedLLMClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.trustedErr != nil {
		return nil, s.trustedErr
	}
	return s.trustedClient, nil
}

func (s *LLMRuntimeService) CloudClient() (llm.CloudLLMClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cloudErr != nil {
		return nil, s.cloudErr
	}
	return s.cloudClient, nil
}

func (s *LLMRuntimeService) TrustedModelName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trustedModelName
}

func (s *LLMRuntimeService) CloudModelName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cloudModelName
}

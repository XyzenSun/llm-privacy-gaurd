package handler

import (
	"net/http"

	"llm-privacy-gaurd/internal/config"
	"llm-privacy-gaurd/internal/service"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	cfg     *config.AppConfig
	configS *service.SystemConfigService
	runtime *service.LLMRuntimeService
}

func NewConfigHandler(cfg *config.AppConfig, configS *service.SystemConfigService, runtime *service.LLMRuntimeService) *ConfigHandler {
	return &ConfigHandler{cfg: cfg, configS: configS, runtime: runtime}
}

func (h *ConfigHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, h.configS.GetPublicConfig(*h.cfg))
}

func (h *ConfigHandler) Update(c *gin.Context) {
	var req struct {
		TrustedLLM struct {
			Provider     string `json:"provider"`
			Model        string `json:"model"`
			BaseURL      string `json:"baseUrl"`
			APIKey       string `json:"apiKey"`
			SystemPrompt string `json:"systemPrompt"`
		} `json:"trustedLlm"`
		CloudLLM struct {
			Provider string `json:"provider"`
			Model    string `json:"model"`
			BaseURL  string `json:"baseUrl"`
			APIKey   string `json:"apiKey"`
		} `json:"cloudLlm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	values := map[string]string{
		"trusted_llm.provider":      req.TrustedLLM.Provider,
		"trusted_llm.model":         req.TrustedLLM.Model,
		"trusted_llm.base_url":      req.TrustedLLM.BaseURL,
		"trusted_llm.system_prompt": req.TrustedLLM.SystemPrompt,
		"cloud_llm.provider": req.CloudLLM.Provider,
		"cloud_llm.model":    req.CloudLLM.Model,
		"cloud_llm.base_url": req.CloudLLM.BaseURL,
	}
	if req.TrustedLLM.APIKey != "" {
		values["trusted_llm.api_key"] = req.TrustedLLM.APIKey
	}
	if req.CloudLLM.APIKey != "" {
		values["cloud_llm.api_key"] = req.CloudLLM.APIKey
	}

	if err := h.configS.Update(values); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resolved, err := h.configS.Resolve(*h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	*h.cfg = resolved
	h.runtime.Rebuild()
	c.JSON(http.StatusOK, h.configS.GetPublicConfig(*h.cfg))
}

func (h *ConfigHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

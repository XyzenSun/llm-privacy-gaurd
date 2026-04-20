package handler

import (
	"net/http"

	"llm-privacy-gaurd/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) Handle(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId" binding:"required"`
		Message   string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.chatSvc.Process(&service.ChatRequest{
		SessionID: req.SessionID,
		Message:   req.Message,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"llm-privacy-gaurd/internal/config"
	"llm-privacy-gaurd/internal/domain"
	"llm-privacy-gaurd/internal/llm"
	"llm-privacy-gaurd/internal/repository"
)

type ChatService struct {
	sessionRepo *repository.SessionRepository
	messageRepo *repository.MessageRepository
	mappingRepo *repository.MaskMappingRepository
	logRepo     *repository.ChatLogRepository
	maskingSvc  *MaskingService
	cloudLLM    llm.CloudLLMClient
	runtime     *LLMRuntimeService
	cfg         config.MaskingConfig
}

func NewChatService(
	sessionRepo *repository.SessionRepository,
	messageRepo *repository.MessageRepository,
	mappingRepo *repository.MaskMappingRepository,
	logRepo *repository.ChatLogRepository,
	maskingSvc *MaskingService,
	cloudLLM llm.CloudLLMClient,
	runtime *LLMRuntimeService,
	cfg config.MaskingConfig,
) *ChatService {
	return &ChatService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		mappingRepo: mappingRepo,
		logRepo:     logRepo,
		maskingSvc:  maskingSvc,
		cloudLLM:    cloudLLM,
		runtime:     runtime,
		cfg:         cfg,
	}
}

type ChatRequest struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

type ChatResponse struct {
	MessageID          string            `json:"messageId"`
	SessionID          string            `json:"sessionId"`
	Response           string            `json:"response"`
	UnmaskedResponse   string            `json:"unmaskedResponse"`
	PromptBeforeMask   string            `json:"promptBeforeMask"`
	PromptAfterMask    string            `json:"promptAfterMask"`
	CloudResponseMasked string           `json:"cloudResponseMasked"`
	ByPlaceholder      map[string]string `json:"byPlaceholder"`
	TurnID             int               `json:"turnId"`
}

func (s *ChatService) Process(req *ChatRequest) (*ChatResponse, error) {
	session, err := s.sessionRepo.FindByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	mappings, err := s.mappingRepo.ListBySession(session.ID)
	if err != nil {
		return nil, err
	}

	knownMappings := make(map[string]string)
	for _, m := range mappings {
		knownMappings[m.Original] = m.Placeholder
	}

	history, err := s.messageRepo.ListBySession(session.ID)
	if err != nil {
		return nil, err
	}

	turnID := len(history)/2 + 1

	maskResult, err := s.maskingSvc.Mask(session.ID, req.Message, knownMappings, turnID)
	if err != nil {
		return nil, err
	}

	cloudMessages := s.buildCloudMessages(history, maskResult.MaskedContent)
	cloudResp, err := s.cloudLLM.Chat(cloudMessages)
	if err != nil {
		return nil, err
	}

	byPlaceholder := make(map[string]string)
	for orig, placeholder := range maskResult.ByPlaceholder {
		byPlaceholder[placeholder] = orig
	}
	unmaskedResp := s.maskingSvc.Unmask(cloudResp, byPlaceholder)

	userMsg := &domain.Message{
		ID:            newID(),
		SessionID:     session.ID,
		Role:          "user",
		Content:       req.Message,
		MaskedContent: maskResult.MaskedContent,
		Hash:          hashContent(req.Message),
		IsMasked:      false,
	}
	if err := s.messageRepo.Create(userMsg); err != nil {
		return nil, err
	}

	assistantMsg := &domain.Message{
		ID:            newID(),
		SessionID:     session.ID,
		Role:          "assistant",
		Content:       unmaskedResp,
		MaskedContent: cloudResp,
		Hash:          hashContent(unmaskedResp),
		IsMasked:      false,
	}
	if err := s.messageRepo.Create(assistantMsg); err != nil {
		return nil, err
	}

	byPlaceholderJSON, _ := json.Marshal(maskResult.ByPlaceholder)
	chatLog := &domain.ChatLog{
		ID:                    newID(),
		SessionID:             session.ID,
		UserMessageID:         userMsg.ID,
		AssistantMessageID:    assistantMsg.ID,
		PromptBeforeMask:      req.Message,
		PromptAfterMask:       maskResult.MaskedContent,
		CloudResponseMasked:   cloudResp,
		CloudResponseUnmasked: unmaskedResp,
		ByPlaceholderJSON:     string(byPlaceholderJSON),
		TrustedLLMModel:       s.runtime.TrustedModelName(),
		CloudLLMModel:         s.runtime.CloudModelName(),
		TurnID:                turnID,
	}
	if err := s.logRepo.Create(chatLog); err != nil {
		return nil, err
	}

	return &ChatResponse{
		MessageID:           assistantMsg.ID,
		SessionID:           session.ID,
		Response:            unmaskedResp,
		UnmaskedResponse:    unmaskedResp,
		PromptBeforeMask:    req.Message,
		PromptAfterMask:     maskResult.MaskedContent,
		CloudResponseMasked: cloudResp,
		ByPlaceholder:       maskResult.ByPlaceholder,
		TurnID:              turnID,
	}, nil
}

func (s *ChatService) buildCloudMessages(history []domain.Message, newMaskedContent string) []llm.Message {
	messages := make([]llm.Message, 0, len(history)+1)
	for _, m := range history {
		if m.Role != "system" && m.Role != "user" && m.Role != "assistant" {
			continue
		}
		content := m.MaskedContent
		if content == "" {
			content = m.Content
		}
		messages = append(messages, llm.Message{
			Role:    m.Role,
			Content: content,
		})
	}
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: newMaskedContent,
	})
	return messages
}

func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func now() time.Time {
	return time.Now().UTC()
}
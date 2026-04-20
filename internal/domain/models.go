package domain

import "time"

type SystemAuth struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	TokenHash string    `gorm:"size:128;not null" json:"-"`
	TokenSalt string    `gorm:"size:64;not null" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Session struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Message struct {
	ID            string    `gorm:"primaryKey;size:36" json:"id"`
	SessionID     string    `gorm:"size:36;index;not null" json:"sessionId"`
	Role          string    `gorm:"size:20;index;not null" json:"role"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	MaskedContent string    `gorm:"type:text" json:"maskedContent"`
	Hash          string    `gorm:"size:64;index;not null" json:"hash"`
	IsMasked      bool      `gorm:"index;not null" json:"isMasked"`
	CreatedAt     time.Time `json:"createdAt"`
}

type MaskMapping struct {
	ID            string    `gorm:"primaryKey;size:36" json:"id"`
	SessionID     string    `gorm:"size:36;index:idx_session_original,unique;index:idx_session_placeholder,unique;not null" json:"sessionId"`
	Original      string    `gorm:"type:text;not null;index:idx_session_original,unique" json:"original"`
	Placeholder   string    `gorm:"size:128;not null;index:idx_session_placeholder,unique" json:"placeholder"`
	Type          string    `gorm:"size:64;not null" json:"type"`
	FirstSeenTurn int       `gorm:"not null" json:"firstSeenTurn"`
	LastSeenTurn  int       `gorm:"not null" json:"lastSeenTurn"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ChatLog struct {
	ID                    string    `gorm:"primaryKey;size:36" json:"id"`
	SessionID             string    `gorm:"size:36;index;not null" json:"sessionId"`
	UserMessageID         string    `gorm:"size:36;not null" json:"userMessageId"`
	AssistantMessageID    string    `gorm:"size:36;not null" json:"assistantMessageId"`
	PromptBeforeMask      string    `gorm:"type:text;not null" json:"promptBeforeMask"`
	PromptAfterMask       string    `gorm:"type:text;not null" json:"promptAfterMask"`
	CloudResponseMasked   string    `gorm:"type:text;not null" json:"cloudResponseMasked"`
	CloudResponseUnmasked string    `gorm:"type:text;not null" json:"cloudResponseUnmasked"`
	ByPlaceholderJSON     string    `gorm:"type:text;not null" json:"byPlaceholderJson"`
	TrustedLLMModel       string    `gorm:"size:128;not null" json:"trustedLlmModel"`
	CloudLLMModel         string    `gorm:"size:128;not null" json:"cloudLlmModel"`
	TurnID                int       `gorm:"index;not null" json:"turnId"`
	CreatedAt             time.Time `json:"createdAt"`
}

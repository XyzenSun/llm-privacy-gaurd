package repository

import (
	"errors"

	"llm-privacy-gaurd/internal/domain"

	"gorm.io/gorm"
)

type SystemAuthRepository struct {
	db *gorm.DB
}

type SessionRepository struct {
	db *gorm.DB
}

type MessageRepository struct {
	db *gorm.DB
}

type MaskMappingRepository struct {
	db *gorm.DB
}

type ChatLogRepository struct {
	db *gorm.DB
}

func NewSystemAuthRepository(db *gorm.DB) *SystemAuthRepository {
	return &SystemAuthRepository{db: db}
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func NewMaskMappingRepository(db *gorm.DB) *MaskMappingRepository {
	return &MaskMappingRepository{db: db}
}

func NewChatLogRepository(db *gorm.DB) *ChatLogRepository {
	return &ChatLogRepository{db: db}
}

func (r *SystemAuthRepository) First() (*domain.SystemAuth, error) {
	var auth domain.SystemAuth
	if err := r.db.First(&auth).Error; err != nil {
		return nil, err
	}
	return &auth, nil
}

func (r *SystemAuthRepository) Save(auth *domain.SystemAuth) error {
	return r.db.Save(auth).Error
}

func (r *SessionRepository) Create(session *domain.Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) List() ([]domain.Session, error) {
	var sessions []domain.Session
	err := r.db.Order("created_at desc").Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) FindByID(id string) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *MessageRepository) Create(message *domain.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) ListBySession(sessionID string) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *MaskMappingRepository) UpsertMany(mappings []domain.MaskMapping) error {
	for i := range mappings {
		mapping := mappings[i]
		var existing domain.MaskMapping
		err := r.db.Where("session_id = ? AND original = ?", mapping.SessionID, mapping.Original).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := r.db.Create(&mapping).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
		existing.Placeholder = mapping.Placeholder
		existing.Type = mapping.Type
		existing.LastSeenTurn = mapping.LastSeenTurn
		if err := r.db.Save(&existing).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *MaskMappingRepository) ListBySession(sessionID string) ([]domain.MaskMapping, error) {
	var mappings []domain.MaskMapping
	err := r.db.Where("session_id = ?", sessionID).Order("length(original) desc").Find(&mappings).Error
	return mappings, err
}

func (r *ChatLogRepository) Create(log *domain.ChatLog) error {
	return r.db.Create(log).Error
}

func (r *ChatLogRepository) List(sessionID string, limit int) ([]domain.ChatLog, error) {
	var logs []domain.ChatLog
	query := r.db.Order("created_at desc")
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

func (r *ChatLogRepository) FindByID(id string) (*domain.ChatLog, error) {
	var log domain.ChatLog
	if err := r.db.First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

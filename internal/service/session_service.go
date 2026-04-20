package service

import (
	"llm-privacy-gaurd/internal/domain"
	"llm-privacy-gaurd/internal/repository"
)

type SessionService struct {
	sessions *repository.SessionRepository
	messages *repository.MessageRepository
	mappings *repository.MaskMappingRepository
}

func NewSessionService(sessions *repository.SessionRepository, messages *repository.MessageRepository, mappings *repository.MaskMappingRepository) *SessionService {
	return &SessionService{sessions: sessions, messages: messages, mappings: mappings}
}

func (s *SessionService) Create(title string) (*domain.Session, error) {
	session := &domain.Session{
		ID:    newID(),
		Title: title,
	}
	if session.Title == "" {
		session.Title = "New Session"
	}
	if err := s.sessions.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionService) List() ([]domain.Session, error) {
	return s.sessions.List()
}

func (s *SessionService) Get(id string) (*domain.Session, error) {
	return s.sessions.FindByID(id)
}

func (s *SessionService) Messages(sessionID string) ([]domain.Message, error) {
	return s.messages.ListBySession(sessionID)
}

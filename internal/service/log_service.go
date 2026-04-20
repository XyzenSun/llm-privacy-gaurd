package service

import (
	"llm-privacy-gaurd/internal/domain"
	"llm-privacy-gaurd/internal/repository"
)

type LogService struct {
	repo *repository.ChatLogRepository
}

func NewLogService(repo *repository.ChatLogRepository) *LogService {
	return &LogService{repo: repo}
}

func (s *LogService) List(sessionID string, limit int) ([]domain.ChatLog, error) {
	return s.repo.List(sessionID, limit)
}

func (s *LogService) Get(id string) (*domain.ChatLog, error) {
	return s.repo.FindByID(id)
}

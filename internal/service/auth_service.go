package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"

	"llm-privacy-gaurd/internal/config"
	"llm-privacy-gaurd/internal/domain"
	"llm-privacy-gaurd/internal/repository"
)

type AuthService struct {
	repo *repository.SystemAuthRepository
	cfg  config.AuthConfig
}

func NewAuthService(repo *repository.SystemAuthRepository, cfg config.AuthConfig) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) EnsureInitialized() error {
	if s.cfg.InitialToken == "" {
		return errors.New("APP_AUTHTOKEN is required")
	}

	salt, err := randomHex(16)
	if err != nil {
		return err
	}
	hash := hashToken(s.cfg.InitialToken, salt)

	existing, err := s.repo.First()
	if err == nil {
		existing.TokenSalt = salt
		existing.TokenHash = hash
		return s.repo.Save(existing)
	}

	auth := &domain.SystemAuth{
		ID:        newID(),
		TokenSalt: salt,
		TokenHash: hash,
	}
	return s.repo.Save(auth)
}

func (s *AuthService) ValidateToken(token string) (bool, error) {
	auth, err := s.repo.First()
	if err != nil {
		return false, err
	}

	computed := hashToken(token, auth.TokenSalt)
	match := subtle.ConstantTimeCompare([]byte(computed), []byte(auth.TokenHash)) == 1
	return match, nil
}

func hashToken(token, salt string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", salt, token)))
	return hex.EncodeToString(sum[:])
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

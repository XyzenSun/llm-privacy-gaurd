package bootstrap

import (
	"fmt"

	"llm-privacy-gaurd/internal/config"
	"llm-privacy-gaurd/internal/domain"
	"llm-privacy-gaurd/internal/repository"
	"llm-privacy-gaurd/internal/service"

	"gorm.io/gorm"
)

type App struct {
	Config              config.AppConfig
	DB                  *gorm.DB
	AuthService         *service.AuthService
	SessionService      *service.SessionService
	LogService          *service.LogService
	SystemConfigService *service.SystemConfigService
	LLMRuntimeService   *service.LLMRuntimeService
}

func Initialize(cfg config.AppConfig) (*App, error) {
	db, err := config.OpenDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	authRepo := repository.NewSystemAuthRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	mappingRepo := repository.NewMaskMappingRepository(db)
	logRepo := repository.NewChatLogRepository(db)
	systemConfigRepo := repository.NewSystemConfigRepository(db)

	authService := service.NewAuthService(authRepo, cfg.Auth)
	if err := authService.EnsureInitialized(); err != nil {
		return nil, fmt.Errorf("apply auth token: %w", err)
	}

	systemConfigService := service.NewSystemConfigService(systemConfigRepo)
	resolvedCfg, err := systemConfigService.Resolve(cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve system config: %w", err)
	}

	app := &App{
		Config:              resolvedCfg,
		DB:                  db,
		AuthService:         authService,
		SessionService:      service.NewSessionService(sessionRepo, messageRepo, mappingRepo),
		LogService:          service.NewLogService(logRepo),
		SystemConfigService: systemConfigService,
	}
	app.LLMRuntimeService = service.NewLLMRuntimeService(&app.Config)

	return app, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.SystemAuth{},
		&domain.SystemConfig{},
		&domain.Session{},
		&domain.Message{},
		&domain.MaskMapping{},
		&domain.ChatLog{},
	)
}

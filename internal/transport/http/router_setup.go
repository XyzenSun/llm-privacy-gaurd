package http

import (
	"io"
	"os"
	"path/filepath"

	"llm-privacy-gaurd/internal/bootstrap"
	"llm-privacy-gaurd/internal/repository"
	"llm-privacy-gaurd/internal/service"
	"llm-privacy-gaurd/internal/transport/http/handler"
	"llm-privacy-gaurd/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func serveStatic(c *gin.Context, filepath string) {
	f, err := os.Open(filepath)
	if err != nil {
		c.String(404, "not found")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.String(500, "read error")
		return
	}
	contentType := "text/html; charset=utf-8"
	c.Data(200, contentType, data)
}

func SetupRouter(app *bootstrap.App) *gin.Engine {
	r := NewRouter()

	authMW := middleware.NewAuthMiddleware(
		app.AuthService,
		app.Config.Auth.SessionCookie,
		app.Config.Auth.SessionTTL,
		app.Config.Auth.SecureCookie,
	)

	mappingRepo := repository.NewMaskMappingRepository(app.DB)
	logRepo := repository.NewChatLogRepository(app.DB)
	messageRepo := repository.NewMessageRepository(app.DB)
	sessionRepo := repository.NewSessionRepository(app.DB)

	trustedLLM := service.NewLiveTrustedLLMClient(app.LLMRuntimeService)
	cloudLLM := service.NewLiveCloudLLMClient(app.LLMRuntimeService)

	maskingSvc := service.NewMaskingService(mappingRepo, trustedLLM)
	chatSvc := service.NewChatService(
		sessionRepo,
		messageRepo,
		mappingRepo,
		logRepo,
		maskingSvc,
		cloudLLM,
		app.LLMRuntimeService,
		app.Config.Masking,
	)

	sessionHandler := handler.NewSessionHandler(app.SessionService)
	logHandler := handler.NewLogHandler(app.LogService)
	configHandler := handler.NewConfigHandler(&app.Config, app.SystemConfigService, app.LLMRuntimeService)
	chatHandler := handler.NewChatHandler(chatSvc)

	r.GET("/api/v1/health", configHandler.Health)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", authMW.LoginHandler())
		auth.POST("/logout", authMW.LogoutHandler())
		auth.GET("/status", authMW.StatusHandler())
	}

	api := r.Group("/api/v1")
	api.Use(authMW.RequireAPIAuth())
	{
		sessions := api.Group("/sessions")
		{
			sessions.POST("", sessionHandler.Create)
			sessions.GET("", sessionHandler.List)
			sessions.GET("/:id", sessionHandler.Get)
			sessions.GET("/:id/messages", sessionHandler.Messages)
		}

		logs := api.Group("/logs")
		{
			logs.GET("", logHandler.List)
			logs.GET("/:id", logHandler.Get)
		}

		api.GET("/config", configHandler.Get)
		api.PUT("/config", configHandler.Update)
		api.POST("/chat", chatHandler.Handle)
	}

	webRoot := filepath.Join(app.Config.ProjectRoot, "web")
	assetsRoot := filepath.Join(webRoot, "assets")

	r.Static("/assets", assetsRoot)
	r.GET("/login.html", func(c *gin.Context) {
		serveStatic(c, filepath.Join(webRoot, "login.html"))
	})
	r.GET("/index.html", func(c *gin.Context) {
		serveStatic(c, filepath.Join(webRoot, "index.html"))
	})
	r.GET("/logs.html", func(c *gin.Context) {
		serveStatic(c, filepath.Join(webRoot, "logs.html"))
	})
	r.GET("/config.html", func(c *gin.Context) {
		serveStatic(c, filepath.Join(webRoot, "config.html"))
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/index.html")
	})

	return r
}
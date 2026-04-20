package middleware

import (
	"net/http"
	"time"

	"llm-privacy-gaurd/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	auth   *service.AuthService
	cookie string
	ttl    time.Duration
	secure bool
}

func NewAuthMiddleware(auth *service.AuthService, cookie string, ttl time.Duration, secure bool) *AuthMiddleware {
	return &AuthMiddleware{auth: auth, cookie: cookie, ttl: ttl, secure: secure}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie(m.cookie)
		if err != nil || session == "" {
			c.Redirect(http.StatusFound, "/login.html")
			c.Abort()
			return
		}
		valid, err := m.validateSession(session)
		if err != nil || !valid {
			c.Redirect(http.StatusFound, "/login.html")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAPIAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie(m.cookie)
		if err != nil || session == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		valid, err := m.validateSession(session)
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *AuthMiddleware) validateSession(session string) (bool, error) {
	return m.auth.ValidateToken(session)
}

func (m *AuthMiddleware) SetSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.ttl.Seconds()),
	})
}

func (m *AuthMiddleware) ClearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

type loginRequest struct {
	Authtoken string `json:"authtoken" binding:"required"`
}

func (m *AuthMiddleware) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		valid, err := m.auth.ValidateToken(req.Authtoken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		m.SetSessionCookie(c, req.Authtoken)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (m *AuthMiddleware) LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.ClearSessionCookie(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (m *AuthMiddleware) StatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie(m.cookie)
		if err != nil || session == "" {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
			return
		}
		valid, err := m.validateSession(session)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"authenticated": valid})
	}
}
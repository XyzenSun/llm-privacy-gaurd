package main

import (
	"log"

	"llm-privacy-gaurd/internal/bootstrap"
	"llm-privacy-gaurd/internal/config"
	httputil "llm-privacy-gaurd/internal/transport/http"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	cfg := config.Load()

	app, err := bootstrap.Initialize(cfg)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := httputil.SetupRouter(app)

	log.Printf("server listening on %s", cfg.Address())
	if err := r.Run(cfg.Address()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
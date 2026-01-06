package main

import (
	"github.com/gin-gonic/gin"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/database"
	"github.com/trenchesdeveloper/go-ai-store/internal/logger"
)

func main() {
	log := logger.NewLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	pool, err := database.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	defer pool.Close()

	log.Info().Msg("Database connection pool created")

	// Create store
	store := db.NewStore(pool)
	_ = store // TODO: pass store to server/handlers

	// Start the server
	gin.SetMode(cfg.Server.GinMode)

	log.Info().Msg("Starting server on port " + cfg.Server.Port)
}

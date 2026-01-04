package main

import (
	"github.com/gin-gonic/gin"
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

	db, err := database.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	defer db.Close()

	log.Info().Msg("Database connection pool created")

	// Start the server
	gin.SetMode(cfg.Server.GinMode)

	log.Info().Msg("Starting server on port " + cfg.Server.Port)




}

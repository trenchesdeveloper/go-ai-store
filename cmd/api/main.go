package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	_ "github.com/trenchesdeveloper/go-ai-store/docs" // Swagger docs
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/database"
	"github.com/trenchesdeveloper/go-ai-store/internal/logger"
	"github.com/trenchesdeveloper/go-ai-store/internal/server"
)

// @title           Go AI Store API
// @version         1.0
// @description     E-commerce API with products, cart, and orders management

// @contact.name   Opeyemi Samuel
// @contact.url  linkedin[https://linkedin.com/in/samuelopeyemi]

// @licence.name Apache 2.0
// @licence.url https://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer {token}
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

	log.Info().Msg("Database connection pool created")

	// Create store
	store := db.NewStore(pool)

	// Start the server
	gin.SetMode(cfg.Server.GinMode)

	srv, err := server.NewServer(cfg, log, store)
	if err != nil {
		pool.Close()
		log.Fatal().Err(err).Msg("failed to create server")
	}

	defer pool.Close()
	router := srv.SetupRoutes()

	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine so it doesn't block
	go func() {
		log.Info().Msg("Starting server on port " + cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
		return
	}

	log.Info().Msg("Server exited properly")
}

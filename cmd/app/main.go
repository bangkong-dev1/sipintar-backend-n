package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/azril773/go-backend/database"
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/azril773/go-backend/internal/repository"
	"github.com/azril773/go-backend/internal/routes"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://root:Tasikmalaya123..@localhost/centralpaud?sslmode=disable"
		logger.Warn("DATABASE_URL not set, using default", zap.String("dsn", dsn))
	}

	db, err := database.NewDB(dsn, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	migrationsDir := filepath.Join("database", "migrations")
	if err := db.RunMigrations(migrationsDir); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	tenantRepo := repository.NewPaudRepository(db.DB)
	tenantUsecase := usecases.NewPaudUseCase(db.DB, tenantRepo, logger)
	tenantHandler := http.NewPaudHandler(tenantUsecase, logger)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	routes.SetupTenantRoutes(r, tenantHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	logger.Info("Starting server", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

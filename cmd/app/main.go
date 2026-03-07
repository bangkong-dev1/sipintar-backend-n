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
		dsn = "postgres://postgres:postgres@localhost:5432/multitenant?sslmode=disable"
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

	tenantRepo := repository.NewTenantRepository(db.DB)
	tenantUsecase := usecases.NewTenantUsecase(db.DB, tenantRepo, logger)
	tenantHandler := http.NewTenantHandler(tenantUsecase, logger)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	routes.SetupTenantRoutes(r, tenantHandler)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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

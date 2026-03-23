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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	// Import docs
	_ "github.com/azril773/go-backend/docs"
)

// @title PAUD Central API
// @version 1.0
// @description API untuk manajemen PAUD Multi-Tenant
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@paud.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

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

	paudRepo := &repository.PaudRepository{DB: db.DB}
	paudUsecase := &usecases.PaudUseCase{DB: db.DB, Repo: paudRepo, Logger: logger.Named("paud")}
	paudHandler := &http.PaudHandler{Usecase: paudUsecase, Logger: logger.Named("handler")}

	roleRepo := &repository.RoleRepository{DB: db.DB}
	roleUsecase := &usecases.RoleUseCase{DB: db.DB, Repo: roleRepo, Logger: logger.Named("role")}
	roleHandler := &http.RoleHandler{Usecase: roleUsecase, Logger: logger.Named("handler")}

	userRepo := &repository.UserRepository{DB: db.DB}
	userUsecase := &usecases.UserUseCase{DB: db.DB, Repo: userRepo, PaudRepo: paudRepo, Logger: logger.Named("user")}
	userHandler := &http.UserHandler{Usecase: userUsecase, Logger: logger.Named("user_handler")}

	authRepo := &repository.AuthRepository{DB: db.DB}
	authUsecase := usecases.NewAuthUseCase(userRepo, paudRepo, authRepo, logger.Named("auth"))
	authHandler := &http.AuthHandler{Usecase: authUsecase, Logger: logger.Named("auth_handler")}
	authMiddleware := &http.AuthMiddleware{Usecase: authUsecase, Logger: logger.Named("auth_middleware")}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.AuthRoute(r, authHandler)
	routes.PaudRoute(r, paudHandler, authMiddleware)
	routes.RoleRoute(r, roleHandler, paudHandler, authMiddleware)
	routes.UserRoute(r, userHandler, paudHandler, authMiddleware)

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

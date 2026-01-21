package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/user/go-stock-viewer-back/src/stockviewer/config"
	"github.com/user/go-stock-viewer-back/src/stockviewer/httpapi"
	"github.com/user/go-stock-viewer-back/src/stockviewer/integrations/karenai"
	"github.com/user/go-stock-viewer-back/src/stockviewer/recommendation"
	"github.com/user/go-stock-viewer-back/src/stockviewer/stocks"

	_ "github.com/user/go-stock-viewer-back/docs"
)

// @title           Stock Viewer API
// @version         1.0
// @description     API for viewing and analyzing stock recommendations
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@stockviewer.local

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.basic  BasicAuth

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	stocksStorage, err := stocks.NewStorage(db)
	if err != nil {
		log.Fatalf("Failed to initialize stocks storage: %v", err)
	}

	karenaiClient := karenai.NewClient(
		cfg.External.KarenAIBaseURL,
		cfg.External.KarenAIToken,
	)

	stocksService := stocks.NewService(stocksStorage, karenaiClient)
	recommendationService := recommendation.NewService(stocksStorage)

	api := httpapi.New(httpapi.Config{
		StocksService:         stocksService,
		RecommendationService: recommendationService,
		BasicAuthUser:         cfg.Auth.Username,
		BasicAuthPassword:     cfg.Auth.Password,
	})

	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	api.ConfigureRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			sqlDB, err := db.DB()
			if err == nil && sqlDB.Ping() == nil {
				log.Println("Database connection established")
				return db, nil
			}
		}
		log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		time.Sleep(3 * time.Second)
	}

	return nil, err
}

package main

import (
	_ "awesomeProject/docs"
	"awesomeProject/handlers"
	"awesomeProject/logger"
	"awesomeProject/models"
	"awesomeProject/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

// @title Music Library API
// @version 1.0
// @description API for managing music library
// @host localhost:8081
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	logger.Init()

	required := []string{"DATABASE_URL", "EXTERNAL_API_URL", "PORT"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Fatalf("Environment variable %s is required", key)
		}
	}

	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		logger.Info("Failed to connect to database", zap.Error(err))
		log.Fatal("Failed to connect to database")
	}

	db.AutoMigrate(&models.Song{})

	musicAPI := services.NewMusicAPIService(os.Getenv("EXTERNAL_API_URL"))
	songHandler := handlers.NewSongHandler(db, musicAPI)

	r := gin.Default()

	r.GET("/api/v1/song", songHandler.List)
	r.GET("/api/v1//song/:id/text", songHandler.GetText)
	r.POST("/api/v1/song", songHandler.Create)
	r.PUT("/api/v1/song/:id", songHandler.Update)
	r.DELETE("/api/v1/song/:id", songHandler.Delete)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	logger.Info("Starting server", zap.String("port", port))
	r.Run(":" + port)
}

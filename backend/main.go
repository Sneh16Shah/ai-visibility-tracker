package main

import (
	"log"
	"os"

	"github.com/Sneh16Shah/ai-visibility-tracker/config"
	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/routes"
	"github.com/Sneh16Shah/ai-visibility-tracker/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file - try current dir first, then parent (for running from backend/)
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("📝 No .env file found, using environment variables")
		} else {
			log.Println("✅ Loaded .env from parent directory")
		}
	} else {
		log.Println("✅ Loaded .env file")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	if err := db.Connect(cfg); err != nil {
		log.Printf("⚠️ Database connection failed: %v", err)
		log.Println("📝 Running in demo mode without database")
	} else {
		defer db.Close()

		// Seed default user if none exists
		userRepo := db.NewUserRepository()
		if err := userRepo.CreateDefaultUser(); err != nil {
			log.Printf("⚠️ Failed to create default user: %v", err)
		} else {
			log.Println("👤 Default user ready (demo@example.com / demo123)")
		}
	}

	// Initialize AI analysis service with rate limiting
	services.InitAnalysisService(cfg)

	// Initialize Compare Models service (OpenRouter multi-model comparison)
	services.InitCompareService(cfg)

	// Initialize router
	router := gin.Default()

	// Enable CORS for frontend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup routes
	routes.Setup(router)

	// Get port from config or default
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 AI Visibility Tracker API starting on port %s", port)
	log.Printf("📊 Environment: %s", cfg.Environment)
	log.Printf("🗄️ Database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

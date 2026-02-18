package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"pawked.com/sendyourpicks/cmd/admin_bootstrap"
	"pawked.com/sendyourpicks/internal/api"
	"pawked.com/sendyourpicks/internal/api/middleware"
	"pawked.com/sendyourpicks/internal/db"
	"pawked.com/sendyourpicks/internal/logger"
)

func main() {

	// Load .env from repo root if present; in containers env vars come from the runtime.
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize logger and read log level from env
	logger.Init()

	// load auth middleware config
	authConfig, err := middleware.LoadAuthConfig()
	if err != nil {
		logger.Error("Failed to load auth middleware configuration", "error", err)
		log.Fatalf("auth config error: %v", err)
	}

	// gin mode: debug, release, test
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
		logger.Info("Gin mode set", "mode", ginMode)
	}

	// site address
	siteAddress := os.Getenv("SITE_ADDRESS")
	if siteAddress == "" {
		logger.Error("Please set SITE_ADDRESS in .env")
		os.Exit(1)
	}
	logger.Info("Site Address loaded", "address", siteAddress)

	// connect to database
	dbx, err := db.NewPostgres()
	if err != nil {
		logger.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	// Register database connection pool metrics
	middleware.RegisterDBMetrics(dbx)

	// seed teams
	logger.Info("checking team table data")
	if err := admin_bootstrap.SeedTeams(dbx); err != nil {
		logger.Error("Team seeding failed", "error", err)
		os.Exit(1)
	}
	logger.Info("team seed check complete")

	// seed default global settings table
	logger.Info("checking default settings table")
	if err := admin_bootstrap.SeedGlobalSettings(dbx); err != nil {
		logger.Error("Global setting seeding failed", "error", err)
		os.Exit(1)
	}

	// create the Gin router with default logger and recovery
	router := gin.Default()

	// Prometheus metrics endpoint (no auth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Add CORS middleware globally so SvelteKit can call the API
	router.Use(corsMiddleware(siteAddress))

	// Prometheus HTTP metrics middleware
	router.Use(middleware.MetricsMiddleware())

	// Register routes
	api.RegisterRoutes(router, dbx, authConfig)

	//get server address
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		logger.Info("HTTP_ADDR not set in .env Defaulting to port 8080")
		addr = ":8080"
	}

	// start serving
	logger.Info("starting HTTP server", "port", addr)
	if err := router.Run(addr); err != nil {
		logger.Error("server initialization failed", "error", err)
		os.Exit(1)
	}
}

// CORS middleware to allow SvelteKit to call the API
func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"nmapshot/config"
	_ "nmapshot/docs" // This line is crucial for swagger to find the docs
	"nmapshot/handler"
	"nmapshot/middleware"
)

// @title Nmap Shot API
// @version 1.0
// @description REST API service to execute Nmap scans and return XML results.

// @license.name MIT
// @license.url https://github.com/gpy/nmapshot/blob/main/LICENSE

// @host localhost:8082
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables")
	}

	// Load allowed ports configuration
	allowedPorts, err := config.LoadAllowedPorts()
	if err != nil {
		log.Fatalf("Failed to load allowed ports config: %v", err)
	}
	log.Printf("Allowed ports: %s", allowedPorts)

	r := gin.Default()

	// Swagger configuration
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/swagger/index.html")
	})
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Routes (protected by API key)
	scanHandler := handler.NewScanHandler(allowedPorts)
	v1 := r.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth())
	{
		v1.POST("/scan", scanHandler.Handle)
	}

	log.Println("Server starting on :8082...")
	if err := r.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}

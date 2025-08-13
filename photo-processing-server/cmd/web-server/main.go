package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"photo-processing-server/internal/services"
	"photo-processing-server/internal/web"
	"photo-processing-server/internal/config"
)

func main() {
	log.Println("Starting EnDeCode Web Server...")

	// Load config
	cfg := config.Load()

	// Initialize services
	logger := services.NewLogger()
	logger.SetLevel(cfg.LogLevel)
	processor := services.NewProcessor(logger)
	
	// Initialize WebSocket hub
	web.InitializeWebSocket(logger)
	
	// Setup Gin router
	router := gin.Default()
	
	// Configure CORS (from env)
	corsCfg := cors.DefaultConfig()
	origins := strings.Split(cfg.AllowedOrigins, ",")
	corsCfg.AllowOrigins = origins
	corsCfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsCfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsCfg.AllowCredentials = true
	router.Use(cors.New(corsCfg))

	// API auth middleware (optional if APIToken provided)
	authMiddleware := func(c *gin.Context) {
		if cfg.APIToken == "" {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != cfg.APIToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
	
	// Create web handler
	webHandler := web.NewWebHandler(processor, logger)
	
	// Setup routes
	webHandler.SetupRoutes(router)
	web.SetupWebSocketRoutes(router)

	// Protect API and WS
	router.Use(func(c *gin.Context) {
		// Only protect /api and /ws when APIToken present
		if cfg.APIToken != "" && (strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/ws")) {
			authMiddleware(c)
			return
		}
		c.Next()
	})
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "EnDeCode Web Server",
			"version": "2.1.1",
		})
	})
	
	// System info endpoint (matches showInfo from desktop)
	router.GET("/api/info", func(c *gin.Context) {
		info := map[string]interface{}{
			"name":     "EnDeCode Web Server",
			"version":  "2.1.1 - Web",
			"platform": "Web Browser",
			"language": "Go + React + TypeScript",
			"updated":  "January 7, 2025",
			"author":   "vsdev. | Vladislav Slugin",
			"contact":  "vslugin@vsdev.top",
			"features": []string{
				"File encryption/decryption",
				"Batch copying with numbering",
				"Visible/invisible watermarks",
				"Smart file swapping",
				"Drag and drop support",
			},
			"fileSupport": []string{
				"Images (.jpg, .jpeg, .png)",
				"Videos (.mp4, .avi, .mov, .mkv)",
				"Text (.txt)",
			},
			"techStack": []string{
				"Go Backend + WebSocket",
				"React + TypeScript",
				"Tailwind CSS",
			},
		}
		c.JSON(http.StatusOK, info)
	})
	
	// Log startup message
	logger.Log(strings.Repeat("=", 70))
	logger.Log(`
███████╗███╗   ██╗██████╗ ███████╗ ██████╗ ██████╗ ██████╗ ███████╗
██╔════╝████╗  ██║██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝
█████╗  ██╔██╗ ██║██║  ██║█████╗  ██║     ██║   ██║██║  ██║█████╗  
██╔══╝  ██║╚██╗██║██║  ██║██╔══╝  ██║     ██║   ██║██║  ██║██╔══╝  
███████╗██║ ╚████║██████╔╝███████╗╚██████╗╚██████╔╝██████╔╝███████╗
╚══════╝╚═╝  ╚═══╝╚═════╝ ╚══════╝ ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝`)
	logger.Log(strings.Repeat("=", 70))
	logger.Log("")
	logger.Log("                    EnDeCode by vsdev.")
	logger.Log("                      [v2.1.1 - Web]")
	logger.Log("")
	logger.Log("Platform        Web Browser")
	logger.Log("Language        Go + React + TypeScript")
	logger.Log("Updated         January 7, 2025")
	logger.Log("Author          vsdev. | Vladislav Slugin")
	logger.Log("Contact         vslugin@vsdev.top")
	logger.Log("")
	logger.Log("Server starting on :" + cfg.Port)
	logger.Log("Web UI available at http://localhost:" + cfg.Port)
	logger.Log("API endpoints at http://localhost:" + cfg.Port + "/api")
	logger.Log("WebSocket at ws://localhost:" + cfg.Port + "/ws")
	logger.Log("")
	logger.Log(strings.Repeat("=", 70))

	// Start server
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
package main

import (
	"fmt"
	"html/template"
	"log"
	"patbin/config"
	"patbin/database"
	"patbin/handlers"
	"patbin/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Custom template functions
	r.SetFuncMap(template.FuncMap{
		"timeAgo": func(t time.Time) string {
			now := time.Now()
			diff := now.Sub(t)

			if diff < time.Minute {
				return "just now"
			} else if diff < time.Hour {
				mins := int(diff.Minutes())
				return fmt.Sprintf("%d minutes ago", mins)
			} else if diff < 24*time.Hour {
				hours := int(diff.Hours())
				return fmt.Sprintf("%d hours ago", hours)
			} else if diff < 7*24*time.Hour {
				days := int(diff.Hours() / 24)
				return fmt.Sprintf("%d days ago", days)
			}
			return t.Format("Jan 2, 2006")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 at 3:04 PM")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := 0; i < count; i++ {
				result[i] = i
			}
			return result
		},
	})

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Apply auth middleware globally
	r.Use(middleware.AuthMiddleware(cfg))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	pasteHandler := handlers.NewPasteHandler()
	userHandler := handlers.NewUserHandler()

	// Public routes
	r.GET("/", pasteHandler.HomePage)
	r.GET("/login", authHandler.LoginPage)
	r.GET("/register", authHandler.RegisterPage)

	// Auth API routes
	api := r.Group("/api")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)
		api.GET("/auth/me", authHandler.GetCurrentUser)

		// Paste API routes
		api.POST("/paste", pasteHandler.CreatePaste)
		api.GET("/paste/:id", pasteHandler.GetPaste)
		api.PUT("/paste/:id", middleware.RequireAuth(), pasteHandler.UpdatePaste)
		api.DELETE("/paste/:id", middleware.RequireAuth(), pasteHandler.DeletePaste)
		api.POST("/paste/:id/fork", pasteHandler.ForkPaste)
		api.GET("/pastes/recent", pasteHandler.RecentPastes)

		// User API routes
		api.GET("/user/:username", userHandler.GetUserProfile)
		api.GET("/dashboard", middleware.RequireAuth(), userHandler.GetDashboard)
	}

	// Web routes
	r.GET("/dashboard", middleware.RequireAuth(), userHandler.GetDashboardPage)
	r.GET("/u/:username", userHandler.GetUserProfilePage)
	r.GET("/:id/edit", middleware.RequireAuth(), pasteHandler.EditPastePage)
	r.GET("/:id/raw", pasteHandler.GetRawPaste)
	r.GET("/:id", pasteHandler.ViewPastePage)

	// Start server
	log.Printf("🚀 Patbin server running on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

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
	cfg := config.Load()
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatal("Database init failed:", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"timeAgo": func(t time.Time) string {
			diff := time.Since(t)
			switch {
			case diff < time.Minute:
				return "just now"
			case diff < time.Hour:
				return fmt.Sprintf("%dm ago", int(diff.Minutes()))
			case diff < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(diff.Hours()))
			case diff < 7*24*time.Hour:
				return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
			default:
				return t.Format("Jan 2, 2006")
			}
		},
		"formatTime": func(t time.Time) string { return t.Format("Jan 2, 2006 at 3:04 PM") },
		"add":        func(a, b int) int { return a + b },
		"iterate": func(n int) []int {
			r := make([]int, n)
			for i := range r {
				r[i] = i
			}
			return r
		},
	})

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	r.Use(middleware.AuthMiddleware(cfg))

	authHandler := handlers.NewAuthHandler(cfg)
	pasteHandler := handlers.NewPasteHandler()
	userHandler := handlers.NewUserHandler()

	r.GET("/", pasteHandler.HomePage)
	r.GET("/login", authHandler.LoginPage)
	r.GET("/register", authHandler.RegisterPage)

	api := r.Group("/api")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)
		api.GET("/auth/me", authHandler.GetCurrentUser)
		api.POST("/paste", pasteHandler.CreatePaste)
		api.GET("/paste/:id", pasteHandler.GetPaste)
		api.PUT("/paste/:id", middleware.RequireAuth(), pasteHandler.UpdatePaste)
		api.DELETE("/paste/:id", middleware.RequireAuth(), pasteHandler.DeletePaste)
		api.POST("/paste/:id/fork", pasteHandler.ForkPaste)
		api.GET("/pastes/recent", pasteHandler.RecentPastes)
		api.GET("/user/:username", userHandler.GetUserProfile)
		api.GET("/dashboard", middleware.RequireAuth(), userHandler.GetDashboard)
	}

	r.GET("/dashboard", middleware.RequireAuth(), userHandler.GetDashboardPage)
	r.GET("/u/:username", userHandler.GetUserProfilePage)
	r.GET("/:id/edit", middleware.RequireAuth(), pasteHandler.EditPastePage)
	r.GET("/:id/raw", pasteHandler.GetRawPaste)
	r.GET("/:id", pasteHandler.ViewPastePage)

	log.Printf("🚀 Patbin running on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Server failed:", err)
	}
}

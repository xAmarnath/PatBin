package handlers

import (
	"net/http"
	"patbin/database"
	"patbin/middleware"
	"patbin/models"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetUserProfile returns a user's public pastes
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	if result := database.DB.Where("username = ?", username).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var pastes []models.Paste
	database.DB.Where("user_id = ? AND is_public = ?", user.ID, true).
		Order("created_at DESC").
		Find(&pastes)

	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"pastes": pastes,
	})
}

// GetUserProfilePage renders the user profile page
func (h *UserHandler) GetUserProfilePage(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	if result := database.DB.Where("username = ?", username).First(&user); result.Error != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Not Found - Patbin",
			"message": "User not found",
		})
		return
	}

	var pastes []models.Paste
	database.DB.Where("user_id = ? AND is_public = ?", user.ID, true).
		Order("created_at DESC").
		Find(&pastes)

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title":       user.Username + " - Patbin",
		"profileUser": user,
		"pastes":      pastes,
	})
}

// GetDashboard returns the current user's dashboard with all their pastes
func (h *UserHandler) GetDashboard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var pastes []models.Paste
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&pastes)

	c.JSON(http.StatusOK, gin.H{"pastes": pastes})
}

// GetDashboardPage renders the user's dashboard page
func (h *UserHandler) GetDashboardPage(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	username, _ := middleware.GetUsername(c)

	var pastes []models.Paste
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&pastes)

	// Count stats
	var publicCount, privateCount int64
	database.DB.Model(&models.Paste{}).Where("user_id = ? AND is_public = ?", userID, true).Count(&publicCount)
	database.DB.Model(&models.Paste{}).Where("user_id = ? AND is_public = ?", userID, false).Count(&privateCount)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":        "Dashboard - Patbin",
		"username":     username,
		"pastes":       pastes,
		"publicCount":  publicCount,
		"privateCount": privateCount,
		"totalCount":   len(pastes),
	})
}

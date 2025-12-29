package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"patbin/database"
	"patbin/middleware"
	"patbin/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PasteHandler struct{}

func NewPasteHandler() *PasteHandler {
	return &PasteHandler{}
}

type CreatePasteRequest struct {
	Title         string `json:"title"`
	Content       string `json:"content" binding:"required"`
	Language      string `json:"language"`
	IsPublic      bool   `json:"is_public"`
	ExpiresIn     string `json:"expires_in"` // "1h", "1d", "1w", "never"
	BurnAfterRead bool   `json:"burn_after_read"`
}

type UpdatePasteRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Language string `json:"language"`
	IsPublic *bool  `json:"is_public"`
}

// generateID creates a random 8-character hex ID
func generateID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreatePaste creates a new paste
func (h *PasteHandler) CreatePaste(c *gin.Context) {
	var req CreatePasteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	// Generate unique ID
	var id string
	for {
		id = generateID()
		var existing models.Paste
		if result := database.DB.First(&existing, "id = ?", id); result.Error != nil {
			break
		}
	}

	paste := models.Paste{
		ID:            id,
		Title:         req.Title,
		Content:       req.Content,
		Language:      req.Language,
		IsPublic:      req.IsPublic,
		BurnAfterRead: req.BurnAfterRead,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Set user if authenticated
	if userID, ok := middleware.GetUserID(c); ok {
		paste.UserID = &userID
	}

	// Set expiration
	if req.ExpiresIn != "" && req.ExpiresIn != "never" {
		var duration time.Duration
		switch req.ExpiresIn {
		case "1h":
			duration = time.Hour
		case "1d":
			duration = 24 * time.Hour
		case "1w":
			duration = 7 * 24 * time.Hour
		case "1m":
			duration = 30 * 24 * time.Hour
		}
		if duration > 0 {
			expiresAt := time.Now().Add(duration)
			paste.ExpiresAt = &expiresAt
		}
	}

	if result := database.DB.Create(&paste); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create paste"})
		return
	}

	c.JSON(http.StatusCreated, paste)
}

// GetPaste retrieves a paste by ID
func (h *PasteHandler) GetPaste(c *gin.Context) {
	id := c.Param("id")

	// Remove extension if present
	if idx := strings.LastIndex(id, "."); idx != -1 {
		id = id[:idx]
	}

	var paste models.Paste
	if result := database.DB.Preload("User").First(&paste, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste not found"})
		return
	}

	// Check if expired
	if paste.ExpiresAt != nil && paste.ExpiresAt.Before(time.Now()) {
		database.DB.Delete(&paste)
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste has expired"})
		return
	}

	// Check visibility
	if !paste.IsPublic {
		userID, authenticated := middleware.GetUserID(c)
		if !authenticated || paste.UserID == nil || *paste.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "This paste is private"})
			return
		}
	}

	// Handle burn after read
	if paste.BurnAfterRead && paste.Views > 0 {
		database.DB.Delete(&paste)
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste has been burned after reading"})
		return
	}

	// Increment views
	database.DB.Model(&paste).Update("views", paste.Views+1)
	paste.Views++

	c.JSON(http.StatusOK, paste)
}

// UpdatePaste updates an existing paste
func (h *PasteHandler) UpdatePaste(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var paste models.Paste
	if result := database.DB.First(&paste, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste not found"})
		return
	}

	// Check ownership
	if paste.UserID == nil || *paste.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own pastes"})
		return
	}

	var req UpdatePasteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if result := database.DB.Model(&paste).Updates(updates); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update paste"})
		return
	}

	database.DB.First(&paste, "id = ?", id)
	c.JSON(http.StatusOK, paste)
}

// DeletePaste deletes a paste
func (h *PasteHandler) DeletePaste(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var paste models.Paste
	if result := database.DB.First(&paste, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste not found"})
		return
	}

	// Check ownership
	if paste.UserID == nil || *paste.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own pastes"})
		return
	}

	if result := database.DB.Delete(&paste); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete paste"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Paste deleted successfully"})
}

// GetRawPaste returns the raw content of a paste
func (h *PasteHandler) GetRawPaste(c *gin.Context) {
	id := c.Param("id")

	var paste models.Paste
	if result := database.DB.First(&paste, "id = ?", id); result.Error != nil {
		c.String(http.StatusNotFound, "Paste not found")
		return
	}

	// Check if expired
	if paste.ExpiresAt != nil && paste.ExpiresAt.Before(time.Now()) {
		database.DB.Delete(&paste)
		c.String(http.StatusNotFound, "Paste has expired")
		return
	}

	// Check visibility
	if !paste.IsPublic {
		userID, authenticated := middleware.GetUserID(c)
		if !authenticated || paste.UserID == nil || *paste.UserID != userID {
			c.String(http.StatusForbidden, "This paste is private")
			return
		}
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, paste.Content)
}

// ForkPaste creates a copy of an existing paste
func (h *PasteHandler) ForkPaste(c *gin.Context) {
	id := c.Param("id")

	var original models.Paste
	if result := database.DB.First(&original, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste not found"})
		return
	}

	// Check visibility for forking
	if !original.IsPublic {
		userID, authenticated := middleware.GetUserID(c)
		if !authenticated || original.UserID == nil || *original.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot fork a private paste"})
			return
		}
	}

	// Generate new ID
	var newID string
	for {
		newID = generateID()
		var existing models.Paste
		if result := database.DB.First(&existing, "id = ?", newID); result.Error != nil {
			break
		}
	}

	forked := models.Paste{
		ID:        newID,
		Title:     original.Title + " (Fork)",
		Content:   original.Content,
		Language:  original.Language,
		IsPublic:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set user if authenticated
	if userID, ok := middleware.GetUserID(c); ok {
		forked.UserID = &userID
	}

	if result := database.DB.Create(&forked); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fork paste"})
		return
	}

	c.JSON(http.StatusCreated, forked)
}

// ViewPastePage renders the paste view page
func (h *PasteHandler) ViewPastePage(c *gin.Context) {
	id := c.Param("id")
	ext := ""

	// Extract extension for syntax highlighting
	if idx := strings.LastIndex(id, "."); idx != -1 {
		ext = id[idx+1:]
		id = id[:idx]
	}

	var paste models.Paste
	if result := database.DB.Preload("User").First(&paste, "id = ?", id); result.Error != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Not Found - Patbin",
			"message": "Paste not found",
		})
		return
	}

	// Check if expired
	if paste.ExpiresAt != nil && paste.ExpiresAt.Before(time.Now()) {
		database.DB.Delete(&paste)
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Expired - Patbin",
			"message": "This paste has expired",
		})
		return
	}

	// Check visibility
	if !paste.IsPublic {
		userID, authenticated := middleware.GetUserID(c)
		if !authenticated || paste.UserID == nil || *paste.UserID != userID {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"title":   "Private - Patbin",
				"message": "This paste is private",
			})
			return
		}
	}

	// Handle burn after read
	if paste.BurnAfterRead && paste.Views > 0 {
		database.DB.Delete(&paste)
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Burned - Patbin",
			"message": "This paste has been burned after reading",
		})
		return
	}

	// Increment views
	database.DB.Model(&paste).Update("views", paste.Views+1)
	paste.Views++

	// Determine language
	language := paste.Language
	if ext != "" {
		language = models.GetLanguageFromExtension(ext)
	}
	if language == "" {
		language = "plaintext"
	}

	// Count lines
	lines := strings.Count(paste.Content, "\n") + 1

	// Check if current user owns this paste
	isOwner := false
	if userID, ok := middleware.GetUserID(c); ok && paste.UserID != nil {
		isOwner = *paste.UserID == userID
	}

	c.HTML(http.StatusOK, "view.html", gin.H{
		"title":    paste.Title + " - Patbin",
		"paste":    paste,
		"language": language,
		"lines":    lines,
		"isOwner":  isOwner,
		"ext":      ext,
	})
}

// HomePage renders the home page with paste creation form
func (h *PasteHandler) HomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Patbin - Modern Pastebin",
	})
}

// EditPastePage renders the paste edit page
func (h *PasteHandler) EditPastePage(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var paste models.Paste
	if result := database.DB.First(&paste, "id = ?", id); result.Error != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Not Found - Patbin",
			"message": "Paste not found",
		})
		return
	}

	// Check ownership
	if paste.UserID == nil || *paste.UserID != userID {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title":   "Forbidden - Patbin",
			"message": "You can only edit your own pastes",
		})
		return
	}

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"title": "Edit - " + paste.Title,
		"paste": paste,
	})
}

// RecentPastes returns recent public pastes
func (h *PasteHandler) RecentPastes(c *gin.Context) {
	var pastes []models.Paste
	database.DB.Where("is_public = ?", true).
		Order("created_at DESC").
		Limit(20).
		Preload("User").
		Find(&pastes)

	c.JSON(http.StatusOK, pastes)
}

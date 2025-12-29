package models

import (
	"time"
)

type Paste struct {
	ID            string     `gorm:"primaryKey;size:12" json:"id"`
	Title         string     `gorm:"size:255" json:"title"`
	Content       string     `gorm:"type:text;not null" json:"content"`
	Language      string     `gorm:"size:50" json:"language"`
	IsPublic      bool       `gorm:"default:true" json:"is_public"`
	Views         int        `gorm:"default:0" json:"views"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	BurnAfterRead bool       `gorm:"default:false" json:"burn_after_read"`
	UserID        *uint      `gorm:"index" json:"user_id,omitempty"`
	User          *User      `gorm:"constraint:OnDelete:SET NULL" json:"user,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Language extension mappings
var LanguageExtensions = map[string]string{
	"go":         "go",
	"py":         "python",
	"python":     "python",
	"js":         "javascript",
	"javascript": "javascript",
	"ts":         "typescript",
	"typescript": "typescript",
	"html":       "html",
	"css":        "css",
	"json":       "json",
	"xml":        "xml",
	"yaml":       "yaml",
	"yml":        "yaml",
	"md":         "markdown",
	"markdown":   "markdown",
	"sql":        "sql",
	"sh":         "bash",
	"bash":       "bash",
	"c":          "c",
	"cpp":        "cpp",
	"h":          "c",
	"hpp":        "cpp",
	"java":       "java",
	"rs":         "rust",
	"rust":       "rust",
	"rb":         "ruby",
	"ruby":       "ruby",
	"php":        "php",
	"swift":      "swift",
	"kt":         "kotlin",
	"kotlin":     "kotlin",
	"scala":      "scala",
	"r":          "r",
	"lua":        "lua",
	"perl":       "perl",
	"pl":         "perl",
	"txt":        "plaintext",
	"text":       "plaintext",
}

func GetLanguageFromExtension(ext string) string {
	if lang, ok := LanguageExtensions[ext]; ok {
		return lang
	}
	return "plaintext"
}

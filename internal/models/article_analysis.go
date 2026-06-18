package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArticleAnalysis stores AI-generated metadata for a single article.
type ArticleAnalysis struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ArticleID     uuid.UUID `gorm:"type:uuid;index;not null" json:"article_id"`
	Summary       string    `gorm:"type:text" json:"summary"`
	Sentiment     string    `gorm:"size:16" json:"sentiment"` // positive | negative | neutral
	Hook          string    `gorm:"type:text" json:"hook"`
	Tweet         string    `gorm:"type:text" json:"tweet"`
	ThreadOpener  string    `gorm:"type:text" json:"thread_opener"`
	Model         string    `gorm:"size:64" json:"model"`
	TokensUsed    int       `json:"tokens_used"`
	CreatedAt     time.Time `json:"created_at"`
}

// BeforeCreate ensures a UUID is generated.
func (a *ArticleAnalysis) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

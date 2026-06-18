package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScrapeLog records the outcome of a single scraper run.
type ScrapeLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SourceCategory string         `gorm:"size:32;index" json:"source_category"`
	SourceName     string         `gorm:"size:64;index" json:"source_name"`
	StartedAt      time.Time      `gorm:"index" json:"started_at"`
	FinishedAt     *time.Time     `json:"finished_at"`
	ArticlesFound  int            `json:"articles_found"`
	ArticlesNew    int            `json:"articles_new"`
	Status         string         `gorm:"size:16;index" json:"status"` // running | success | failed
	ErrorMessage   string         `gorm:"type:text" json:"error_message"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate ensures a UUID is generated.
func (s *ScrapeLog) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.StartedAt.IsZero() {
		s.StartedAt = time.Now().UTC()
	}
	return nil
}

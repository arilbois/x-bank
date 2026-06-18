package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringSlice is a []string that is stored in PostgreSQL as a JSONB column.
// It satisfies sql.Scanner / driver.Valuer.
type StringSlice []string

func (s *StringSlice) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("StringSlice: unsupported scan type")
	}
	if len(b) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(b, s)
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Article is a single scraped news / blog post.
type Article struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SourceCategory string         `gorm:"size:32;index;not null" json:"source_category"`
	SourceName     string         `gorm:"size:64;index;not null" json:"source_name"`
	Title          string         `gorm:"type:text;not null" json:"title"`
	URL            string         `gorm:"type:text;uniqueIndex;not null" json:"url"`
	Excerpt        string         `gorm:"type:text" json:"excerpt"`
	Author         string         `gorm:"size:128" json:"author"`
	ImageURL       string         `gorm:"type:text" json:"image_url"`
	PublishedAt    *time.Time     `json:"published_at"`
	ScrapedAt      time.Time      `gorm:"index" json:"scraped_at"`
	Score          float64        `gorm:"index" json:"score"`
	Status         string         `gorm:"size:16;index;default:scraped" json:"status"`
	ContentHash    string         `gorm:"size:64;index" json:"content_hash"`
	Tags           StringSlice    `gorm:"type:jsonb" json:"tags"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate ensures a UUID is generated.
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.ScrapedAt.IsZero() {
		a.ScrapedAt = time.Now().UTC()
	}
	return nil
}

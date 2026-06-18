package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/arilbois/x-bank/internal/models"
)

// ScrapeLogRepository handles persistence for ScrapeLog entities.
type ScrapeLogRepository struct {
	db *gorm.DB
}

func NewScrapeLogRepository(db *gorm.DB) *ScrapeLogRepository {
	return &ScrapeLogRepository{db: db}
}

func (r *ScrapeLogRepository) Create(ctx context.Context, l *models.ScrapeLog) error {
	if err := r.db.WithContext(ctx).Create(l).Error; err != nil {
		return fmt.Errorf("create scrape log: %w", err)
	}
	return nil
}

func (r *ScrapeLogRepository) GetByID(ctx context.Context, id string) (*models.ScrapeLog, error) {
	var l models.ScrapeLog
	if err := r.db.WithContext(ctx).First(&l, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get scrape log by id: %w", err)
	}
	return &l, nil
}

func (r *ScrapeLogRepository) Update(ctx context.Context, l *models.ScrapeLog) error {
	if err := r.db.WithContext(ctx).Save(l).Error; err != nil {
		return fmt.Errorf("update scrape log: %w", err)
	}
	return nil
}

func (r *ScrapeLogRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ScrapeLog{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete scrape log: %w", err)
	}
	return nil
}

// ListFilter holds filters for List.
type ScrapeLogListFilter struct {
	Category string
	Source   string
	Status   string
	Page     int
	Limit    int
}

func (r *ScrapeLogRepository) List(ctx context.Context, f ScrapeLogListFilter) ([]models.ScrapeLog, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 200 {
		f.Limit = 20
	}
	q := r.db.WithContext(ctx).Model(&models.ScrapeLog{})
	if f.Category != "" {
		q = q.Where("source_category = ?", f.Category)
	}
	if f.Source != "" {
		q = q.Where("source_name = ?", f.Source)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count scrape logs: %w", err)
	}
	var out []models.ScrapeLog
	if err := q.Order("started_at DESC").
		Offset((f.Page - 1) * f.Limit).
		Limit(f.Limit).
		Find(&out).Error; err != nil {
		return nil, 0, fmt.Errorf("list scrape logs: %w", err)
	}
	return out, total, nil
}

// LatestForSource returns the most recent log for a given source, if any.
func (r *ScrapeLogRepository) LatestForSource(ctx context.Context, source string) (*models.ScrapeLog, error) {
	var l models.ScrapeLog
	if err := r.db.WithContext(ctx).
		Where("source_name = ?", source).
		Order("started_at DESC").
		First(&l).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("latest scrape log: %w", err)
	}
	return &l, nil
}

// Helper used by the orchestrator when finalising a run.
func (r *ScrapeLogRepository) Finish(ctx context.Context, id string, status string, articlesFound, articlesNew int, errMsg string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"finished_at":    &now,
		"status":         status,
		"articles_found": articlesFound,
		"articles_new":   articlesNew,
		"error_message":  errMsg,
	}
	if err := r.db.WithContext(ctx).
		Model(&models.ScrapeLog{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("finish scrape log: %w", err)
	}
	return nil
}

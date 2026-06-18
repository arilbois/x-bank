package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/arilbois/x-bank/internal/models"
)

// AnalysisRepository handles persistence for ArticleAnalysis entities.
type AnalysisRepository struct {
	db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

func (r *AnalysisRepository) Create(ctx context.Context, a *models.ArticleAnalysis) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("create analysis: %w", err)
	}
	return nil
}

func (r *AnalysisRepository) GetByID(ctx context.Context, id string) (*models.ArticleAnalysis, error) {
	var a models.ArticleAnalysis
	if err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get analysis by id: %w", err)
	}
	return &a, nil
}

func (r *AnalysisRepository) GetByArticleID(ctx context.Context, articleID string) (*models.ArticleAnalysis, error) {
	var a models.ArticleAnalysis
	if err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("created_at DESC").
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get analysis by article id: %w", err)
	}
	return &a, nil
}

func (r *AnalysisRepository) List(ctx context.Context, page, limit int) ([]models.ArticleAnalysis, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	var out []models.ArticleAnalysis
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.ArticleAnalysis{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count analyses: %w", err)
	}
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&out).Error; err != nil {
		return nil, 0, fmt.Errorf("list analyses: %w", err)
	}
	return out, total, nil
}

func (r *AnalysisRepository) Update(ctx context.Context, a *models.ArticleAnalysis) error {
	if err := r.db.WithContext(ctx).Save(a).Error; err != nil {
		return fmt.Errorf("update analysis: %w", err)
	}
	return nil
}

func (r *AnalysisRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ArticleAnalysis{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete analysis: %w", err)
	}
	return nil
}

// DeleteByArticleID removes all analysis rows attached to an article.
// Used when re-running the prompt so we never end up with stale rows.
func (r *AnalysisRepository) DeleteByArticleID(ctx context.Context, articleID string) error {
	if err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Delete(&models.ArticleAnalysis{}).Error; err != nil {
		return fmt.Errorf("delete analysis by article id: %w", err)
	}
	return nil
}

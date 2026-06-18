package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/arilbois/contentbank-v2/internal/models"
)

// ArticleRepository handles persistence for Article entities.
type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(ctx context.Context, a *models.Article) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("create article: %w", err)
	}
	return nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id string) (*models.Article, error) {
	var a models.Article
	if err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article by id: %w", err)
	}
	return &a, nil
}

func (r *ArticleRepository) GetByURL(ctx context.Context, url string) (*models.Article, error) {
	var a models.Article
	if err := r.db.WithContext(ctx).First(&a, "url = ?", url).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article by url: %w", err)
	}
	return &a, nil
}

func (r *ArticleRepository) GetByContentHash(ctx context.Context, hash string) (*models.Article, error) {
	var a models.Article
	if err := r.db.WithContext(ctx).First(&a, "content_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article by content hash: %w", err)
	}
	return &a, nil
}

func (r *ArticleRepository) GetTopScored(ctx context.Context, category string, limit int) ([]models.Article, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	q := r.db.WithContext(ctx).Order("score DESC, scraped_at DESC").Limit(limit)
	if category != "" {
		q = q.Where("source_category = ?", category)
	}
	var out []models.Article
	if err := q.Find(&out).Error; err != nil {
		return nil, fmt.Errorf("get top scored: %w", err)
	}
	return out, nil
}

func (r *ArticleRepository) GetByStatus(ctx context.Context, status string, limit int) ([]models.Article, error) {
	if limit < 1 || limit > 500 {
		limit = 50
	}
	var out []models.Article
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("scraped_at ASC").
		Limit(limit).
		Find(&out).Error; err != nil {
		return nil, fmt.Errorf("get by status: %w", err)
	}
	return out, nil
}

// ListFilter holds optional filters for List.
type ListFilter struct {
	Category string
	Source   string
	Status   string
	Sort     string // "recent" | "score"
	Page     int
	Limit    int
}

func (r *ArticleRepository) List(ctx context.Context, f ListFilter) ([]models.Article, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	q := r.db.WithContext(ctx).Model(&models.Article{})
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
		return nil, 0, fmt.Errorf("count articles: %w", err)
	}
	order := "scraped_at DESC"
	switch f.Sort {
	case "score":
		order = "score DESC, scraped_at DESC"
	case "oldest":
		order = "scraped_at ASC"
	}
	var out []models.Article
	if err := q.Order(order).
		Offset((f.Page - 1) * f.Limit).
		Limit(f.Limit).
		Find(&out).Error; err != nil {
		return nil, 0, fmt.Errorf("list articles: %w", err)
	}
	return out, total, nil
}

func (r *ArticleRepository) Update(ctx context.Context, a *models.Article) error {
	if err := r.db.WithContext(ctx).Save(a).Error; err != nil {
		return fmt.Errorf("update article: %w", err)
	}
	return nil
}

func (r *ArticleRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.Article{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	return nil
}

// MarkStatus updates only the Status column of an article. Used by the
// analyzer to flip scraped -> analyzed without touching the rest of the
// row.
func (r *ArticleRepository) MarkStatus(ctx context.Context, id, status string) error {
	res := r.db.WithContext(ctx).
		Model(&models.Article{}).
		Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("mark status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListUnanalyzed returns articles whose Status is "scraped" (i.e. the
// AI pass has not yet been run). Used by the batch analyze endpoint and
// the post-scrape scheduler hook.
func (r *ArticleRepository) ListUnanalyzed(ctx context.Context, category string, limit int) ([]models.Article, error) {
	if limit < 1 || limit > 500 {
		limit = 50
	}
	q := r.db.WithContext(ctx).
		Where("status = ?", "scraped").
		Order("scraped_at DESC").
		Limit(limit)
	if category != "" {
		q = q.Where("source_category = ?", category)
	}
	var out []models.Article
	if err := q.Find(&out).Error; err != nil {
		return nil, fmt.Errorf("list unanalyzed: %w", err)
	}
	return out, nil
}

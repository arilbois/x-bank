package scraper

import (
	"context"
	"time"
)

// Category constants. These are the three product categories
// ContentBank v2 aggregates.
const (
	CategorySambatWarga = "sambatWarga"
	CategoryPersibWay   = "persibWay"
	CategoryBytmod      = "bytmod"
)

// Scraper is the contract every source-specific scraper implements.
type Scraper interface {
	// Name is the stable identifier (e.g. "cnbc", "hackernews").
	Name() string
	// Category is one of the three product categories.
	Category() string
	// Scrape fetches the latest items from the source.
	Scrape(ctx context.Context) ([]ArticleInput, error)
}

// ArticleInput is the normalised shape produced by every scraper.
// Persistence + scoring are applied at the orchestrator level.
type ArticleInput struct {
	Title       string
	URL         string
	Excerpt     string
	Author      string
	ImageURL    string
	PublishedAt *time.Time
	Tags        []string
}

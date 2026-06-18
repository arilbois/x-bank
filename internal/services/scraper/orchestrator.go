package scraper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/arilbois/x-bank/internal/models"
	"github.com/arilbois/x-bank/internal/repositories"
	scorersvc "github.com/arilbois/x-bank/internal/services/scorer"
)

// Orchestrator runs every registered scraper, deduplicates by content
// hash, scores each new article, and writes a ScrapeLog per run.
//
// It can also be paired with an Analyzer: when one is configured via
// SetAnalyzer, every freshly created article is handed off to the AI
// immediately after the scrape, so the caller doesn't have to run a
// second pass to get a fully written, voice-aware version.
type Orchestrator struct {
	articles *repositories.ArticleRepository
	logs     *repositories.ScrapeLogRepository
	scrapers []Scraper
	analyzer AnalyzedArticle
}

// AnalyzedArticle is the interface the orchestrator needs from the
// analyzer. Defined as an interface so the orchestrator does not depend
// on the AI package (which would create an import cycle).
type AnalyzedArticle interface {
	AnalyzeArticle(ctx context.Context, a *models.Article) (*models.ArticleAnalysis, error)
}

// NewOrchestrator wires together the dependencies and the scraper set.
// If `only` is non-empty, only scrapers matching that category are run.
func NewOrchestrator(
	articles *repositories.ArticleRepository,
	logs *repositories.ScrapeLogRepository,
	allScrapers []Scraper,
) *Orchestrator {
	return &Orchestrator{
		articles: articles,
		logs:     logs,
		scrapers: allScrapers,
	}
}

// SetAnalyzer attaches an AI analyzer. Pass nil to disable the
// post-scrape analyze hook.
func (o *Orchestrator) SetAnalyzer(a AnalyzedArticle) {
	o.analyzer = a
}

// AllScrapers returns the canonical list of all built-in scrapers.
func AllScrapers() []Scraper {
	return []Scraper{
		NewCNBCScraper(),
		NewDetikScraper(),
		NewKompasScraper(),
		NewPersibOfficialScraper(),
		NewSimamaungScraper(),
		NewBolanetScraper(),
		NewHackerNewsScraper(),
		NewGitHubTrendingScraper(),
		NewDevToScraper(),
	}
}

// RunAll triggers every registered scraper concurrently and aggregates
// the results. Each scraper writes its own ScrapeLog entry.
func (o *Orchestrator) RunAll(ctx context.Context) error {
	return o.RunCategory(ctx, "")
}

// RunCategory triggers only scrapers of the given category (or all
// when category is empty).
func (o *Orchestrator) RunCategory(ctx context.Context, category string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(o.scrapers))

	for _, s := range o.scrapers {
		if category != "" && s.Category() != category {
			continue
		}
		wg.Add(1)
		go func(s Scraper) {
			defer wg.Done()
			if err := o.runOne(ctx, s); err != nil {
				slog.Error("scraper run failed", "source", s.Name(), "error", err)
				errCh <- fmt.Errorf("%s: %w", s.Name(), err)
			}
		}(s)
	}

	wg.Wait()
	close(errCh)

	// We deliberately do not propagate the first error so that one bad
	// source does not take down the whole run. Callers can inspect logs
	// to see which sources failed.
	for err := range errCh {
		slog.Warn("scraper error reported", "error", err)
	}
	return nil
}

func (o *Orchestrator) runOne(ctx context.Context, s Scraper) error {
	log := &models.ScrapeLog{
		SourceCategory: s.Category(),
		SourceName:     s.Name(),
		StartedAt:      time.Now().UTC(),
		Status:         "running",
	}
	if err := o.logs.Create(ctx, log); err != nil {
		return fmt.Errorf("create log: %w", err)
	}

	items, scrapeErr := s.Scrape(ctx)
	found := len(items)
	newCount := 0

	for _, it := range items {
		if it.URL == "" || it.Title == "" {
			continue
		}
		hash := contentHash(it)
		existing, err := o.articles.GetByContentHash(ctx, hash)
		if err != nil && err != repositories.ErrNotFound {
			return fmt.Errorf("dedupe lookup: %w", err)
		}
		if existing != nil {
			continue
		}

		article := &models.Article{
			SourceCategory: s.Category(),
			SourceName:     s.Name(),
			Title:          strings.TrimSpace(it.Title),
			URL:            it.URL,
			Excerpt:        it.Excerpt,
			Author:         it.Author,
			ImageURL:       it.ImageURL,
			PublishedAt:    it.PublishedAt,
			ScrapedAt:      time.Now().UTC(),
			ContentHash:    hash,
			Status:         "scraped",
			Tags:           models.StringSlice(it.Tags),
		}
		article.Score = scorersvc.ScoreArticle(article, time.Now().UTC())

		if err := o.articles.Create(ctx, article); err != nil {
			// Likely a unique-URL collision from a race; safe to skip.
			slog.Warn("article create failed", "url", article.URL, "error", err)
			continue
		}
		newCount++

		// Hand the freshly created article off to the AI for a
		// voice-aware rewrite. Failure here does not roll back the
		// article — the user can re-run the analyzer later.
		if o.analyzer != nil {
			aCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			if _, err := o.analyzer.AnalyzeArticle(aCtx, article); err != nil {
				slog.Warn("post-scrape analyze failed", "url", article.URL, "error", err)
			}
			cancel()
		}
	}

	status := "success"
	errMsg := ""
	if scrapeErr != nil {
		status = "failed"
		errMsg = scrapeErr.Error()
	}
	if err := o.logs.Finish(ctx, log.ID.String(), status, found, newCount, errMsg); err != nil {
		return fmt.Errorf("finish log: %w", err)
	}
	return scrapeErr
}

func contentHash(in ArticleInput) string {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(strings.TrimSpace(in.URL))))
	h.Write([]byte{0})
	h.Write([]byte(strings.ToLower(strings.TrimSpace(in.Title))))
	return hex.EncodeToString(h.Sum(nil))[:64]
}

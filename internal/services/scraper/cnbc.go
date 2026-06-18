package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// CNBCScraper scrapes cnbcindonesia.com news listings.
type CNBCScraper struct{}

func NewCNBCScraper() *CNBCScraper { return &CNBCScraper{} }

func (s *CNBCScraper) Name() string     { return "cnbc" }
func (s *CNBCScraper) Category() string { return CategorySambatWarga }

func (s *CNBCScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	// CNBC Indonesia wraps article cards in <article> blocks. We try a few
	// selectors so the scraper survives minor layout changes.
	c.OnHTML("article a[href], .list-news a[href], .gtm-headline a[href]", func(e *colly.HTMLElement) {
		href := cleanURL(e.Attr("href"))
		if href == "" || seen[href] {
			return
		}
		title := strings.TrimSpace(e.Text)
		if title == "" {
			title = strings.TrimSpace(e.Attr("title"))
		}
		if title == "" || len(title) < 10 {
			return
		}
		seen[href] = true
		out = append(out, ArticleInput{
			Title:       title,
			URL:         href,
			PublishedAt: ptrTime(time.Now().UTC()),
			Tags:        []string{"cnbc", "news"},
		})
	})

	if err := c.Visit("https://www.cnbcindonesia.com/"); err != nil {
		return out, fmt.Errorf("cnbc visit: %w", err)
	}
	return out, nil
}

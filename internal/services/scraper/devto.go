package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// DevToScraper scrapes dev.to front page listings.
type DevToScraper struct{}

func NewDevToScraper() *DevToScraper { return &DevToScraper{} }

func (s *DevToScraper) Name() string     { return "devto" }
func (s *DevToScraper) Category() string { return CategoryBytmod }

func (s *DevToScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("a.crayons-story__cover, h2 a, h3 a, .crayons-story__title a", func(e *colly.HTMLElement) {
		href := cleanURL(e.Attr("href"))
		if href == "" || seen[href] {
			return
		}
		title := strings.TrimSpace(e.Text)
		if title == "" || len(title) < 8 {
			return
		}
		seen[href] = true
		if !strings.HasPrefix(href, "http") {
			href = "https://dev.to" + href
		}
		out = append(out, ArticleInput{
			Title:       title,
			URL:         cleanURL(href),
			PublishedAt: ptrTime(time.Now().UTC()),
			Tags:        []string{"devto", "tech"},
		})
	})

	if err := c.Visit("https://dev.to/"); err != nil {
		return out, fmt.Errorf("devto visit: %w", err)
	}
	return out, nil
}

package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// PersibOfficialScraper scrapes persib.co.id.
type PersibOfficialScraper struct{}

func NewPersibOfficialScraper() *PersibOfficialScraper { return &PersibOfficialScraper{} }

func (s *PersibOfficialScraper) Name() string     { return "persib_official" }
func (s *PersibOfficialScraper) Category() string { return CategoryPersibWay }

func (s *PersibOfficialScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("a[href*='/news/'], a[href*='/berita/'], article a, .post a", func(e *colly.HTMLElement) {
		href := cleanURL(e.Attr("href"))
		if href == "" || seen[href] {
			return
		}
		title := strings.TrimSpace(e.Text)
		if title == "" {
			title = strings.TrimSpace(e.Attr("title"))
		}
		if title == "" || len(title) < 8 {
			return
		}
		seen[href] = true
		out = append(out, ArticleInput{
			Title:       title,
			URL:         href,
			PublishedAt: ptrTime(time.Now().UTC()),
			Tags:        []string{"persib", "official"},
		})
	})

	if err := c.Visit("https://www.persib.co.id/"); err != nil {
		return out, fmt.Errorf("persib visit: %w", err)
	}
	return out, nil
}

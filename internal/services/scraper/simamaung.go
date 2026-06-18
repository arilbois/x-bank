package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// SimamaungScraper scrapes simamaung.com.
type SimamaungScraper struct{}

func NewSimamaungScraper() *SimamaungScraper { return &SimamaungScraper{} }

func (s *SimamaungScraper) Name() string     { return "simamaung" }
func (s *SimamaungScraper) Category() string { return CategoryPersibWay }

func (s *SimamaungScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("h2 a, h3 a, .entry-title a, article a", func(e *colly.HTMLElement) {
		href := cleanURL(e.Attr("href"))
		if href == "" || seen[href] {
			return
		}
		title := strings.TrimSpace(e.Text)
		if title == "" || len(title) < 8 {
			return
		}
		seen[href] = true
		out = append(out, ArticleInput{
			Title:       title,
			URL:         href,
			PublishedAt: ptrTime(time.Now().UTC()),
			Tags:        []string{"persib", "simamaung"},
		})
	})

	if err := c.Visit("https://www.simamaung.com/"); err != nil {
		return out, fmt.Errorf("simamaung visit: %w", err)
	}
	return out, nil
}

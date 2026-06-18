package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// DetikScraper scrapes finance.detik.com listings.
type DetikScraper struct{}

func NewDetikScraper() *DetikScraper { return &DetikScraper{} }

func (s *DetikScraper) Name() string     { return "detik" }
func (s *DetikScraper) Category() string { return CategorySambatWarga }

func (s *DetikScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("article a[href], .list-content__item a[href], .media__title a[href]", func(e *colly.HTMLElement) {
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
			Tags:        []string{"detik", "finance"},
		})
	})

	if err := c.Visit("https://finance.detik.com/"); err != nil {
		return out, fmt.Errorf("detik visit: %w", err)
	}
	return out, nil
}

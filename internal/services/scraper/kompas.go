package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// KompasScraper scrapes kompas.com news listings.
type KompasScraper struct{}

func NewKompasScraper() *KompasScraper { return &KompasScraper{} }

func (s *KompasScraper) Name() string     { return "kompas" }
func (s *KompasScraper) Category() string { return CategorySambatWarga }

func (s *KompasScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("a.article__link, .latest--news a, .articleList a[href]", func(e *colly.HTMLElement) {
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
			Tags:        []string{"kompas", "news"},
		})
	})

	if err := c.Visit("https://www.kompas.com/"); err != nil {
		return out, fmt.Errorf("kompas visit: %w", err)
	}
	return out, nil
}

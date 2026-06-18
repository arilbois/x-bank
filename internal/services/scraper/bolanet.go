package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// BolanetScraper scrapes bola.net/persib section.
type BolanetScraper struct{}

func NewBolanetScraper() *BolanetScraper { return &BolanetScraper{} }

func (s *BolanetScraper) Name() string     { return "bolanet" }
func (s *BolanetScraper) Category() string { return CategoryPersibWay }

func (s *BolanetScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("a.news-list__link, h1 a, h2 a, h3 a, article a", func(e *colly.HTMLElement) {
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
			Tags:        []string{"persib", "bolanet"},
		})
	})

	if err := c.Visit("https://www.bola.net/persib/"); err != nil {
		return out, fmt.Errorf("bolanet visit: %w", err)
	}
	return out, nil
}

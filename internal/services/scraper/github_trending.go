package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// GitHubTrendingScraper scrapes github.com/trending.
type GitHubTrendingScraper struct{}

func NewGitHubTrendingScraper() *GitHubTrendingScraper { return &GitHubTrendingScraper{} }

func (s *GitHubTrendingScraper) Name() string     { return "github_trending" }
func (s *GitHubTrendingScraper) Category() string { return CategoryBytmod }

func (s *GitHubTrendingScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput
	seen := map[string]bool{}

	c.OnHTML("article.Box-row h2 a, article.Box-row h1 a", func(e *colly.HTMLElement) {
		href := cleanURL(e.Attr("href"))
		if href == "" || seen[href] {
			return
		}
		title := strings.TrimSpace(e.Text)
		title = strings.ReplaceAll(title, "\n", "")
		title = strings.TrimSpace(title)
		if title == "" {
			return
		}
		seen[href] = true
		full := "https://github.com" + href
		out = append(out, ArticleInput{
			Title:       title,
			URL:         cleanURL(full),
			Excerpt:     "GitHub trending repository",
			PublishedAt: ptrTime(time.Now().UTC()),
			Tags:        []string{"github", "trending"},
		})
	})

	if err := c.Visit("https://github.com/trending?since=daily"); err != nil {
		return out, fmt.Errorf("github trending visit: %w", err)
	}
	return out, nil
}

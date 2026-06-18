package scraper

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// HackerNewsScraper scrapes news.ycombinator.com.
type HackerNewsScraper struct{}

func NewHackerNewsScraper() *HackerNewsScraper { return &HackerNewsScraper{} }

func (s *HackerNewsScraper) Name() string     { return "hackernews" }
func (s *HackerNewsScraper) Category() string { return CategoryBytmod }

// Scrape extracts top stories from the front page.
func (s *HackerNewsScraper) Scrape(ctx context.Context) ([]ArticleInput, error) {
	c := newCollector()
	var out []ArticleInput

	c.OnHTML("tr.athing", func(row *colly.HTMLElement) {
		rankStr := row.ChildText("span.rank")
		rankStr = strings.TrimSuffix(strings.TrimSpace(rankStr), ".")
		_ = rankStr // rank not currently used

		titleLink := row.DOM.Find("a.titlelink, a.storylink")
		title := strings.TrimSpace(titleLink.Text())
		href, _ := titleLink.Attr("href")
		href = cleanURL(href)
		if title == "" || href == "" {
			return
		}
		// Subtext row (sibling) gives score + author + time.
		sub := row.DOM.Next()
		scoreStr := sub.Find("span.score").Text()
		author := sub.Find("a.hnuser").Text()
		age := sub.Find("span.age a").Text()

		var score int
		if scoreStr != "" {
			parts := strings.Fields(scoreStr)
			if len(parts) > 0 {
				score, _ = strconv.Atoi(parts[0])
			}
		}

		out = append(out, ArticleInput{
			Title:       title,
			URL:         href,
			Author:      strings.TrimSpace(author),
			Excerpt:     fmt.Sprintf("score=%d | %s", score, strings.TrimSpace(age)),
			PublishedAt: ptrTime(time.Now().UTC()),
			Tags:        []string{"hackernews", "tech"},
		})
	})

	if err := c.Visit("https://news.ycombinator.com/"); err != nil {
		return out, fmt.Errorf("hn visit: %w", err)
	}
	return out, nil
}

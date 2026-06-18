package scraper

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// BaseURL helpers & a Colly builder shared by all scrapers.
const (
	userAgent = "Mozilla/5.0 (compatible; ContentBankBot/2.0; +https://github.com/arilbois/x-bank)"
)

func newCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowURLRevisit(),
	)
	c.SetRequestTimeout(15 * time.Second)
	c.OnError(func(r *colly.Response, err error) {
		slog.Warn("scraper request failed", "url", r.Request.URL, "error", err)
	})
	return c
}

func mustParse(u string) *url.URL {
	p, err := url.Parse(u)
	if err != nil {
		panic(fmt.Errorf("parse %q: %w", u, err))
	}
	return p
}

// cleanURL strips tracking query params.
func cleanURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	q := parsed.Query()
	for _, k := range []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"} {
		q.Del(k)
	}
	parsed.RawQuery = q.Encode()
	return strings.TrimSpace(parsed.String())
}

func ptrTime(t time.Time) *time.Time { return &t }

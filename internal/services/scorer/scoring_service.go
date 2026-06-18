package scorer

import (
	"strings"
	"time"

	"github.com/arilbois/contentbank-v2/internal/models"
)

// Keyword maps per source category. Used by the keyword_match factor.
var keywordMap = map[string][]string{
	"sambatWarga": {
		"pemilu", "politik", "ekonomi", "inflasi", "harga", "subsidi",
		"bpjs", "kesehatan", "pendidikan", "uu", "pemerintah", "presiden",
		"dpr", "mahkamah", "korupsi", "kpk", "unjuk rasa", "demo",
		"krisis", "bansos", "pangan", "energi", "ikn",
	},
	"persibWay": {
		"persib", "maung", "bandung", "vikingo", "bobotoh", "liga 1",
		"bri liga", "indonesian", "sepak bola", "stadion", "geprek",
		"sidulloh", "bojan", "david da silva", "ciro", "el clasico",
		"kebon sirih", "persija", "bonek", "arema", "pssi",
	},
	"bytmod": {
		"ai", "llm", "gpt", "startup", "open source", "github",
		"rust", "go", "python", "kubernetes", "docker", "webdev",
		"backend", "frontend", "database", "postgres", "redis", "api",
		"indie hacker", "show hn", "launch hn",
	},
}

// sourceWeight assigns a constant bonus per source name.
var sourceWeight = map[string]float64{
	"persib_official": 20,
	"simamaung":       18,
	"bolanet":         15,
	"hackernews":      15,
	"github_trending": 12,
	"devto":           10,
	"cnbc":            10,
	"detik":           8,
	"kompas":          8,
}

// ScoreInput is the minimum surface needed to score an article.
// Defined locally so the scorer has no dependency on the models package
// (keeps the package importable in tests without a DB).
type ScoreInput struct {
	SourceCategory string
	SourceName     string
	Title          string
	Excerpt        string
	PublishedAt    *time.Time
}

// Score returns a rule-based score in the 0..100 range.
func Score(in ScoreInput, now time.Time) float64 {
	var total float64

	// 1) keyword match
	kws, ok := keywordMap[in.SourceCategory]
	if ok {
		haystack := strings.ToLower(in.Title + " " + in.Excerpt)
		for _, kw := range kws {
			if strings.Contains(haystack, kw) {
				total += 30
				break // only count once
			}
		}
	}

	// 2) source weight
	if w, ok := sourceWeight[in.SourceName]; ok {
		total += w
	}

	// 3) recency
	if in.PublishedAt != nil {
		age := now.Sub(*in.PublishedAt)
		switch {
		case age <= time.Hour:
			total += 20
		case age <= 6*time.Hour:
			total += 15
		case age <= 24*time.Hour:
			total += 10
		case age <= 72*time.Hour:
			total += 5
		default:
			total += 0
		}
	}

	// clamp to [0, 100]
	if total < 0 {
		total = 0
	}
	if total > 100 {
		total = 100
	}
	return total
}

// ScoreArticle is a convenience wrapper for scoring a *models.Article
// against the supplied "now" reference.
func ScoreArticle(a *models.Article, now time.Time) float64 {
	return Score(ScoreInput{
		SourceCategory: a.SourceCategory,
		SourceName:     a.SourceName,
		Title:          a.Title,
		Excerpt:        a.Excerpt,
		PublishedAt:    a.PublishedAt,
	}, now)
}

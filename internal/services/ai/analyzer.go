package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/arilbois/contentbank-v2/internal/models"
	"github.com/arilbois/contentbank-v2/internal/repositories"
)

// Analyzer turns scraped articles into structured AnalysisResult records.
// The article body is NEVER sent to the AI — only title + excerpt.
//
// Each category gets its own VoiceProfile injected into the prompt so
// the analysis reads in the user's tone for the niche.
type Analyzer struct {
	provider Provider
	analyses *repositories.AnalysisRepository
	articles *repositories.ArticleRepository
	voices   map[string]VoiceProfile
}

func NewAnalyzer(p Provider, ar *repositories.AnalysisRepository, artRepo *repositories.ArticleRepository, voices map[string]VoiceProfile) *Analyzer {
	if voices == nil {
		voices = DefaultVoices()
	}
	return &Analyzer{provider: p, analyses: ar, articles: artRepo, voices: voices}
}

// BuildPrompt is exported so tests can assert that the prompt contains
// only title + excerpt and never the article body. The category is
// optional; empty falls back to a neutral voice.
func BuildPrompt(title, excerpt, category string) string {
	t := strings.TrimSpace(title)
	e := strings.TrimSpace(excerpt)
	if len(e) > 600 {
		e = e[:600]
	}

	profile, _ := VoiceFor(category, DefaultVoices())
	if category == "" {
		profile.Niche = "umum"
	}

	return fmt.Sprintf(`Kamu adalah penulis redaksi untuk niche: %s.

GAYA BICARA:
%s

AUDIENS:
%s

FORMAT OUTPUT (JSON):
%s

ATURAN WAJIB:
%s

HINDARI:
%s

HASHTAG/STYLE:
%s

Kamu akan menganalisis artikel di bawah ini. Jawab HANYA dengan JSON valid, tanpa markdown, tanpa komentar, tanpa preamble.

Schema:
%s

Article title: %s
Article excerpt: %s`,
		profile.Niche,
		profile.Tone,
		profile.Audience,
		profile.Format,
		profile.Rules,
		profile.Avoid,
		profile.Hashtags,
		profile.OutputFields,
		t, e,
	)
}

var jsonBlock = regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)\\s*```")

// extractJSON strips markdown fences and returns the first JSON object in
// the supplied text. Falls back to returning the text unchanged if no
// braces are found.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if m := jsonBlock.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}

// AnalyzeArticle runs the prompt, parses the JSON response and persists
// an ArticleAnalysis record. Existing analysis for the same article is
// overwritten (a fresh scrape re-runs the prompt).
func (a *Analyzer) AnalyzeArticle(ctx context.Context, article *models.Article) (*models.ArticleAnalysis, error) {
	if a.provider == nil {
		return nil, fmt.Errorf("ai provider not configured")
	}
	prompt := BuildPrompt(article.Title, article.Excerpt, article.SourceCategory)
	content, err := a.provider.Analyze(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("provider analyze: %w", err)
	}

	var res AnalysisResult
	if err := json.Unmarshal([]byte(extractJSON(content)), &res); err != nil {
		return nil, fmt.Errorf("parse ai response: %w (raw=%s)", err, truncate(content, 200))
	}

	// Wipe any prior analysis for this article so we don't end up with
	// stale rows when the prompt format evolves.
	_ = a.analyses.DeleteByArticleID(ctx, article.ID.String())

	rec := &models.ArticleAnalysis{
		ArticleID:    article.ID,
		Summary:      res.Summary,
		Sentiment:    normaliseSentiment(res.Sentiment),
		Hook:         res.Hook,
		Tweet:        res.Tweet,
		ThreadOpener: res.ThreadOpener,
		Model:        a.provider.Name(),
	}
	if err := a.analyses.Create(ctx, rec); err != nil {
		return nil, fmt.Errorf("persist analysis: %w", err)
	}
	if err := a.articles.MarkStatus(ctx, article.ID.String(), "analyzed"); err != nil {
		slog.Warn("mark analysed failed", "article_id", article.ID, "error", err)
	}
	slog.Info("article analysed", "article_id", article.ID, "category", article.SourceCategory, "sentiment", rec.Sentiment)
	return rec, nil
}

func normaliseSentiment(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "positive", "positif":
		return "positive"
	case "negative", "negatif":
		return "negative"
	default:
		return "neutral"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

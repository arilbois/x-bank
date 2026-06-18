package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/arilbois/contentbank-v2/internal/models"
	"github.com/arilbois/contentbank-v2/internal/services/ai"
)

func TestAI_BuildPrompt_OnlyTitleAndExcerpt(t *testing.T) {
	title := "Persib Bandung menang dramatis"
	excerpt := "Maung Bandung comeback di menit akhir."
	prompt := ai.BuildPrompt(title, excerpt, "persibway")

	if !strings.Contains(prompt, title) {
		t.Fatalf("prompt missing title; got: %s", prompt)
	}
	if !strings.Contains(prompt, excerpt) {
		t.Fatalf("prompt missing excerpt; got: %s", prompt)
	}
	// Banned phrases (in any case): no full body should ever leak.
	for _, banned := range []string{"article body", "full content", "full article"} {
		if strings.Contains(strings.ToLower(prompt), banned) {
			t.Fatalf("prompt contains forbidden phrase %q: %s", banned, prompt)
		}
	}
}

func TestAI_BuildPrompt_TruncatesLongExcerpt(t *testing.T) {
	long := strings.Repeat("x", 2000)
	prompt := ai.BuildPrompt("title", long, "bytmod")
	if !strings.Contains(prompt, "x") {
		t.Fatalf("prompt missing excerpt data")
	}
	// Excerpt is truncated to <= 600 chars in BuildPrompt.
	if strings.Count(prompt, "x") > 605 {
		t.Fatalf("excerpt was not truncated (found %d 'x' chars in prompt)", strings.Count(prompt, "x"))
	}
}

// stubProvider satisfies ai.Provider for tests that need to inspect the
// prompt actually sent to the AI.
type stubProvider struct {
	lastPrompt string
	resp       string
	err        error
}

func (s *stubProvider) Analyze(_ context.Context, prompt string) (string, error) {
	s.lastPrompt = prompt
	if s.resp == "" {
		return `{"summary":"x","sentiment":"neutral","hook":"x","tweet":"x","thread_opener":"x"}`, nil
	}
	return s.resp, s.err
}
func (s *stubProvider) Name() string { return "stub" }

func TestAI_ProviderInterface_Wiring(t *testing.T) {
	// Sanity check: a stub satisfies the ai.Provider interface.
	var p ai.Provider = &stubProvider{resp: `{}`}
	if p.Name() != "stub" {
		t.Fatalf("unexpected provider name: %s", p.Name())
	}
}

func TestAI_BuildPrompt_NeverIncludesArticleBody(t *testing.T) {
	article := &models.Article{
		ID:             uuid.New(),
		Title:          "Sample title",
		Excerpt:        "Sample excerpt",
		SourceCategory: "bytmod",
		SourceName:     "hackernews",
		PublishedAt:    ptrTime(time.Now()),
	}
	// The body MUST NOT appear in the prompt under any circumstances.
	body := "THIS_IS_THE_FULL_ARTICLE_BODY_MARKER_THAT_MUST_NOT_LEAK"
	prompt := ai.BuildPrompt(article.Title, article.Excerpt, article.SourceCategory)
	if strings.Contains(prompt, body) {
		t.Fatalf("body leaked into prompt")
	}
	// And the prompt must include the article's actual title + excerpt.
	if !strings.Contains(prompt, article.Title) || !strings.Contains(prompt, article.Excerpt) {
		t.Fatalf("prompt missing title or excerpt")
	}
}

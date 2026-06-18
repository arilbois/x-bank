package tests

import (
	"testing"
	"time"

	"github.com/arilbois/x-bank/internal/services/scorer"
)

func TestScore_ClampedToRange(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		in   scorer.ScoreInput
		min  float64
		max  float64
	}{
		{
			name: "fresh persib keyword match → high score",
			in: scorer.ScoreInput{
				SourceCategory: "persibWay",
				SourceName:     "persib_official",
				Title:          "Persib Bandung menang dramatis di Liga 1",
				Excerpt:        "Maung Bandung berhasil comeback di menit akhir",
				PublishedAt:    ptrTime(now.Add(-30 * time.Minute)),
			},
			min: 60, // 30 (keyword) + 20 (source) + 20 (recency) = 70
			max: 80,
		},
		{
			name: "old cnbc with no keyword match",
			in: scorer.ScoreInput{
				SourceCategory: "sambatWarga",
				SourceName:     "cnbc",
				Title:          "Cuaca Jakarta hari ini",
				Excerpt:        "BMKG memprediksi hujan ringan",
				PublishedAt:    ptrTime(now.Add(-7 * 24 * time.Hour)),
			},
			min: 8,  // source=10, recency=0
			max: 12,
		},
		{
			name: "unknown source, no recency → 0",
			in: scorer.ScoreInput{
				SourceCategory: "bytmod",
				SourceName:     "unknown_source",
				Title:          "Hello world",
			},
			min: 0,
			max: 0,
		},
		{
			name: "hackernews with ai keyword",
			in: scorer.ScoreInput{
				SourceCategory: "bytmod",
				SourceName:     "hackernews",
				Title:          "Show HN: I built a new LLM in Rust",
				Excerpt:        "open source and fast",
				PublishedAt:    ptrTime(now.Add(-3 * time.Hour)),
			},
			min: 50, // 30 (keyword) + 15 (source) + 15 (recency <6h) = 60
			max: 65,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := scorer.Score(tc.in, now)
			if got < tc.min || got > tc.max {
				t.Fatalf("score %v out of expected range [%v, %v]", got, tc.min, tc.max)
			}
		})
	}
}

func TestScore_RecencyAffectsScore(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	fresh := scorer.Score(scorer.ScoreInput{
		SourceCategory: "bytmod",
		SourceName:     "hackernews",
		Title:          "Anything",
		PublishedAt:    ptrTime(now.Add(-10 * time.Minute)),
	}, now)
	old := scorer.Score(scorer.ScoreInput{
		SourceCategory: "bytmod",
		SourceName:     "hackernews",
		Title:          "Anything",
		PublishedAt:    ptrTime(now.Add(-5 * 24 * time.Hour)),
	}, now)
	if fresh <= old {
		t.Fatalf("expected fresh (%v) > old (%v)", fresh, old)
	}
}

func TestScore_KeywordMatchAddsPoints(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	with := scorer.Score(scorer.ScoreInput{
		SourceCategory: "persibWay",
		SourceName:     "persib_official",
		Title:          "Persib Bandung bobotoh penuh semangat",
		PublishedAt:    ptrTime(now.Add(-1 * time.Hour)),
	}, now)
	without := scorer.Score(scorer.ScoreInput{
		SourceCategory: "persibWay",
		SourceName:     "persib_official",
		Title:          "Random title with no niche keyword",
		PublishedAt:    ptrTime(now.Add(-1 * time.Hour)),
	}, now)
	if with-without < 25 {
		t.Fatalf("expected keyword match to add ~30 points, got diff=%v", with-without)
	}
}

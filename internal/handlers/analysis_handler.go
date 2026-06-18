package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arilbois/x-bank/internal/repositories"
	"github.com/arilbois/x-bank/internal/services/ai"
)

type AnalysisHandler struct {
	analyses *repositories.AnalysisRepository
	articles *repositories.ArticleRepository
	analyzer *ai.Analyzer
}

func NewAnalysisHandler(
	a *repositories.AnalysisRepository,
	articles *repositories.ArticleRepository,
	analyzer *ai.Analyzer,
) *AnalysisHandler {
	return &AnalysisHandler{analyses: a, articles: articles, analyzer: analyzer}
}

// GetByID returns the analysis for an analysis row.
func (h *AnalysisHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	a, err := h.analyses.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

// GetByArticleID returns the most recent analysis for an article.
func (h *AnalysisHandler) GetByArticleID(c *gin.Context) {
	id := c.Param("id")
	a, err := h.analyses.GetByArticleID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

// Trigger runs the AI analyzer for a single article and returns the
// fresh result. Idempotent: any prior analysis is overwritten.
func (h *AnalysisHandler) Trigger(c *gin.Context) {
	id := c.Param("id")
	article, err := h.articles.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	rec, err := h.analyzer.AnalyzeArticle(ctx, article)
	if err != nil {
		slog.Error("analyze failed", "article_id", id, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rec})
}

type batchAnalyzeRequest struct {
	Category string `json:"category"`
	Limit    int    `json:"limit"`
}

// BatchAnalyze picks up unanalyzed articles (optionally filtered by
// category) and runs the AI prompt against each. The endpoint is
// synchronous but caps the batch to 50 by default so a runaway run
// cannot pin the request forever.
func (h *AnalysisHandler) BatchAnalyze(c *gin.Context) {
	var req batchAnalyzeRequest
	_ = c.ShouldBindJSON(&req) // body optional
	if req.Limit <= 0 {
		req.Limit = 50
	}
	articles, err := h.articles.ListUnanalyzed(c.Request.Context(), req.Category, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type result struct {
		ArticleID string `json:"article_id"`
		Title     string `json:"title"`
		OK        bool   `json:"ok"`
		Error     string `json:"error,omitempty"`
	}
	results := make([]result, 0, len(articles))

	success := 0
	ctx := c.Request.Context()
	for i := range articles {
		a := articles[i]
		ctxI, cancel := context.WithTimeout(ctx, 90*time.Second)
		rec, err := h.analyzer.AnalyzeArticle(ctxI, &a)
		cancel()
		if err != nil {
			slog.Warn("batch analyze failed", "article_id", a.ID, "error", err)
			results = append(results, result{ArticleID: a.ID.String(), Title: a.Title, OK: false, Error: err.Error()})
			continue
		}
		_ = rec
		results = append(results, result{ArticleID: a.ID.String(), Title: a.Title, OK: true})
		success++
	}

	c.JSON(http.StatusOK, gin.H{
		"requested_category": req.Category,
		"total":              len(articles),
		"success":            success,
		"failed":             len(articles) - success,
		"results":            results,
	})
}

// limitFromQuery returns an int from a Gin query string, falling back
// to def if missing/invalid.
func limitFromQuery(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

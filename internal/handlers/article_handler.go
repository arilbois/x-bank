package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/arilbois/contentbank-v2/internal/repositories"
)

type ArticleHandler struct {
	articles *repositories.ArticleRepository
}

func NewArticleHandler(a *repositories.ArticleRepository) *ArticleHandler {
	return &ArticleHandler{articles: a}
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// List returns a paginated, filterable list of articles.
func (h *ArticleHandler) List(c *gin.Context) {
	filter := repositories.ListFilter{
		Category: c.Query("category"),
		Source:   c.Query("source"),
		Status:   c.Query("status"),
		Sort:     c.Query("sort"),
		Page:     parseIntDefault(c.Query("page"), 1),
		Limit:    parseIntDefault(c.Query("limit"), 20),
	}
	items, total, err := h.articles.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// GetByID returns one article.
func (h *ArticleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	a, err := h.articles.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

// Trending returns top-scored articles.
func (h *ArticleHandler) Trending(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 20)
	category := c.Query("category")
	items, err := h.articles.GetTopScored(c.Request.Context(), category, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "count": len(items)})
}

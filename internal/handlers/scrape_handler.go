package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arilbois/x-bank/internal/services/scraper"
)

type ScrapeHandler struct {
	orchestrator *scraper.Orchestrator
}

func NewScrapeHandler(o *scraper.Orchestrator) *ScrapeHandler {
	return &ScrapeHandler{orchestrator: o}
}

type runRequest struct {
	Category string `json:"category"`
}

// Run triggers a scrape run. Admin-only at the route level.
func (h *ScrapeHandler) Run(c *gin.Context) {
	var req runRequest
	_ = c.ShouldBindJSON(&req) // body is optional
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	if err := h.orchestrator.RunCategory(ctx, req.Category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

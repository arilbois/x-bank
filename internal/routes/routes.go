package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/arilbois/x-bank/internal/handlers"
	"github.com/arilbois/x-bank/internal/middleware"
	"github.com/arilbois/x-bank/internal/services/auth"
)

// Deps bundles every dependency the router needs.
type Deps struct {
	Auth        *auth.Service
	AuthHandler *handlers.AuthHandler
	Article     *handlers.ArticleHandler
	Analysis    *handlers.AnalysisHandler
	Scrape      *handlers.ScrapeHandler
}

// New builds the Gin engine and wires every route.
func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Public.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "x-bank"})
	})
	r.POST("/auth/login", d.AuthHandler.Login)

	// Authenticated.
	authed := r.Group("/")
	authed.Use(middleware.RequireAuth(d.Auth))
	{
		authed.GET("/articles", d.Article.List)
		authed.GET("/articles/:id", d.Article.GetByID)
		authed.GET("/trending", d.Article.Trending)

		authed.GET("/analysis/:id", d.Analysis.GetByID)
		// alias: GET /articles/:id/analysis
		authed.GET("/articles/:id/analysis", d.Analysis.GetByArticleID)
	}

	// Admin-only.
	admin := r.Group("/")
	admin.Use(middleware.RequireAuth(d.Auth), middleware.RequireRole("admin"))
	{
		admin.POST("/scrape/run", d.Scrape.Run)
		admin.POST("/articles/:id/analyze", d.Analysis.Trigger)
		admin.POST("/analyze/batch", d.Analysis.BatchAnalyze)
	}

	return r
}

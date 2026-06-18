package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"

	"github.com/arilbois/contentbank-v2/internal/config"
	"github.com/arilbois/contentbank-v2/internal/database"
	"github.com/arilbois/contentbank-v2/internal/handlers"
	"github.com/arilbois/contentbank-v2/internal/repositories"
	"github.com/arilbois/contentbank-v2/internal/routes"
	"github.com/arilbois/contentbank-v2/internal/scheduler"
	"github.com/arilbois/contentbank-v2/internal/services/ai"
	"github.com/arilbois/contentbank-v2/internal/services/auth"
	"github.com/arilbois/contentbank-v2/internal/services/scraper"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	slog.Info("config loaded", "env", cfg.AppEnv, "port", cfg.AppPort)

	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			slog.Warn("db close", "error", err)
		}
	}()

	// Seed admin if missing.
	if err := seedAdmin(db, cfg); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	// Wire dependencies.
	userRepo := repositories.NewUserRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	analysisRepo := repositories.NewAnalysisRepository(db)
	logRepo := repositories.NewScrapeLogRepository(db)

	authSvc := auth.NewService(userRepo, cfg.JWTSecret, cfg.JWTTTL)

	// AI provider: only constructed if configured.
	var aiProvider ai.Provider
	var analyzer *ai.Analyzer
	if cfg.AIBaseURL != "" {
		aiProvider = ai.NewOpenAICompatibleProvider(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel)
		analyzer = ai.NewAnalyzer(aiProvider, analysisRepo, articleRepo, ai.DefaultVoices())
	} else {
		slog.Warn("AI_BASE_URL is empty; ai endpoints will not be usable until configured")
	}

	orch := scraper.NewOrchestrator(articleRepo, logRepo, scraper.AllScrapers())
	if analyzer != nil {
		orch.SetAnalyzer(analyzer)
		slog.Info("post-scrape analyze hook enabled")
	}
	sched := scheduler.New(orch)
	if err := sched.Start(); err != nil {
		return fmt.Errorf("scheduler start: %w", err)
	}
	defer sched.Stop()

	// Router.
	r := routes.New(routes.Deps{
		Auth:        authSvc,
		AuthHandler: handlers.NewAuthHandler(authSvc),
		Article:     handlers.NewArticleHandler(articleRepo),
		Analysis:    handlers.NewAnalysisHandler(analysisRepo, articleRepo, analyzer),
		Scrape:      handlers.NewScrapeHandler(orch),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server in a goroutine.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for signal or fatal server error.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case sig := <-stop:
		slog.Info("signal received, shutting down", "signal", sig.String())
	}

	// Graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}

func seedAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		slog.Warn("ADMIN_USERNAME/ADMIN_PASSWORD not set; skipping admin seed")
		return nil
	}
	users := repositories.NewUserRepository(db)
	if _, err := users.GetByUsername(context.Background(), cfg.AdminUsername); err == nil {
		return nil // already seeded
	}
	svc := auth.NewService(users, cfg.JWTSecret, cfg.JWTTTL)
	u, err := svc.Register(context.Background(), cfg.AdminUsername, cfg.AdminPassword, "admin")
	if err != nil {
		return err
	}
	slog.Info("admin user seeded", "username", u.Username)
	return nil
}

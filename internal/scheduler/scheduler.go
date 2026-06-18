package scheduler

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"

	"github.com/arilbois/contentbank-v2/internal/services/scraper"
)

// Scheduler runs scrape jobs on a cron schedule.
type Scheduler struct {
	orchestrator *scraper.Orchestrator
	cron         *cron.Cron
}

// New constructs a Scheduler. Use Start/Stop to control its lifecycle.
func New(o *scraper.Orchestrator) *Scheduler {
	return &Scheduler{
		orchestrator: o,
		cron: cron.New(cron.WithSeconds(), cron.WithLogger(
			cron.DefaultLogger,
		)),
	}
}

// Start registers the three category jobs and starts the cron runner.
func (s *Scheduler) Start() error {
	jobs := []struct {
		spec     string
		category string
		label    string
	}{
		{"0 */15 * * * *", scraper.CategorySambatWarga, "news (sambatWarga)"},
		{"0 */10 * * * *", scraper.CategoryPersibWay, "persib (persibWay)"},
		{"0 */30 * * * *", scraper.CategoryBytmod, "tech (bytmod)"},
	}

	for _, j := range jobs {
		spec := j.spec
		category := j.category
		label := j.label
		_, err := s.cron.AddFunc(spec, func() {
			slog.Info("cron job fired", "job", label, "category", category)
			ctx, cancel := context.WithTimeout(context.Background(), 2*60*1000*1000*1000) // 2 min
			defer cancel()
			if err := s.orchestrator.RunCategory(ctx, category); err != nil {
				slog.Error("scheduled run failed", "job", label, "error", err)
			}
		})
		if err != nil {
			return err
		}
	}
	s.cron.Start()
	slog.Info("scheduler started")
	return nil
}

// Stop halts the cron runner. Safe to call multiple times.
func (s *Scheduler) Stop() {
	if s.cron != nil {
		stopCtx := s.cron.Stop()
		<-stopCtx.Done()
		slog.Info("scheduler stopped")
	}
}

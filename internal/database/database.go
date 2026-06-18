package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/arilbois/x-bank/internal/config"
	"github.com/arilbois/x-bank/internal/models"
)

// Connect opens a GORM connection, runs AutoMigrate for all known models,
// and returns the *gorm.DB handle.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	gormLogger := logger.New(
		slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo),
		logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("acquire sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Ping to make sure the DB is actually reachable.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := db.WithContext(ctx).AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.ArticleAnalysis{},
		&models.ScrapeLog{},
	); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}

	return db, nil
}

// Close releases the underlying database connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("acquire sql.DB: %w", err)
	}
	return sqlDB.Close()
}

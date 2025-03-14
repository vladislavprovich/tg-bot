package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Import for initializing the PostgreSQL driver.
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Response time. If longer returns an error.
// Example: 'waitingTime = 10' ---> wait 10 seconds.
const waitingTime = 10

func PrepareConnection(ctx context.Context, config Config, logger *logrus.Logger) (*sql.DB, error) {
	if err := config.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validate Postgres config: %w", err)
	}

	db, err := sql.Open("postgres", config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	ctxPing, cancelPing := context.WithTimeout(ctx, waitingTime*time.Second)
	defer cancelPing()

	if err = db.PingContext(ctxPing); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err = ensureTables(ctx, db, config, logger); err != nil {
		return nil, fmt.Errorf("ensure tables: %w", err)
	}

	return db, nil
}

func ensureTables(ctx context.Context, db *sql.DB, cfg Config, logger *logrus.Logger) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, cfg.EnsureIdxTimeout)
	defer cancel()

	queries := []string{
		// Create Users table
		`
    CREATE TABLE IF NOT EXISTS users (
        user_id UUID PRIMARY KEY,
        telegram_id BIGINT UNIQUE NOT NULL,
        username VARCHAR(100) UNIQUE NOT NULL
    );
    `,
		// Create URLs table
		`
    CREATE TABLE IF NOT EXISTS urls (
        url_id SERIAL PRIMARY KEY,
        user_id UUID NOT NULL,
        original_url TEXT NOT NULL,
        short_url VARCHAR(255) UNIQUE NOT NULL,
        redirect_count INT,
        created_at TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
    );
    `,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctxTimeout, query); err != nil {
			logger.WithFields(logrus.Fields{"query": query, "error": err}).Error("Failed to execute query")
			return fmt.Errorf("execute query: %w", err)
		}
	}

	return nil
}

package repository

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations applies all pending database migrations.
// Each .sql file must contain "-- Up" and "-- Down" sections.
// Applied migrations are tracked in the schema_migrations table.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	log.Info().Msg("running database migrations")

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migration files: %w", err)
	}

	var filenames []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			filenames = append(filenames, e.Name())
		}
	}
	sort.Strings(filenames)

	applied := 0
	for _, name := range filenames {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("checking migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("reading migration file %s: %w", name, err)
		}

		upSQL := extractUpSection(string(content))
		if upSQL == "" {
			return fmt.Errorf("migration %s has no -- Up section", name)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning transaction for %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, upSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("applying migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, name,
		); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("recording migration %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("committing migration %s: %w", name, err)
		}

		log.Info().Str("migration", name).Msg("applied migration")
		applied++
	}

	log.Info().Int("applied", applied).Int("total", len(filenames)).Msg("migrations completed")
	return nil
}

// extractUpSection returns the SQL between "-- Up" and "-- Down" markers.
func extractUpSection(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUp := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- Up" {
			inUp = true
			continue
		}
		if trimmed == "-- Down" {
			break
		}
		if inUp {
			upLines = append(upLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(upLines, "\n"))
}

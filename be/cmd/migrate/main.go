package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"minhquang/be/internal/config"
	"minhquang/be/internal/db"
)

func main() {
	config.LoadDotEnv()

	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./cmd/migrate up|down")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := ensureMigrationTable(ctx, pool); err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "up":
		err = run(ctx, pool, "*.up.sql", true)
	case "down":
		err = run(ctx, pool, "*.down.sql", false)
	default:
		err = errors.New("usage: go run ./cmd/migrate up|down")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func ensureMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT NOW()
		)
	`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func run(ctx context.Context, pool *pgxpool.Pool, pattern string, up bool) error {
	files, err := filepath.Glob(filepath.Join("migrations", pattern))
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}
	sort.Strings(files)
	if strings.Contains(pattern, ".down.") {
		reverse(files)
	}

	for _, file := range files {
		version := migrationVersion(file)
		applied, err := isApplied(ctx, pool, version)
		if err != nil {
			return err
		}
		if up && applied {
			log.Printf("skipped %s (already applied)", file)
			continue
		}
		if !up && !applied {
			log.Printf("skipped %s (not applied)", file)
			continue
		}

		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin %s: %w", file, err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("run %s: %w", file, err)
		}
		if up {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version)
		} else {
			_, err = tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", file, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", file, err)
		}
		log.Printf("applied %s", file)
	}

	return nil
}

func migrationVersion(file string) string {
	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".up.sql")
	return strings.TrimSuffix(name, ".down.sql")
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var applied bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&applied); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return applied, nil
}

func reverse[T any](items []T) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

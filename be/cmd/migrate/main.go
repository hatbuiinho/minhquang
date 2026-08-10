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

	"reminder/be/internal/config"
	"reminder/be/internal/db"
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

	switch os.Args[1] {
	case "up":
		err = run(ctx, pool, "*.up.sql")
	case "down":
		err = run(ctx, pool, "*.down.sql")
	default:
		err = errors.New("usage: go run ./cmd/migrate up|down")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, pool *pgxpool.Pool, pattern string) error {
	files, err := filepath.Glob(filepath.Join("migrations", pattern))
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}
	sort.Strings(files)
	if strings.Contains(pattern, ".down.") {
		reverse(files)
	}

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("run %s: %w", file, err)
		}
		log.Printf("applied %s", file)
	}

	return nil
}

func reverse[T any](items []T) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"reminder/be/internal/account"
	"reminder/be/internal/config"
	"reminder/be/internal/db"
	"reminder/be/internal/device"
	"reminder/be/internal/event"
	"reminder/be/internal/httpapi"
	"reminder/be/internal/ota"
	"reminder/be/internal/push"
	"reminder/be/internal/reminder"
)

func main() {
	config.LoadDotEnv()

	addr := env("ADDR", ":8080")
	ctx := context.Background()

	stores := newStores(ctx)
	if stores.pool != nil {
		defer stores.pool.Close()
	}

	eventService := event.NewService(stores.events, time.Now)
	deviceService := device.NewService(stores.devices, time.Now)
	accountService := account.NewService(stores.accounts, time.Now)
	otaStorageDir := env("OTA_STORAGE_DIR", "storage/ota")
	otaService := ota.NewService(otaStorageDir)
	eventService.SetRecipientResolver(accountService)
	startReminderWorker(ctx, stores.events, deviceService)
	router := httpapi.NewRouter(eventService, deviceService, accountService, otaService, otaStorageDir)

	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func startReminderWorker(ctx context.Context, events event.Store, devices *device.Service) {
	enabled, err := strconv.ParseBool(env("REMINDER_WORKER_ENABLED", "false"))
	if err != nil || !enabled {
		return
	}

	sender, err := push.NewFirebaseSender(
		ctx,
		os.Getenv("FIREBASE_PROJECT_ID"),
		os.Getenv("FIREBASE_SERVICE_ACCOUNT_FILE"),
	)
	if err != nil {
		log.Fatalf("create firebase sender: %v", err)
	}

	interval, err := time.ParseDuration(env("REMINDER_WORKER_INTERVAL", "30s"))
	if err != nil {
		log.Fatalf("parse REMINDER_WORKER_INTERVAL: %v", err)
	}

	batchSize, err := strconv.Atoi(env("REMINDER_WORKER_BATCH_SIZE", "25"))
	if err != nil {
		log.Fatalf("parse REMINDER_WORKER_BATCH_SIZE: %v", err)
	}

	worker := reminder.NewWorker(events, devices, sender, time.Now, reminder.Config{
		BatchSize: batchSize,
		Logger:    log.Default(),
	})

	log.Printf("reminder worker enabled; interval=%s batch_size=%d", interval, batchSize)
	go worker.Run(ctx, interval)
}

type stores struct {
	pool     *pgxpool.Pool
	events   event.Store
	devices  device.Store
	accounts account.Store
}

func newStores(ctx context.Context) stores {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("DATABASE_URL is not set; using in-memory stores")
		return stores{
			events:   event.NewMemoryStore(),
			devices:  device.NewMemoryStore(),
			accounts: account.NewMemoryStore(),
		}
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	log.Printf("using postgres stores")
	return stores{
		pool:     pool,
		events:   event.NewPostgresStore(pool),
		devices:  device.NewPostgresStore(pool),
		accounts: account.NewPostgresStore(pool),
	}
}

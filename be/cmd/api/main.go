package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"minhquang/be/internal/config"
	"minhquang/be/internal/db"
	"minhquang/be/internal/department"
	"minhquang/be/internal/device"
	"minhquang/be/internal/httpapi"
	"minhquang/be/internal/ota"
	"minhquang/be/internal/user"
	"minhquang/be/internal/volunteer"
)

func main() {
	config.LoadDotEnv()

	addr := env("ADDR", ":8080")
	ctx := context.Background()

	stores := newStores(ctx)
	if stores.pool != nil {
		defer stores.pool.Close()
	}

	deviceService := device.NewService(stores.devices, time.Now)
	userService := user.NewService(stores.users, time.Now)
	volunteerService := volunteer.NewService(stores.volunteers, time.Now)
	departmentService := department.NewService(stores.departments, time.Now)
	volunteerService.SetDepartmentResolver(departmentService)
	if err := userService.EnsureInitialAdmin(ctx, user.CreateInput{
		Username:    os.Getenv("INITIAL_ADMIN_USERNAME"),
		DisplayName: env("INITIAL_ADMIN_DISPLAY_NAME", "Ban quản trị"),
		Password:    os.Getenv("INITIAL_ADMIN_PASSWORD"),
	}); err != nil {
		log.Fatalf("create initial admin: %v", err)
	}
	otaStorageDir := env("OTA_STORAGE_DIR", "storage/ota")
	otaService := ota.NewService(otaStorageDir)
	router := httpapi.NewRouter(deviceService, userService, volunteerService, departmentService, otaService, otaStorageDir)

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

type stores struct {
	pool        *pgxpool.Pool
	devices     device.Store
	users       user.Store
	volunteers  volunteer.Store
	departments department.Store
}

func newStores(ctx context.Context) stores {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("DATABASE_URL is not set; using in-memory stores")
		return stores{
			devices:     device.NewMemoryStore(),
			users:       user.NewMemoryStore(),
			volunteers:  volunteer.NewMemoryStore(),
			departments: department.NewMemoryStore(),
		}
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	log.Printf("using postgres stores")
	return stores{
		pool:        pool,
		devices:     device.NewPostgresStore(pool),
		users:       user.NewPostgresStore(pool),
		volunteers:  volunteer.NewPostgresStore(pool),
		departments: department.NewPostgresStore(pool),
	}
}

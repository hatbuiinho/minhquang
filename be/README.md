# Reminder API

## Development

Use the root `.env` file for backend configuration. Set `DATABASE_URL` there to
your Neon connection string.

```sh
cd ..
cp .env.example .env
```

Do not commit `.env`.

Run migrations:

```sh
go run ./cmd/migrate up
```

Start the API:

```sh
air
```

If `DATABASE_URL` is not set, the API falls back to the in-memory event store.

## Verification

```sh
gofmt -w ./cmd ./internal
go test ./...
go build -o /tmp/reminder-api ./cmd/api
go build -o /tmp/reminder-migrate ./cmd/migrate
```

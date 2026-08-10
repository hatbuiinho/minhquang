# Minh Quang Personnel API

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

Set `INITIAL_ADMIN_USERNAME` and `INITIAL_ADMIN_PASSWORD` before the first start.
If `DATABASE_URL` is not set, the API falls back to in-memory stores for local development.

## Verification

```sh
gofmt -w ./cmd ./internal
go test ./...
go build -o /tmp/minhquang-api ./cmd/api
go build -o /tmp/minhquang-migrate ./cmd/migrate
```

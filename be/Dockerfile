FROM golang:1.25-alpine AS build

WORKDIR /src/be

COPY be/go.mod be/go.sum ./
RUN go mod download

COPY be/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY be/migrations /app/migrations

USER app

EXPOSE 8080
CMD ["/app/api"]

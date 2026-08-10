SHELL := /bin/sh

COMPOSE_ENV_FILE ?= .env
COMPOSE := COMPOSE_ENV_FILE=$(COMPOSE_ENV_FILE) docker compose --env-file $(COMPOSE_ENV_FILE)
COMPOSE_TOOLS := $(COMPOSE) --profile tools
OTA_VERSION ?=
ENV_FILE ?= .env
DEPLOYGATE_APK_PATH ?= fe/android/app/build/outputs/apk/release/app-release.apk

.PHONY: help
help:
	@printf '%s\n' 'Available targets:'
	@printf '%s\n' '  make be-check          Run backend tests and build'
	@printf '%s\n' '  make be-test           Run backend tests'
	@printf '%s\n' '  make be-build          Build backend binary'
	@printf '%s\n' '  make be-migrate        Run local backend migrations'
	@printf '%s\n' '  make fe-check          Run frontend format, lint, check, build'
	@printf '%s\n' '  make fe-sync-android   Build frontend and sync Capacitor Android'
	@printf '%s\n' '  make android-build     Build Android debug APK'
	@printf '%s\n' '  make android-keystore  Create Android release signing keystore'
	@printf '%s\n' '  make android-release   Build signed Android release APK'
	@printf '%s\n' '  make deploygate-upload Upload Android APK to DeployGate'
	@printf '%s\n' '  make android-deploygate Build release APK and upload to DeployGate'
	@printf '%s\n' '  make app-check         Run backend, frontend, Android checks'
	@printf '%s\n' '  make compose-config    Validate docker compose config'
	@printf '%s\n' '  make compose-build     Build docker images'
	@printf '%s\n' '  make compose-up        Start API with docker compose'
	@printf '%s\n' '  make compose-down      Stop docker compose services'
	@printf '%s\n' '  make compose-logs      Follow API logs'
	@printf '%s\n' '  make compose-migrate   Run migrations in docker compose'
	@printf '%s\n' '  make ota-publish       Publish OTA bundle locally'
	@printf '%s\n' '  make compose-ota       Publish OTA bundle in docker compose'

.PHONY: be-test
be-test:
	cd be && go test ./...

.PHONY: be-build
be-build:
	cd be && go build -o /tmp/minhquang-api ./cmd/api

.PHONY: be-check
be-check: be-test be-build

.PHONY: be-migrate
be-migrate:
	cd be && go run ./cmd/migrate up

.PHONY: fe-format
fe-format:
	cd fe && yarn format

.PHONY: fe-lint
fe-lint:
	cd fe && yarn lint

.PHONY: fe-svelte-check
fe-svelte-check:
	cd fe && yarn run check

.PHONY: fe-build
fe-build:
	cd fe && yarn build

.PHONY: fe-check
fe-check: fe-format fe-lint fe-svelte-check fe-build

.PHONY: fe-sync-android
fe-sync-android:
	cd fe && yarn build && npx cap sync android

.PHONY: android-build
android-build:
	cd fe/android && ./gradlew assembleDebug

.PHONY: android-keystore
android-keystore:
	./scripts/create-android-keystore.sh

.PHONY: android-release
android-release:
	ENV_FILE="$(ENV_FILE)" ./scripts/build-android-release.sh

.PHONY: deploygate-upload
deploygate-upload:
	ENV_FILE="$(ENV_FILE)" DEPLOYGATE_APK_PATH="$(DEPLOYGATE_APK_PATH)" ./scripts/upload-deploygate.sh

.PHONY: android-deploygate
android-deploygate: android-release deploygate-upload

.PHONY: app-check
app-check: be-check fe-check fe-sync-android android-build

.PHONY: compose-config
compose-config:
	$(COMPOSE_TOOLS) config

.PHONY: compose-build
compose-build:
	$(COMPOSE_TOOLS) build api ota-publisher

.PHONY: compose-up
compose-up:
	$(COMPOSE) up -d --build api

.PHONY: compose-down
compose-down:
	$(COMPOSE) down

.PHONY: compose-logs
compose-logs:
	$(COMPOSE) logs -f api

.PHONY: compose-migrate
compose-migrate:
	$(COMPOSE) run --rm api /app/migrate up

.PHONY: ota-publish
ota-publish:
	cd fe && OTA_VERSION="$(OTA_VERSION)" yarn ota:publish

.PHONY: compose-ota
compose-ota:
	$(COMPOSE_TOOLS) run --rm -e OTA_VERSION="$(OTA_VERSION)" ota-publisher

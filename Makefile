.PHONY: sqlc migrate-up migrate-down migrate-status run compose-up compose-down compose-logs compose-ps

# Host-side targets (migrate-*, run) read DATABASE_URL from .env.
# Compose loads .env via env_file; Make does not unless we export it here.
ifneq (,$(wildcard .env))
include .env
export DATABASE_URL HTTP_ADDR HTTP_PORT POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_PORT
endif

sqlc:
	sqlc generate

migrate-up:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir db/migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir db/migrations postgres "$(DATABASE_URL)" status

run:
	go run ./cmd/builder-api

# Full local stack: postgres + goose migrate + builder-api
# Compose auto-loads .env for interpolation; services also use env_file: .env
compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f

compose-ps:
	docker compose ps

.PHONY: sqlc migrate-up migrate-down migrate-status run compose-up compose-down compose-logs compose-ps

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

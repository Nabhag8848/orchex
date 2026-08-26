.PHONY: sqlc migrate-up migrate-down migrate-status run run-execution run-worker test compose-up compose-down compose-logs compose-ps seed-local sam-local

# Host-side targets (migrate-*, run) read DATABASE_URL from .env.
# Compose loads .env via env_file; Make does not unless we export it here.
ifneq (,$(wildcard .env))
include .env
export DATABASE_URL HTTP_ADDR HTTP_PORT EXECUTION_HTTP_PORT WORKER_HTTP_PORT POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_PORT SQS_QUEUE_URL AWS_REGION AWS_ENDPOINT_URL AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY FUNCTION_SANDBOX_ARN LAMBDA_ENDPOINT_URL
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

run-execution:
	go run ./cmd/execution-api

run-worker:
	go run ./cmd/execution-worker

# Local Lambda sandbox (needs Docker). Pair with make run-worker / compose worker.
sam-local:
	sam local start-lambda --port 3001 --host 0.0.0.0

test:
	go test ./...

# Full local stack: postgres + goose migrate + ElasticMQ + builder-api + execution-api + execution-worker
# Compose auto-loads .env for interpolation; services also use env_file: .env
compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f

compose-ps:
	docker compose ps

# Local builder-api only. Creates draft / publishable / published / archived workflows.
seed-local:
	./scripts/seed-local-workflows.sh

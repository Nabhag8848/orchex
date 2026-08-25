# Orchex

**A durable workflow orchestration + execution engine — we own both**

Build a graph. Publish an immutable version. Run it reliably. Resume exactly where it failed.

Design decisions, schemas, benches, and learning notes live in [docs/](./docs/).

## Local setup

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose
- [Go 1.26+](https://go.dev/dl/) (only if you run the API on the host)
- [goose](https://github.com/pressly/goose) and [sqlc](https://docs.sqlc.dev/) (only for host-side migrate / codegen)

### 1. Configure env

```bash
cp .env.example .env
```

`.env` is used by Compose and by host tools (`make migrate-*`, `make run`).  
`DATABASE_URL` uses `localhost` for the host; Compose rewrites it to the `postgres` service name inside the network.

### 2. Start the stack

```bash
make compose-up
```

This starts:

1. **postgres** — PostgreSQL 17 (`postgres:17-alpine`)
2. **migrate** — goose applies `db/migrations`
3. **builder-api** — workflows API on host port `8080` (after migrate succeeds)
4. **execution-api** — execution API on host port `8081` (after migrate succeeds)
5. **execution-worker** — internal worker on host port `8082`
6. **sqs** — ElasticMQ (SQS-compatible) on host port `9324`

Compose lives at the repo root. Image definitions are under [`docker/`](./docker/) (`Dockerfile*.local` for Compose; distroless `Dockerfile*` for ECS). Build context is still the repository root.

Useful commands:

```bash
make compose-ps
make compose-logs
make compose-down
```

### 3. Smoke check

```bash
curl http://localhost:8080/health/builder
# → {"status":"ok"}

curl http://localhost:8081/health/execution
# → {"status":"ok"}

curl http://localhost:8082/health/worker
# → {"status":"ok"}
```

Local SQS (ElasticMQ) — send a message, then watch the worker:

```bash
aws --endpoint-url http://localhost:9324 sqs send-message \
  --queue-url http://localhost:9324/000000000000/orchex-node-jobs \
  --message-body '{"run_id":"smoke","node_id":"smoke","attempt":1,"source":"cli"}' \
  --region ap-south-1

docker compose logs -f execution-worker
```

Dummy AWS keys in `.env` (`test` / `test`) are enough for ElasticMQ.

Override the execution host port with `EXECUTION_HTTP_PORT` in `.env` (default `8081`), and the worker with `WORKER_HTTP_PORT` (default `8082`).

Workflows live in the `public.workflows` table. Seed local samples that hit **public APIs** (httpbin, Open-Meteo, JSONPlaceholder) and use **all five node types** on every non-empty graph:

```bash
make seed-local
```

Then:

```bash
curl -sS http://localhost:8080/v1/workflows | jq '.items[] | {id, name, status}'
curl -sS http://localhost:8080/v1/workflows/<uuid> | jq '.graph.nodes[] | {name, node_type, config}'
```

The script prints more builder curls after seeding.

### Optional: API on the host

With Compose Postgres already up:

```bash
make migrate-status
```

Host APIs (need `DATABASE_URL` in `.env`):

```bash
make run              # builder-api on HTTP_ADDR (default :8080)
make run-execution    # execution-api
make run-worker       # execution-worker
```

After changing SQL queries or migrations:

```bash
make sqlc
```

## Local vs production

|                 | Local                                                                                                             | Production                                                                                                                 |
| --------------- | ----------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| **Compute**     | Docker Compose (`docker/Dockerfile.local`, `docker/Dockerfile.execution.local`, `docker/Dockerfile.worker.local`) | ECS Fargate (`docker/Dockerfile` / `docker/Dockerfile.execution` / `docker/Dockerfile.worker` — `linux/amd64`, distroless) |
| **Database**    | `postgres:17-alpine` in Compose                                                                                   | Amazon RDS for PostgreSQL 17                                                                                               |
| **Queue**       | ElasticMQ (`softwaremill/elasticmq-native`) on host port `9324`, queue `orchex-node-jobs`                         | AWS SQS `orchex-node-jobs` + DLQ (14-day retention, DLQ after 5 receives)                                                  |
| **Function JS** | No sandbox Lambda; worker logs `sandbox: skip invoke (local)`                                                     | Shared zip Lambda `orchex-function-sandbox` (`nodejs24.x`); worker sync `Invoke` (`FUNCTION_SANDBOX_ARN`)                  |
| **Config**      | `.env` from `.env.example` (dummy AWS keys + `AWS_ENDPOINT_URL` for ElasticMQ)                                    | Terraform task definition + task role ([infra/](./infra/)). No dummy keys; `AWS_ENDPOINT_URL` unset                        |
| **Migrations**  | goose one-shot `migrate` service on compose up                                                                    | `aws ecs run-task` on `orchex-db-migrate` (see [infra/README.md](./infra/README.md#run-database-migrations))               |
| **TLS to DB**   | `sslmode=disable`                                                                                                 | `sslmode=require` (via `orchex/DATABASE_URL` secret)                                                                       |
| **Networking**  | localhost ports `5432` / `8080` / `8081` / `8082` / `9324`                                                        | ALB path rules → APIs; worker is internal (no ALB); ECS talks to RDS and SQS in AWS                                        |

Infra (ECR, ALB, ECS, RDS, SQS, Lambda sandbox, Secrets Manager) is managed with Terraform under [infra/](./infra/) — see [infra/README.md](./infra/README.md) for create, migrate, deploy, and destroy.

## Design docs

| Path                                               | What it is                                                             |
| -------------------------------------------------- | ---------------------------------------------------------------------- |
| [docs/README.md](./docs/README.md)                 | Full design narrative (requirements → HLD → API → schema → deep dives) |
| [docs/orchex.excalidraw](./docs/orchex.excalidraw) | Architecture boards (source of truth for diagrams)                     |
| [docs/schema.dbml](./docs/schema.dbml)             | PostgreSQL OLTP model                                                  |
| [docs/bench/postgres](./docs/bench/postgres)       | OLTP capacity harness                                                  |
| [docs/node-type-schemas](./docs/node-type-schemas) | JSON Schema contracts for node types                                   |
| [docs/data-structure](./docs/data-structure)       | Graph experiments behind the DAG deep dive                             |

## License

[MIT](./LICENSE)

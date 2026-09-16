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

`.env` is loaded by Compose and by the Makefile for host commands (`make migrate-*`, `make run`). The application reads the resulting process environment directly, so production can inject variables without any environment-specific code.
`DATABASE_URL` uses `localhost` for the host; Compose rewrites it to the `postgres` service name inside the network.
Set `LOG_LEVEL` to `debug`, `info`, `warn`, or `error`. Compose and Terraform default to `debug`; the application defaults to `info`. Services emit JSON logs to stdout.

### 2. Start the stack

```bash
make compose-up
```

This starts:

1. **postgres** — PostgreSQL 17 (`postgres:17-alpine`)
2. **migrate** — goose applies `db/migrations`
3. **builder-api** — workflows API on host port `8080` (after migrate succeeds)
4. **execution-api** — execution API on host port `8081` (after migrate succeeds)
5. **execution-worker** — internal worker with no published host port
6. **sqs** — ElasticMQ (SQS-compatible) on host port `9324`

Compose lives at the repo root. Image definitions are under [`docker/`](./docker/) (`Dockerfile*.local` for Compose; distroless `Dockerfile*` for ECS). Build context is still the repository root.

Useful commands:

```bash
make compose-ps
make compose-logs
make compose-down
```

### Scale workers up or down

With the stack already running, set the desired number of worker containers:

```bash
# Scale up to four workers total.
docker compose up -d --no-deps --scale execution-worker=4 execution-worker

# Scale down to one worker.
docker compose up -d --no-deps --scale execution-worker=1 execution-worker

# Scale down to zero workers (stop consuming jobs).
docker compose up -d --no-deps --scale execution-worker=0 execution-worker

# Check replica health and follow logs from all workers.
docker compose ps execution-worker
docker compose logs -f execution-worker
```

All replicas consume the same SQS queue. These commands use the existing image; add `--build` to rebuild after code changes. After removing the stack with `make compose-down`, include the desired `--scale` value when starting it again.

The worker has no fixed host port, so replicas can run together without port conflicts. Each container checks its own `/health/worker` endpoint on internal port `8080`.


### 3. Smoke check

```bash
curl http://localhost:8080/health/builder
# → {"status":"ok"}

curl http://localhost:8081/health/execution
# → {"status":"ok"}

docker compose exec --index 1 execution-worker curl -fsS http://127.0.0.1:8080/health/worker
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

### SQS relay

The execution API runs an in-process relay. Every second it reads due rows from
`run_node_jobs_outbox`, locks them with `FOR UPDATE SKIP LOCKED`, and sends jobs
for `pending` and `running` runs to SQS. Each message contains `run_id`,
`workflow_version_id`, `node_id`, and `attempt`. A successfully sent row is
deleted; rows for paused, cancelled, completed, or failed runs are deleted
without being sent. If sending fails, the database transaction rolls back and
the row is retried on a later tick.

Override the execution host port with `EXECUTION_HTTP_PORT` in `.env` (default `8081`). The Compose worker does not publish a host port.

Workflows live in the `public.workflows` table. Seed local samples that hit **public APIs** (httpbin, Open-Meteo, JSONPlaceholder) and use **all five node types** on every non-empty graph:

```bash
make seed-local
```

To seed three valid drafts and three published examples in production after applying migrations:

```bash
BUILDER_URL="http://YOUR_ALB_HOST" make seed-production
```

This skips invalid and archived fixtures. It prints execution commands without running workflows; reruns create additional records.

Then:

```bash
curl -sS http://localhost:8080/v1/workflows | jq '.items[] | {id, name, status}'
curl -sS http://localhost:8080/v1/workflows/<uuid> | jq '.graph.nodes[] | {name, node_type, config}'
```

The script prints builder curls plus directly runnable `POST /v1/runs` examples for the published workflows.

### Execute a workflow locally

The seeded workflows include a Function node, so start the Lambda emulator before invoking one:

```bash
make sam-local
```

In another terminal, run `make seed-local`, then copy one of its printed `POST /v1/runs` commands. The request body is ordinary JSON: `payload` is the input object, not a quoted JSON string and not wrapped in another `data` object.

```bash
WORKFLOW_ID=<published-workflow-id>

RUN_ID="$(curl -sS -X POST http://localhost:8081/v1/runs \
  -H 'Content-Type: application/json' \
  -d "{\"workflow_id\":\"$WORKFLOW_ID\",\"payload\":{\"order_id\":\"ord_9001\",\"email\":\"buyer@example.com\",\"amount_cents\":14999,\"currency\":\"usd\"}}" \
  | jq -r '.id')"

curl -sS "http://localhost:8081/v1/runs/$RUN_ID" | jq
```

Starting a run creates it as `pending`, records `started_at`, and writes the Start node job to the outbox in the same transaction. The relay sends that job to SQS; the worker processes one node hop at a time, persists its output, and writes the next node job to the outbox. A Response node marks the run `completed` and sets `completed_at`.

The current worker executes these node types:

| Node type | Runtime behavior |
| --- | --- |
| `start` | Forwards the input envelope to the default edge. |
| `function` | Invokes the shared Lambda sandbox with JavaScript source and the previous output. Sandbox code receives the prior `data` object as `data`. |
| `conditional` | Evaluates a strict `==` expression against the previous output and follows its `true` or `false` edge. |
| `api` | Makes the configured HTTP request and stores response status, headers, and body as output. |
| `response` | Produces the final response output and completes the workflow run. |

For API and Response nodes, `body_template` is the base object. Fields from the previous output's `data` object are merged on top, so received data wins when both contain the same top-level field. Node configuration contracts live in [`docs/node-type-schemas/`](./docs/node-type-schemas/).

### Worker debug trace

Compose defaults to `LOG_LEVEL=debug`. Follow the worker while running a seeded workflow:

```bash
docker compose logs -f --tail=0 execution-worker
```

Each hop logs the run and node IDs, node type, executor input, executor output, selected edge, outbox enqueue, and final completion. These payload-level events are `debug` only; ECS runs at `info` by default. Do not enable `debug` in production for workflows containing sensitive payload data.

### Optional: API on the host

With Compose Postgres already up:

```bash
make migrate-status
```

Host APIs (need `DATABASE_URL` in `.env`):

```bash
make run              # builder-api on HTTP_ADDR (default :8080)
make run-execution    # execution-api
make run-worker       # execution-worker (needs SAM below for Function sandbox)
```

Local Function sandbox (Docker + [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)):

```bash
make sam-local        # sam local start-lambda on :3001 (template.yaml)
```

`make sam-local` uses SAM's `--skip-pull-image` flag, so it reuses a cached Lambda runtime image instead of pulling it on every start. Pull the image manually first when the runtime changes: `sam local start-lambda --port 3001 --host 0.0.0.0`.

Keep that running, then `make run-worker` (or Compose worker). SQS stays on ElasticMQ (`AWS_ENDPOINT_URL`); Lambda uses SAM (`LAMBDA_ENDPOINT_URL`). Production leaves both endpoint vars unset.

After changing SQL queries or migrations:

```bash
make sqlc
```

## Local vs production

|                 | Local                                                                                                                | Production                                                                                                                 |
| --------------- | -------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| **Compute**     | Docker Compose (`docker/Dockerfile.local`, `docker/Dockerfile.execution.local`, `docker/Dockerfile.worker.local`)    | ECS Fargate (`docker/Dockerfile` / `docker/Dockerfile.execution` / `docker/Dockerfile.worker` — `linux/amd64`, distroless) |
| **Database**    | `postgres:17-alpine` in Compose                                                                                      | Amazon RDS for PostgreSQL 17                                                                                               |
| **Queue**       | ElasticMQ (`softwaremill/elasticmq-native`) on host port `9324`, queue `orchex-node-jobs`                            | AWS SQS `orchex-node-jobs` + DLQ (14-day retention, DLQ after 5 receives)                                                  |
| **Function JS** | SAM local (`make sam-local`) → `orchex-function-sandbox`; worker uses `LAMBDA_ENDPOINT_URL` + `FUNCTION_SANDBOX_ARN` | Shared zip Lambda `orchex-function-sandbox` (`nodejs24.x`); worker sync `Invoke` (`FUNCTION_SANDBOX_ARN` from Terraform)   |
| **Config**      | Make and Compose load `.env` (dummy keys; `AWS_ENDPOINT_URL` → ElasticMQ; `LAMBDA_ENDPOINT_URL` → SAM; `LOG_LEVEL=debug`) | Terraform task definition + task role ([infra/](./infra/)); `LOG_LEVEL=debug`; injected environment variables are used directly |
| **Migrations**  | goose one-shot `migrate` service on compose up                                                                       | `aws ecs run-task` on `orchex-db-migrate` (see [infra/README.md](./infra/README.md#run-database-migrations))               |
| **TLS to DB**   | `sslmode=disable`                                                                                                    | `sslmode=require` (via `orchex/DATABASE_URL` secret)                                                                       |
| **Networking**  | localhost ports `5432` / `8080` / `8081` / `9324` / `3001` (SAM); worker port `8080` is internal                      | ALB path rules → APIs; worker is internal (no ALB); ECS talks to RDS, SQS, and Lambda in AWS                               |

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

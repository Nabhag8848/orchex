# Orchex

**A durable workflow execution engine**

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
3. **builder-api** — API on port `8080` (after migrate succeeds)

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
```

Workflows live in the `public.workflows` table. Example:

```bash
curl http://localhost:8080/v1/workflows/<uuid>
```

### Optional: API on the host

With Compose Postgres already up:

```bash
make migrate-status
```

After changing SQL queries or migrations:

```bash
make sqlc
```

## Local vs production

|                | Local                                          | Production                                                                    |
| -------------- | ---------------------------------------------- | ----------------------------------------------------------------------------- |
| **Compute**    | Docker Compose (`Dockerfile.local`)            | ECS Fargate (`Dockerfile` — `linux/amd64`, distroless)                        |
| **Database**   | `postgres:17-alpine` in Compose                | Amazon RDS for PostgreSQL 17                                                  |
| **Config**     | `.env` from `.env.example`                     | Task definition / secrets (see `.env.production.example`)                     |
| **Migrations** | goose one-shot `migrate` service on compose up | goose against RDS (run as a release step / job; not baked into the API image) |
| **TLS to DB**  | `sslmode=disable`                              | `sslmode=require`                                                             |
| **Networking** | localhost ports `5432` / `8080`                | ALB → ECS; ECS talks to RDS in the VPC                                        |

Infra (ECR, ALB, ECS, RDS) is managed with Terraform under [infra/](./infra/).

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

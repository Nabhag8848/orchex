# Orchex infrastructure

Terraform for AWS resources. Provisions:

- **ECR** — `orchex-builder-api`, `orchex-execution-api`, `orchex-execution-worker`, and `orchex-db-migrate` (goose migrations)
- **ECS Fargate** — shared cluster, builder and execution **API services** (behind ALB), internal **execution-worker** (no ALB), and migrate **run-task** definition
- **SQS** — standard queue `orchex-node-jobs` plus DLQ (14-day retention, DLQ after 5 receives)
- **Lambda** — shared zip function `orchex-function-sandbox` (Node.js); only the execution worker may `Invoke`. Terraform zips [`modules/lambda_sandbox/src`](modules/lambda_sandbox/src).
- **Secrets Manager** — shared `orchex/DATABASE_URL` (Postgres URL for ECS tasks)
- **RDS PostgreSQL 17** — `orchex-postgres` (private, master password in Secrets Manager)
- **ALB** — HTTP :80 with path rules to builder and execution on `:8080`; unmatched paths return 404. The worker is not on the ALB.

Terraform creates and wires the infrastructure. Container images are built with Docker and pushed to ECR separately (see [Lifecycle](#lifecycle)).

## Layout

```text
infra/
  terraform.tf      # required Terraform / provider versions
  providers.tf      # AWS provider (region, profile, default tags)
  variables.tf      # shared variables (region, environment)
  main.tf           # root modules (ECR, ALB, ECS, ECS services, SQS, Lambda sandbox, migrate task, RDS, secrets)
  outputs.tf        # root outputs
  modules/
    ecr/            # ECR repository for container images
    alb/            # internet-facing ALB + target groups + listener rules
    ecs/            # shared Fargate cluster + data_plane_client SG
    ecs_service/    # Fargate service (optional ALB attachment)
    ecs_run_task/   # one-shot Fargate task definition (migrations via run-task)
    sqs/            # node-jobs queue + DLQ + resource policies
    lambda_sandbox/ # shared Function-node JS sandbox (zip + src/index.js)
    secrets_manager/# Secrets Manager secret + version
    rds/            # PostgreSQL RDS (subnet group, SG, instance)
```

### Root modules (`main.tf`)

| Module                 | AWS resources                           | Purpose                                        |
| ---------------------- | --------------------------------------- | ---------------------------------------------- |
| `ecr_builder_api`      | ECR `orchex-builder-api`                | Builder API image                              |
| `ecr_execution_api`    | ECR `orchex-execution-api`              | Execution API image                            |
| `ecr_execution_worker` | ECR `orchex-execution-worker`           | Execution worker image                         |
| `ecr_db_migrate`       | ECR `orchex-db-migrate`                 | Goose migrate image                            |
| `alb`                  | ALB, listener, rules, target groups     | Public HTTP → builder / execution APIs         |
| `ecs`                  | Fargate cluster, `data_plane_client` SG | Shared compute + RDS client identity           |
| `rds`                  | RDS PostgreSQL 17                       | Application database                           |
| `database_url`         | Secrets Manager `orchex/DATABASE_URL`   | Shared Postgres URL for ECS tasks              |
| `sqs_node_jobs`        | SQS queue + DLQ + policies              | Node-job queue (API produces, worker consumes) |
| `function_sandbox`     | Lambda `orchex-function-sandbox`        | Shared JS sandbox; worker Invoke only          |
| `ecs_db_migrate`       | ECS task definition only                | One-shot `aws ecs run-task` migrations         |
| `ecs_builder_api`      | ECS service + task definition           | Long-running builder API behind ALB            |
| `ecs_execution_api`    | ECS service + task definition           | Long-running execution API behind ALB          |
| `ecs_execution_worker` | ECS service + task definition           | Internal worker (no ALB)                       |

Root-level `data.aws_secretsmanager_secret_version.rds_master` reads the RDS-managed master secret so Terraform can compose `DATABASE_URL` (not injected into ECS tasks directly).

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) matching `required_version` in `terraform.tf`
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) v2
- [Docker](https://docs.docker.com/get-docker/)
- [`jq`](https://jqlang.github.io/jq/) (optional, for reading JSON outputs)

### AWS CLI login (required)

You **must** be logged in to AWS via the CLI before running `terraform plan`, `terraform apply`, or any `aws` commands. Terraform uses the profile named in `providers.tf` (default: `orchex`).

**SSO (recommended):**

```bash
aws sso login --profile orchex
```

**Named profile with access keys:**

```bash
export AWS_PROFILE=orchex
# or: aws configure --profile orchex
```

Verify credentials:

```bash
aws sts get-caller-identity --profile orchex
```

Set the AWS profile name in `providers.tf` to match your local setup. Do not commit access keys or secrets.

Default region is `ap-south-1` (`var.aws_region`). Override when needed:

```bash
terraform apply -var='aws_region=YOUR_REGION'
```

Default environment is `production` (`var.environment`). Override when needed:

```bash
terraform apply -var='environment=staging'
```

## Tagging

All resources receive provider default tags plus module-level tags:

| Tag           | Source            | Example values                                                                                |
| ------------- | ----------------- | --------------------------------------------------------------------------------------------- |
| `Project`     | provider default  | `orchex`                                                                                      |
| `Environment` | `var.environment` | `production`                                                                                  |
| `Service`     | per module        | `builder-api`, `execution-api`, `shared`                                                      |
| `Component`   | per module        | `ecr`, `alb`, `ecs`, `ecs-service`, `ecs-run-task`, `sqs`, `lambda`, `rds`, `secrets-manager` |
| `Name`        | per resource      | `orchex-postgres`, `orchex-builder-api`                                                       |

App-specific resources use `Service = builder-api` or `Service = execution-api`. The worker is tagged `Service = execution-api`. Shared infrastructure (ALB, ECS cluster, RDS, SQS, migrate ECR/task, app secret) uses `Service = shared`.

## Infrastructure architecture

Root `main.tf` composes the modules above. Public traffic flows from the ALB to Fargate **API** tasks. The **worker** is internal (no ALB). Tasks reach RDS over the VPC using a shared **`data_plane_client`** security group. Execution-api and the worker talk to SQS with the task role (no static AWS keys). The worker may `Invoke` the shared Function sandbox Lambda; the Lambda is not in a VPC (outbound internet). Database credentials reach tasks via the shared **`orchex/DATABASE_URL`** secret (not the RDS master secret).

### Runtime architecture

```mermaid
flowchart TB
  User([Internet])

  subgraph edge["modules/alb"]
    ALB["Application Load Balancer<br/>HTTP :80 · path rules · default 404"]
  end

  subgraph ecr["modules/ecr"]
    ECRAPI["orchex-builder-api:latest"]
    ECREX["orchex-execution-api:latest"]
    ECRW["orchex-execution-worker:latest"]
    ECRMIG["orchex-db-migrate:latest"]
  end

  subgraph ecs["modules/ecs — orchex-cluster"]
    SG["data_plane_client SG"]
    API["modules/ecs_service<br/>orchex-builder-api<br/>ALB · :8080"]
    EX["modules/ecs_service<br/>orchex-execution-api<br/>ALB · :8080"]
    WRK["modules/ecs_service<br/>orchex-execution-worker<br/>internal · no ALB"]
    MIG["modules/ecs_run_task<br/>orchex-db-migrate<br/>one-shot · aws ecs run-task"]
  end

  subgraph sqs["modules/sqs"]
    Q["orchex-node-jobs"]
    DLQ["orchex-node-jobs-dlq"]
  end

  subgraph lam["modules/lambda_sandbox"]
    SB["orchex-function-sandbox<br/>zip · nodejs24.x"]
  end

  subgraph sm["Secrets Manager"]
    APPURL["orchex/DATABASE_URL"]
  end

  subgraph rdsmod["modules/rds"]
    RDS[("orchex-postgres<br/>PostgreSQL 17 · private")]
  end

  User -->|HTTP| ALB
  ALB -->|"/health/builder · /v1/workflows*"| API
  ALB -->|"/health/execution · /v1/runs*"| EX
  ECRAPI -.->|pull image| API
  ECREX -.->|pull image| EX
  ECRW -.->|pull image| WRK
  ECRMIG -.->|pull image| MIG
  APPURL -->|DATABASE_URL| API
  APPURL -->|DATABASE_URL| EX
  APPURL -->|DATABASE_URL| WRK
  APPURL -->|GOOSE_DBSTRING| MIG
  API -->|:5432| RDS
  EX -->|:5432| RDS
  WRK -->|:5432| RDS
  MIG -->|:5432| RDS
  EX -->|SendMessage| Q
  WRK -->|Receive / Delete| Q
  WRK -->|Invoke| SB
  Q -->|after 5 receives| DLQ
  SG -.- API
  SG -.- EX
  SG -.- WRK
  SG -.- MIG
```

Solid arrows are request, queue, or database traffic. Dotted arrows are image pulls or security-group attachment (API, worker, and migrate tasks use **`data_plane_client`** for RDS access). The worker is not registered with the ALB.

### Secrets and bootstrap

Terraform composes the app connection string from the RDS-managed master secret at apply time. ECS tasks never read the master secret at runtime.

```mermaid
flowchart LR
  subgraph tf["Terraform apply"]
    DS["data.aws_secretsmanager_secret_version<br/>rds_master"]
    LOC["locals.database_url<br/>postgres://…?sslmode=require"]
    MOD["module.database_url<br/>modules/secrets_manager"]
  end

  RDS[("module.rds")]
  MASTER["RDS master secret<br/>AWS-managed JSON"]
  APPURL["orchex/DATABASE_URL"]

  RDS -->|manage_master_user_password| MASTER
  MASTER --> DS
  DS --> LOC
  LOC --> MOD
  MOD --> APPURL
  APPURL -->|task_exec_secret_arns| API["ecs_builder_api"]
  APPURL -->|task_exec_secret_arns| EX["ecs_execution_api"]
  APPURL -->|task_exec_secret_arns| WRK["ecs_execution_worker"]
  APPURL -->|task_exec_secret_arns| MIG["ecs_db_migrate"]
```

### Release flow (operations)

```mermaid
flowchart LR
  A["terraform apply<br/>(incl. zip sandbox Lambda)"] --> B["docker push<br/>orchex-db-migrate"]
  B --> C["aws ecs run-task<br/>goose up"]
  C --> D["docker push<br/>orchex-builder-api"]
  D --> E["docker push<br/>orchex-execution-api"]
  E --> F["docker push<br/>orchex-execution-worker"]
  F --> G["ECS force-new-deployment"]
  G --> H["curl ALB /health/builder<br/>curl ALB /health/execution"]
```

See [Lifecycle](#lifecycle) for commands.

### Configuration and secrets

| Secret / env                                           | Set by                            | Consumed by                                                                    |
| ------------------------------------------------------ | --------------------------------- | ------------------------------------------------------------------------------ |
| RDS master JSON (`username`, `password`, …)            | RDS `manage_master_user_password` | Terraform data source only                                                     |
| `orchex/DATABASE_URL` (`postgres://…?sslmode=require`) | `module.database_url`             | ECS APIs, worker (`DATABASE_URL`), migrate task (`GOOSE_DBSTRING`)             |
| `HTTP_ADDR=:8080`                                      | Task definition env               | Builder, execution API, and worker containers                                  |
| `SQS_QUEUE_URL`                                        | Task definition env               | Execution API (relay) and worker                                               |
| `SQS_DLQ_URL`                                          | Task definition env               | Execution API                                                                  |
| `AWS_REGION`                                           | Task definition env               | Execution API and worker (real SQS; do **not** set `AWS_ENDPOINT_URL` in prod) |
| `FUNCTION_SANDBOX_ARN`                                 | Task definition env               | Worker (sync `Invoke` of `orchex-function-sandbox`)                            |

ECS **task execution roles** get `secretsmanager:GetSecretValue` on `orchex/DATABASE_URL` via `task_exec_secret_arns`. Tasks do **not** read the RDS master secret at runtime.

### `modules/ecr`

- Wraps [terraform-aws-modules/ecr/aws](https://registry.terraform.io/modules/terraform-aws-modules/ecr/aws) v3.2.0.
- Four private repositories in root `main.tf`:
  - **`orchex-builder-api`** — `Service = builder-api`
  - **`orchex-execution-api`** — `Service = execution-api`
  - **`orchex-execution-worker`** — `Service = execution-api`
  - **`orchex-db-migrate`** — `Service = shared`
- **`repository_force_delete = true`** — full `terraform destroy` deletes repos even when images exist (see [Destroy infrastructure](#destroy-infrastructure)).
- Lifecycle policy keeps only the newest tagged image and expires untagged layers after one day.

### `modules/alb`

- Wraps [terraform-aws-modules/alb/aws](https://registry.terraform.io/modules/terraform-aws-modules/alb/aws).
- Creates an internet-facing ALB in the account **default VPC** and its public subnets.
- Listener on **port 80**. Default action is a **fixed 404** JSON body (`{"error":"not found"}`) — unmatched paths never hit a target group.
- Listener **rules** (path patterns; lower `priority` is evaluated first):
  - **priority 10** `builder` — `/health/builder`, `/health/builder/*`, `/v1/workflows`, `/v1/workflows/*`
  - **priority 20** `execution` — `/health/execution`, `/health/execution/*`, `/v1/runs`, `/v1/runs/*`
- Two **IP** target groups on **port 8080** (Fargate `awsvpc`); ECS registers task IPs—`create_attachment = false` in Terraform.
  - **builder** health check: `/health/builder`
  - **execution** health check: `/health/execution`
- ALB security group allows inbound **80** from `0.0.0.0/0`.

### `modules/ecs`

- Wraps [terraform-aws-modules/ecs/aws//modules/cluster](https://registry.terraform.io/modules/terraform-aws-modules/ecs/aws).
- Creates a shared **Fargate** cluster (`orchex-cluster`).
- Creates a shared **`data_plane_client`** security group (`orchex-cluster-data-plane-client`). Every ECS service attaches this SG to its task ENI so RDS can allow Postgres access by security group identity.
- Container Insights is disabled for early-stage cost; enable later if needed.

### `modules/ecs_service`

- Wraps [terraform-aws-modules/ecs/aws//modules/service](https://registry.terraform.io/modules/terraform-aws-modules/ecs/aws) v7.5.0.
- Runs Fargate **services** on the shared cluster (`desired_count = 1`): `orchex-builder-api`, `orchex-execution-api`, and `orchex-execution-worker`.
- Pulls images from **ECR** (`…:latest`).
- **Secrets**: injects `DATABASE_URL` from `var.database_url_secret_arn`; grants the task execution role access via `task_exec_secret_arns`.
- **Environment**: `HTTP_ADDR=:8080` (each task has its own ENI). Execution-api and worker also get `SQS_QUEUE_URL` and `AWS_REGION` (API also `SQS_DLQ_URL`; worker also `FUNCTION_SANDBOX_ARN`).
- **Load balancer**: `attach_load_balancer` must be a **literal** `true`/`false` (not derived from a computed ALB ID). Terraform cannot plan `for_each` on security-group ingress rules if the map itself is unknown until apply.
  - APIs: `attach_load_balancer = true` plus `target_group_arn` and `alb_security_group_id`.
  - Worker: leave `attach_load_balancer = false` (default); no target group, no ALB ingress.
- **Networking**: tasks get a public IP in default VPC subnets. Each task ENI has two security groups:
  - **Service SG** — for APIs, ingress on `:8080` only from the **ALB security group**; for the worker, no ALB ingress
  - **`data_plane_client` SG** — required client identity for RDS access (`data_plane_client_security_group_id` from `module.ecs`).
- Execution-api task role name is pinned `orchex-execution-api`; worker `orchex-execution-worker` (`tasks_iam_role_use_name_prefix = false`) so SQS queue policies and the Lambda resource policy can allow those ARNs without a Terraform cycle. The worker task role also gets `lambda:InvokeFunction` on the sandbox ARN.
- Tagged with `Service = builder-api` or `Service = execution-api` and `Component = ecs-service`.
- Container runs with a read-only root filesystem (distroless image).

### `modules/sqs`

- Wraps [terraform-aws-modules/sqs/aws](https://registry.terraform.io/modules/terraform-aws-modules/sqs/aws) v5.2.2.
- Standard queue **`orchex-node-jobs`** plus DLQ; 14-day retention; `maxReceiveCount = 5`; default visibility 60s; long poll 20s.
- Role ARNs are built from account ID + pinned role **names** (not from ECS module outputs) to avoid a cycle with queue policies.
- **Queue resource policy** (SQS allows at most 7 actions per statement; `*Batch` action names are **invalid** on the queue itself — IAM still gets batch APIs):
  - Producer (`orchex-execution-api`): send/receive/delete/visibility/get on the source queue **and** the DLQ
  - Consumer (`orchex-execution-worker`): receive/delete/visibility/get on the **source queue only** (no `SendMessage`, no DLQ)
- IAM task-role statements use the module outputs `producer_queue_actions` / `consumer_queue_actions` (includes `*Batch`).

### `modules/lambda_sandbox`

- Wraps [terraform-aws-modules/lambda/aws](https://registry.terraform.io/modules/terraform-aws-modules/lambda/aws) v8.1.2.
- One zip function **`orchex-function-sandbox`**: `nodejs24.x`, handler `index.handler`, source [`modules/lambda_sandbox/src`](modules/lambda_sandbox/src). Terraform zips that directory on apply (no manual zip, no ECR, **`store_on_s3` off**). AWS stores the deployment package in Lambda-managed storage — not an S3 bucket in this account. Local zip artifacts land in `infra/builds/` (gitignored; module default `artifacts_dir`).
- **Invoke contract (sync):** workers call `Invoke` with `{ source, input, timeout_ms }`. User Function `source` is data (Postgres at run time), not files in the zip. The zip is only our sandbox. `index.js` validates **runtime `input`** (`{ data: object }` only); it does not re-check draft config (`source` / `timeout_ms`).
- **Timeout 300s** (Function node `timeout_ms` max is 300000). **128 MB**. **x86_64**. **`publish = false`** (no versions; invoke `$LATEST`).
- **Not in a VPC** — the sandbox can reach the public internet.
- CloudWatch Logs retention **7 days**. Lambda execution role is **logs only**.
- **Invoke ACL** (same dual-IAM pattern as SQS): resource policy allows only the pinned **`orchex-execution-worker`** role; that role’s identity policy has `lambda:InvokeFunction` on this function. Execution-api cannot invoke. Same-account IAM admins with `lambda:InvokeFunction` on `*` can still invoke; that is not gated by this resource policy.
- `recursive_loop = Terminate`. No Function URL, layers, EFS, reserved/provisioned concurrency, or async event config / Lambda DLQ.
- Tagged `Service = execution-api`, `Component = lambda`.

### `modules/ecs_run_task`

- Wraps the same ECS service module with **`create_service = false`** — task definition only, no long-running service.
- Used for **`orchex-db-migrate`**: goose migrations via `aws ecs run-task`.
- **`create_service = false`** — no `desired_count`; the task starts on demand and exits when goose finishes.
- **Secrets**: maps `orchex/DATABASE_URL` → container env **`GOOSE_DBSTRING`**.
- **Environment**: `GOOSE_DRIVER=postgres`, `GOOSE_MIGRATION_DIR=/migrations`.
- Same **`data_plane_client` SG** and default VPC subnets as the API services (public IP for egress).
- Outputs **`run_task_network_configuration`** (subnets, security groups) for CLI `run-task` commands.
- Tagged with `Component = ecs-run-task`.

### `modules/secrets_manager`

- Wraps [terraform-aws-modules/secrets-manager/aws](https://registry.terraform.io/modules/terraform-aws-modules/secrets-manager/aws) v2.1.0.
- Generic wrapper: creates a Secrets Manager secret + version from `var.secret_string`.
- Root **`module.database_url`** stores the composed Postgres URL at **`orchex/DATABASE_URL`** (`Service = shared`).
- **`recovery_window_in_days`** (default **0**) — destroy removes the secret immediately so apply can reuse the same name. Override on the module if you want a recovery window.
- Secret value is built in root `main.tf` from the RDS master secret + RDS endpoint metadata + `sslmode=require`.
- **Note:** the composed URL exists in Terraform state; re-apply after RDS password rotation to refresh it.

### `modules/rds`

- Wraps [terraform-aws-modules/rds/aws](https://registry.terraform.io/modules/terraform-aws-modules/rds/aws) v7.2.1.
- **PostgreSQL 17** on `db.t4g.medium` (2 vCPU, 4 GiB RAM).
- **20 GiB** `gp3` storage (RDS minimum for gp3 Postgres).
- **Single-AZ**, no automated backups (`backup_retention_period = 0`).
- **`publicly_accessible = false`** — no public IP; not reachable from the internet.
- Master password managed by AWS (**Secrets Manager** via `manage_master_user_password`).
- Default database / user: `orchex` / `orchex`.
- **DB subnet group** and **RDS security group** created as standalone resources; the RDS module uses `create_db_subnet_group = false` and `vpc_security_group_ids` pointing at the module SG.
- **RDS security group** allows inbound PostgreSQL (`5432`) only from the shared **`data_plane_client`** security group.
- **Outputs** expose connection metadata only (endpoint, address, port, engine, etc.) plus **`db_instance_master_user_secret_arn`** (sensitive, for Terraform composition only).

Wiring in root `main.tf`:

```hcl
module "ecr_builder_api"        { repository_name = "orchex-builder-api" ... }
module "ecr_execution_api"      { repository_name = "orchex-execution-api" ... }
module "ecr_execution_worker"   { repository_name = "orchex-execution-worker" ... }
module "ecr_db_migrate"         { repository_name = "orchex-db-migrate" ... }

module "sqs_node_jobs" {
  producer_task_role_name = "orchex-execution-api"
  consumer_task_role_name = "orchex-execution-worker"
}

module "function_sandbox" {
  invoker_task_role_name = "orchex-execution-worker"
}

module "ecs_builder_api" {
  attach_load_balancer = true
  target_group_arn     = module.alb.target_groups["builder"].arn
  alb_security_group_id = module.alb.security_group_id
}

module "ecs_execution_api" {
  attach_load_balancer = true
  extra_environment    = [SQS_QUEUE_URL, SQS_DLQ_URL, AWS_REGION]
}

module "ecs_execution_worker" {
  # attach_load_balancer defaults to false
  extra_environment = [SQS_QUEUE_URL, AWS_REGION, FUNCTION_SANDBOX_ARN]
}
```

New ECS services that need database access should:

1. Pass `data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id`.
2. Pass `database_url_secret_arn = module.database_url.arn` (shared secret) or a dedicated secret module instance.
3. Map the secret to `DATABASE_URL` (or app-specific env names) and set `task_exec_secret_arns`.
4. Set `attach_load_balancer = true` with ALB ids for public APIs, or leave it `false` for internal workers.

For one-shot jobs (migrations, batch work), use **`modules/ecs_run_task`** instead of **`modules/ecs_service`**.

### Container images (built outside Terraform)

| Dockerfile                                                      | ECR repository                   | ECS consumer                            |
| --------------------------------------------------------------- | -------------------------------- | --------------------------------------- |
| [`docker/Dockerfile`](../docker/Dockerfile)                     | `orchex-builder-api:latest`      | `module.ecs_builder_api` (service)      |
| [`docker/Dockerfile.execution`](../docker/Dockerfile.execution) | `orchex-execution-api:latest`    | `module.ecs_execution_api` (service)    |
| [`docker/Dockerfile.worker`](../docker/Dockerfile.worker)       | `orchex-execution-worker:latest` | `module.ecs_execution_worker` (service) |
| [`docker/Dockerfile.migrate`](../docker/Dockerfile.migrate)     | `orchex-db-migrate:latest`       | `module.ecs_db_migrate` (`run-task`)    |

Production images pin **`linux/amd64`** and use **distroless** runtimes. The Function sandbox is **not** a container — Terraform zips [`modules/lambda_sandbox/src`](modules/lambda_sandbox/src) during apply. Local development uses [`docker/Dockerfile.local`](../docker/Dockerfile.local), [`docker/Dockerfile.execution.local`](../docker/Dockerfile.execution.local), [`docker/Dockerfile.worker.local`](../docker/Dockerfile.worker.local), ElasticMQ, and Compose ([`docker-compose.yml`](../docker-compose.yml)) — not deployed by this Terraform stack. Compose workers skip the sandbox (`FUNCTION_SANDBOX_ARN` unset and/or `AWS_ENDPOINT_URL` set).

After `terraform apply`, the ALB DNS name is available via `terraform output -json alb`. RDS connection details are under `terraform output -json rds`.

## Lifecycle

End-to-end flow for this stack:

| Step | Action                                           | Section                                                     |
| ---- | ------------------------------------------------ | ----------------------------------------------------------- |
| 1    | Create / update AWS resources                    | [Create infrastructure](#create-infrastructure)             |
| 2    | Build & push migrate image, run goose on RDS     | [Run database migrations](#run-database-migrations)         |
| 3    | Build & push builder image, deploy ECS service   | [Deploy the builder API](#deploy-the-builder-api)           |
| 4    | Build & push execution image, deploy ECS service | [Deploy the execution API](#deploy-the-execution-api)       |
| 5    | Build & push worker image, deploy ECS service    | [Deploy the execution worker](#deploy-the-execution-worker) |
| 6    | Tear down (optional)                             | [Destroy infrastructure](#destroy-infrastructure)           |

Terraform resolves dependency order on **create**. You only need a strict manual order for **migrations** (step 2 before relying on DB-backed API routes) and for **partial destroy** (see below).

Long `terraform apply` / `destroy` runs can exceed SSO session length. Re-authenticate if you see `ExpiredToken`:

```bash
aws sso login --profile orchex
aws sts get-caller-identity --profile orchex
```

## Create infrastructure

Provisions ECR repos, ECS cluster, RDS, SQS, the Function sandbox Lambda, Secrets Manager (`orchex/DATABASE_URL`), ALB, builder/execution API services, the internal worker, and the migrate task definition.

```bash
cd infra

terraform init
terraform fmt -recursive
terraform plan
terraform apply
```

**Incremental apply** (optional — Terraform handles dependencies):

| Phase | What                                          | `-target` (optional)                                                                                                                                 |
| ----- | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| A     | Cluster + RDS + app secret + migrate task def | `module.ecs`, `module.rds`, `module.database_url`, `module.ecs_db_migrate`, `module.ecr_db_migrate`                                                  |
| B     | Migrations                                    | [Run database migrations](#run-database-migrations) (not Terraform)                                                                                  |
| C     | ALB + SQS + Lambda + API + worker services    | `module.alb`, `module.sqs_node_jobs`, `module.function_sandbox`, `module.ecs_builder_api`, `module.ecs_execution_api`, `module.ecs_execution_worker` |

For most cases, a single **`terraform apply`** is enough.

Inspect outputs:

```bash
terraform output
terraform output -json ecr_builder_api
terraform output -json ecr_execution_api
terraform output -json ecr_execution_worker
terraform output -json sqs_node_jobs
terraform output -json function_sandbox
terraform output -json alb | jq -r '.dns_name'
terraform output -json rds | jq
```

| Output                 | Meaning                                                         |
| ---------------------- | --------------------------------------------------------------- |
| `ecr_builder_api`      | ECR repository URL, name, ARN                                   |
| `ecr_execution_api`    | Execution API ECR repository URL, name, ARN                     |
| `ecr_execution_worker` | Execution worker ECR repository URL, name, ARN                  |
| `ecr_db_migrate`       | Shared goose migration image repository                         |
| `sqs_node_jobs`        | Queue / DLQ URLs and ARNs, producer and consumer role ARNs      |
| `function_sandbox`     | Sandbox Lambda name / ARN, log group, invoker role ARN          |
| `alb`                  | ALB DNS name, target groups, listener rules, security groups    |
| `ecs`                  | Shared Fargate cluster ARN / name, `data_plane_client` SG       |
| `ecs_db_migrate`       | Migrate task definition + `run_task_network_configuration`      |
| `ecs_builder_api`      | Builder service task definition, security group, etc.           |
| `ecs_execution_api`    | Execution API service task definition, security group, etc.     |
| `ecs_execution_worker` | Worker service task definition, security group, etc.            |
| `database_url_secret`  | Shared `orchex/DATABASE_URL` secret ARN / name                  |
| `rds`                  | RDS connection metadata only (see fields below; no credentials) |

**`rds` output fields** (non-sensitive):

| Field                               | Meaning                             |
| ----------------------------------- | ----------------------------------- |
| `db_instance_endpoint`              | Host:port connection string         |
| `db_instance_address`               | Hostname                            |
| `db_instance_port`                  | Port (default `5432`)               |
| `db_instance_name`                  | Database name (`orchex`)            |
| `db_instance_identifier`            | RDS instance id (`orchex-postgres`) |
| `db_instance_arn`                   | Instance ARN                        |
| `db_instance_status`                | e.g. `available`                    |
| `db_instance_engine`                | `postgres`                          |
| `db_instance_engine_version_actual` | Running Postgres version            |
| `db_subnet_group_id`                | DB subnet group name                |
| `security_group_id`                 | RDS security group id               |

Module-level `modules/rds` outputs match the above plus `db_subnet_group_arn` and `security_group_arn`. Username, password, and Secrets Manager ARNs are intentionally omitted.

### RDS connection (after apply)

- **Host**: `terraform output -json rds | jq -r '.db_instance_address'`
- **Port**: `terraform output -json rds | jq -r '.db_instance_port'`
- **Database**: `terraform output -json rds | jq -r '.db_instance_name'`
- **Username / password**: not in Terraform outputs. RDS stores the master password in **Secrets Manager**. Retrieve via the AWS console (RDS → `orchex-postgres` → Configuration) or CLI:

```bash
# List the secret attached to the instance, then fetch the password JSON
aws rds describe-db-instances \
  --db-instance-identifier orchex-postgres \
  --query 'DBInstances[0].MasterUserSecret.SecretArn' \
  --output text \
  --region ap-south-1 \
  --profile orchex

# Then (replace SECRET_ARN):
aws secretsmanager get-secret-value \
  --secret-id SECRET_ARN \
  --query SecretString \
  --output text \
  --region ap-south-1 \
  --profile orchex | jq -r '.password'
```

Only ECS tasks with the `data_plane_client` security group attached can reach the database. The instance is not publicly accessible.

## Run database migrations

After **RDS** and the shared **`orchex/DATABASE_URL`** secret exist (`terraform apply`), apply schema changes with a **one-shot Fargate task** (not a long-running service). The task runs `goose up` and exits.

Use the same region as `var.aws_region`. In **zsh**, always write `${REPO_URL}:latest` (not `$REPO_URL:latest`).

### Build and push the migrate image

```bash
cd infra

MIGRATE_REPO=$(terraform output -json ecr_db_migrate | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region
CLUSTER=$(terraform output -json ecs | jq -r '.name')

aws sso login --profile orchex   # if needed

aws ecr get-login-password --region "$REGION" --profile orchex \
  | docker login --username AWS --password-stdin "$(echo "$MIGRATE_REPO" | cut -d/ -f1)"

cd ..
docker build -f docker/Dockerfile.migrate -t "${MIGRATE_REPO}:latest" .
docker push "${MIGRATE_REPO}:latest"
```

### Run migrations (`aws ecs run-task`)

Reads `GOOSE_DBSTRING` from the shared `orchex/DATABASE_URL` secret. Set `CLUSTER` if you are not reusing the shell from above:

```bash
cd infra

REGION=ap-south-1
CLUSTER=$(terraform output -json ecs | jq -r '.name')
TASK_DEF=$(terraform output -json ecs_db_migrate | jq -r '.task_definition_family')
NET=$(terraform output -json ecs_db_migrate | jq -r '.run_task_network_configuration')
SUBNETS=$(echo "$NET" | jq -r '.subnets')
SGS=$(echo "$NET" | jq -r '.security_groups')

TASK_ARN=$(aws ecs run-task \
  --cluster "$CLUSTER" \
  --task-definition "$TASK_DEF" \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNETS],securityGroups=[$SGS],assignPublicIp=ENABLED}" \
  --region "$REGION" \
  --profile orchex \
  --query 'tasks[0].taskArn' \
  --output text)

aws ecs wait tasks-stopped --cluster "$CLUSTER" --tasks "$TASK_ARN" --region "$REGION" --profile orchex

aws ecs describe-tasks \
  --cluster "$CLUSTER" \
  --tasks "$TASK_ARN" \
  --region "$REGION" \
  --profile orchex \
  --query 'tasks[0].containers[0].{exitCode:exitCode,reason:reason}' \
  --output json
```

Expect `"exitCode": 0`. Re-run after pushing a new migrate image whenever `db/migrations/` changes.

On a **fresh database**, RDS logs may show `relation "goose_db_version" does not exist` once while goose bootstraps its version table — that is normal if the task still exits `0`.

## Deploy the builder API

Use the same region as `var.aws_region`. In **zsh**, always write `${REPO_URL}:tag` (not `$REPO_URL:tag`).

```bash
cd infra

REPO_URL=$(terraform output -json ecr_builder_api | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region

aws sso login --profile orchex   # if needed

# Username is always "AWS". Password is a short-lived token from the CLI.
aws ecr get-login-password --region "$REGION" --profile orchex \
  | docker login --username AWS --password-stdin "$(echo "$REPO_URL" | cut -d/ -f1)"

# Build from the repository root, tagged for ECR (amd64 is pinned in the Dockerfile)
cd ..
docker build -f docker/Dockerfile -t "${REPO_URL}:latest" .

docker push "${REPO_URL}:latest"
```

Confirm the image landed in ECR:

```bash
aws ecr describe-images \
  --repository-name "$(cd infra && terraform output -json ecr_builder_api | jq -r '.repository_name')" \
  --region "$REGION" \
  --profile orchex
```

ECS picks up a new deployment when the task definition image changes or when you force a new deployment after pushing `:latest`:

```bash
aws ecs update-service \
  --cluster orchex-cluster \
  --service orchex-builder-api \
  --force-new-deployment \
  --region ap-south-1 \
  --profile orchex
```

Hit the builder API via the ALB:

```bash
curl "$(cd infra && terraform output -json alb | jq -r '.dns_name')/health/builder"
```

## Deploy the execution API

Same region / zsh quoting as builder. Use **`docker/Dockerfile.execution`** (`docker/Dockerfile` is builder-api).

```bash
cd infra

REPO_URL=$(terraform output -json ecr_execution_api | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region

aws sso login --profile orchex   # if needed

aws ecr get-login-password --region "$REGION" --profile orchex \
  | docker login --username AWS --password-stdin "$(echo "$REPO_URL" | cut -d/ -f1)"

cd ..
docker build -f docker/Dockerfile.execution -t "${REPO_URL}:latest" .

docker push "${REPO_URL}:latest"
```

Confirm the image landed in ECR:

```bash
aws ecr describe-images \
  --repository-name "$(cd infra && terraform output -json ecr_execution_api | jq -r '.repository_name')" \
  --region "$REGION" \
  --profile orchex
```

Force a new deployment after pushing `:latest`:

```bash
aws ecs update-service \
  --cluster orchex-cluster \
  --service orchex-execution-api \
  --force-new-deployment \
  --region ap-south-1 \
  --profile orchex
```

Hit execution via the ALB path rule:

```bash
curl "$(cd infra && terraform output -json alb | jq -r '.dns_name')/health/execution"
```

Unmatched paths (for example `/`) return the listener default `404` `{"error":"not found"}` without reaching ECS.

## Deploy the execution worker

Same region / zsh quoting as the APIs. Use **`docker/Dockerfile.worker`**. The worker is **not** on the ALB; `/health/worker` is only on the task itself.

```bash
cd infra

REPO_URL=$(terraform output -json ecr_execution_worker | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region

aws sso login --profile orchex   # if needed

aws ecr get-login-password --region "$REGION" --profile orchex \
  | docker login --username AWS --password-stdin "$(echo "$REPO_URL" | cut -d/ -f1)"

cd ..
docker build -f docker/Dockerfile.worker -t "${REPO_URL}:latest" .

docker push "${REPO_URL}:latest"
```

Confirm the image landed in ECR:

```bash
aws ecr describe-images \
  --repository-name "$(cd infra && terraform output -json ecr_execution_worker | jq -r '.repository_name')" \
  --region "$REGION" \
  --profile orchex
```

Force a new deployment after pushing `:latest`:

```bash
aws ecs update-service \
  --cluster orchex-cluster \
  --service orchex-execution-worker \
  --force-new-deployment \
  --region ap-south-1 \
  --profile orchex
```

The worker is not on the ALB. After apply, `FUNCTION_SANDBOX_ARN` is on the task. The worker process **connects to RDS first**; if Postgres is unreachable it never Invokes. With the current worker image, startup then sync-Invokes the sandbox once (`internal/sandbox` ping: `return { ping: true }` and `input: { data: {} }`). Watch the task’s CloudWatch logs:

- `sandbox: invoke ok payload=...` — task role can Invoke and the handler ran
- `sandbox: invoke failed: ...` — IAM, timeout, or runtime error (`AccessDeniedException` is the worker-ACL case)
- `sandbox: skip invoke (local)` — `FUNCTION_SANDBOX_ARN` empty or `AWS_ENDPOINT_URL` set (Compose)

Push **`docker/Dockerfile.worker`** after the ping code landed, then force a new deployment; an old `:latest` image will not log those lines.

Optional: confirm the zip/handler without the worker (your IAM user, not the task role):

```bash
FN=$(cd infra && terraform output -json function_sandbox | jq -r '.function_name')

aws lambda invoke \
  --function-name "$FN" \
  --cli-binary-format raw-in-base64-out \
  --payload '{"source":"return { ping: true };","input":{"data":{}},"timeout_ms":5000}' \
  --region ap-south-1 \
  --profile orchex \
  /tmp/orchex-sandbox-out.json

cat /tmp/orchex-sandbox-out.json
```

To exercise SQS in AWS, send a message then check worker CloudWatch logs:

```bash
QUEUE_URL=$(cd infra && terraform output -json sqs_node_jobs | jq -r '.queue_url')

aws sqs send-message \
  --queue-url "$QUEUE_URL" \
  --message-body '{"run_id":"smoke","node_id":"smoke","attempt":1,"source":"cli"}' \
  --region ap-south-1 \
  --profile orchex
```

## Destroy infrastructure

**Warning:** Destroying **RDS** deletes the database (`skip_final_snapshot = true`, no automated backups). Export anything you need before destroy.

Refresh AWS credentials before a long destroy (same as apply):

```bash
aws sso login --profile orchex
```

### Full destroy (everything, including ECR)

Removes **all** Terraform-managed resources, including ECR repositories (`orchex-builder-api`, `orchex-execution-api`, `orchex-execution-worker`, `orchex-db-migrate`) and their images. Repositories use `repository_force_delete = true`, so non-empty repos are deleted too.

```bash
cd infra

terraform plan -destroy
terraform destroy
```

After destroy, `terraform.tfstate` no longer tracks those resources. Run `terraform apply` to recreate from scratch.

### Partial destroy (keep ECR repositories and images)

Use this when you want to tear down **compute, networking, database, and secrets**, but **keep** the ECR repos and the images already pushed (`:latest` tags survive until lifecycle rules expire old tags).

**Destroyed:** ALB, ECS cluster, builder / execution / worker services, migrate task definition, SQS queues, Function sandbox Lambda, RDS, `orchex/DATABASE_URL` secret, related IAM roles and security groups.

**Kept in AWS:** `module.ecr_builder_api`, `module.ecr_execution_api`, `module.ecr_execution_worker`, `module.ecr_db_migrate` (repos + images).

```bash
cd infra

terraform destroy \
  -target=module.ecs_builder_api \
  -target=module.ecs_execution_api \
  -target=module.ecs_execution_worker \
  -target=module.ecs_db_migrate \
  -target=module.sqs_node_jobs \
  -target=module.function_sandbox \
  -target=module.database_url \
  -target=module.rds \
  -target=module.alb \
  -target=module.ecs
```

Review the plan carefully — `-target` limits what Terraform destroys, but **dependencies of those modules** are still removed. ECR modules are **not** in the target list, so they stay.

To recreate the stack later (reusing existing ECR images):

```bash
terraform apply
# then: run migrations (RDS will be empty) → deploy / force-new-deployment
```

**Note:** Partial destroy leaves ECR in state. A later full `terraform destroy` (without `-target`) will still delete ECR unless you remove those modules from the config or state first.

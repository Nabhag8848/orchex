# Orchex infrastructure

Terraform for AWS resources. Provisions **ECR**, an **Application Load Balancer (ALB)**, a shared **ECS Fargate cluster**, an **ECS service** for the workflow builder API (`orchex-builder-api`), a **one-shot migrate task definition** (`orchex-db-migrate`), and **RDS PostgreSQL** (`orchex-postgres`).

Terraform creates and wires the infrastructure. Building and pushing the container image is done with Docker + the AWS CLI.

## Layout

```text
infra/
  terraform.tf      # required Terraform / provider versions
  providers.tf      # AWS provider (region, profile, default tags)
  variables.tf      # shared variables (region, environment)
  main.tf           # root modules (ECR, ALB, ECS, ECS service, ECS migrate task, RDS, secrets)
  outputs.tf        # root outputs
  modules/
    ecr/            # ECR repository for container images
    alb/            # internet-facing ALB + target group
    ecs/            # shared Fargate cluster + data_plane_client SG
    ecs_service/    # Fargate service (task definition, SG, ALB attachment)
    ecs_run_task/   # one-shot Fargate task definition (migrations via run-task)
    secrets_manager/# Secrets Manager secret + version
    rds/            # PostgreSQL RDS (subnet group, SG, instance)
```

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

| Tag           | Source            | Example values                                                               |
| ------------- | ----------------- | ---------------------------------------------------------------------------- |
| `Project`     | provider default  | `orchex`                                                                     |
| `Environment` | `var.environment` | `production`                                                                 |
| `Service`     | per module        | `builder-api`, `shared`                                                      |
| `Component`   | per module        | `ecr`, `alb`, `ecs`, `ecs-service`, `ecs-run-task`, `rds`, `secrets-manager` |
| `Name`        | per resource      | `orchex-postgres`, `orchex-builder-api`                                      |

App-specific resources (ECR, ECS builder service) use `Service = builder-api`. Shared infrastructure (ALB, ECS cluster, RDS) uses `Service = shared`.

## Infrastructure architecture

Root `main.tf` composes five modules. Traffic flows from the ALB to Fargate tasks in the default VPC. ECS tasks reach RDS over the VPC using a shared client security group.

```text
Internet
   │
   ▼
┌─────────────────────────────────────┐
│  ALB (modules/alb)                  │
│  • default VPC, public subnets      │
│  • HTTP :80 → target group "builder"│
│  • health check: GET /health/builder│
└──────────────┬──────────────────────┘
               │ target group (IP mode, :8080)
               ▼
┌─────────────────────────────────────┐
│  ECS service (modules/ecs_service)  │
│  • Fargate task(s) in default VPC   │
│  • container :8080                  │
│  • service SG: ALB may reach :8080  │
│  • data_plane_client SG on task ENI │
│  • registers task IPs with ALB TG   │
└──────┬──────────────────┬───────────┘
       │                  │
       │ runs on          │ TCP :5432 (SG-to-SG)
       ▼                  ▼
┌──────────────────┐  ┌─────────────────────────────┐
│  ECS cluster     │  │  RDS (modules/rds)          │
│  (modules/ecs)   │  │  • PostgreSQL 17          │
│  • Fargate       │  │  • db.t4g.medium (2C/4GiB)  │
│  • data_plane_   │  │  • 20 GiB gp3 storage       │
│    client SG     │  │  • not publicly accessible  │
└──────────────────┘  │  • ingress: data_plane_     │
                      │    client SG only           │
                      └─────────────────────────────┘

Image sources: ECR → `orchex-builder-api:latest` (API), `orchex-db-migrate:latest` (goose migrations via `aws ecs run-task`)
```

### `modules/ecr`

- Wraps [terraform-aws-modules/ecr/aws](https://registry.terraform.io/modules/terraform-aws-modules/ecr/aws).
- Private repository for the builder API image (`orchex-builder-api`).
- Tagged with `Service = builder-api` (via `service` variable) and `Component = ecr`.

### `modules/alb`

- Wraps [terraform-aws-modules/alb/aws](https://registry.terraform.io/modules/terraform-aws-modules/alb/aws).
- Creates an internet-facing ALB in the account **default VPC** and its public subnets.
- Listener on **port 80** forwards all HTTP traffic to the `builder` target group.
- Target group uses **IP** mode on **port 8080** (Fargate `awsvpc`); ECS registers task IPs—`create_attachment = false` in Terraform.
- Health checks hit **`/health/builder`** and expect HTTP 200.
- ALB security group allows inbound **80** from `0.0.0.0/0`.

### `modules/ecs`

- Wraps [terraform-aws-modules/ecs/aws//modules/cluster](https://registry.terraform.io/modules/terraform-aws-modules/ecs/aws).
- Creates a shared **Fargate** cluster (`orchex-cluster`).
- Creates a shared **`data_plane_client`** security group (`orchex-cluster-data-plane-client`). Every ECS service attaches this SG to its task ENI so RDS can allow Postgres access by security group identity.
- Container Insights is disabled for early-stage cost; enable later if needed.

### `modules/ecs_service`

- Wraps [terraform-aws-modules/ecs/aws//modules/service](https://registry.terraform.io/modules/terraform-aws-modules/ecs/aws).
- Runs the builder API as a Fargate service (`orchex-builder-api`) on the shared cluster.
- Pulls the image from **ECR** (`orchex-builder-api:latest`).
- **Load balancer**: attaches the service to the ALB target group; ECS keeps registrations in sync as tasks start/stop.
- **Networking**: tasks get a public IP in default VPC subnets. Each task ENI has two security groups:
  - **Service SG** — ingress on `:8080` only from the **ALB security group**
  - **`data_plane_client` SG** — required client identity for RDS access (`data_plane_client_security_group_id` from `module.ecs`).
- Tagged with `Service = builder-api` (via `service` variable) and `Component = ecs-service`.
- Container runs with a read-only root filesystem (distroless image).

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
- **Outputs** expose connection metadata only (endpoint, address, port, engine, etc.). Username and password are **not** exported — credentials live in **Secrets Manager** (`manage_master_user_password`).

Wiring in root `main.tf`:

```hcl
module "ecr_builder_api" {
  source          = "./modules/ecr"
  repository_name = "orchex-builder-api"
  service         = "builder-api"
}

module "ecs_builder_api" {
  # ...
  service                             = "builder-api"
  target_group_arn                    = module.alb.target_groups["builder"].arn
  alb_security_group_id               = module.alb.security_group_id
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
}

module "rds" {
  source = "./modules/rds"

  name                                = "orchex"
  data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id
}
```

New ECS services that need database access must pass both `service` and `data_plane_client_security_group_id = module.ecs.data_plane_client_security_group_id` into `modules/ecs_service`.

After `terraform apply`, the ALB DNS name is available via `terraform output -json alb`. RDS connection details are under `terraform output -json rds`.

## 1. Apply infrastructure

```bash
cd infra

terraform init
terraform fmt -recursive
terraform plan
terraform apply
```

Inspect outputs:

```bash
terraform output
terraform output -json ecr_builder_api
terraform output -json alb | jq -r '.dns_name'
terraform output -json rds | jq
```

| Output                | Meaning                                                         |
| --------------------- | --------------------------------------------------------------- |
| `ecr_builder_api`     | ECR repository URL, name, ARN                                   |
| `ecr_db_migrate`      | Shared goose migration image repository                         |
| `alb`                 | ALB DNS name, target groups, security groups                    |
| `ecs`                 | Shared Fargate cluster ARN / name, `data_plane_client` SG       |
| `ecs_db_migrate`      | Migrate task definition + `run_task_network_configuration`      |
| `ecs_builder_api`     | Builder service task definition, security group, etc.           |
| `database_url_secret` | Shared `orchex/DATABASE_URL` secret ARN / name                  |
| `rds`                 | RDS connection metadata only (see fields below; no credentials) |

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

## 2. Build and push the migrate image, then run migrations

After RDS and the shared `orchex/DATABASE_URL` secret exist, apply schema changes with a **one-shot Fargate task** (not a long-running service).

Use the same region as `var.aws_region`. In **zsh**, always write `${REPO_URL}:latest` (not `$REPO_URL:latest`).

```bash
cd infra

MIGRATE_REPO=$(terraform output -json ecr_db_migrate | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region
CLUSTER=$(terraform output -json ecs | jq -r '.name')

aws ecr get-login-password --region "$REGION" --profile orchex \
  | docker login --username AWS --password-stdin "$(echo "$MIGRATE_REPO" | cut -d/ -f1)"

cd ..
docker build -f Dockerfile.migrate -t "${MIGRATE_REPO}:latest" .
docker push "${MIGRATE_REPO}:latest"
```

Run goose against RDS (reads `GOOSE_DBSTRING` from the shared `orchex/DATABASE_URL` secret):

```bash
cd infra

TASK_DEF=$(terraform output -json ecs_db_migrate | jq -r '.task_definition_family')
NET=$(terraform output -json ecs_db_migrate | jq -r '.run_task_network_configuration')
SUBNETS=$(echo "$NET" | jq -r '.subnets')
SGS=$(echo "$NET" | jq -r '.security_groups')

TASK_ARN=$(aws ecs run-task \
  --cluster "$CLUSTER" \
  --task-definition "$TASK_DEF" \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNETS],securityGroups=[$SGS],assignPublicIp=ENABLED}" \
  --region ap-south-1 \
  --profile orchex \
  --query 'tasks[0].taskArn' \
  --output text)

aws ecs wait tasks-stopped --cluster "$CLUSTER" --tasks "$TASK_ARN" --region ap-south-1 --profile orchex

aws ecs describe-tasks \
  --cluster "$CLUSTER" \
  --tasks "$TASK_ARN" \
  --region ap-south-1 \
  --profile orchex \
  --query 'tasks[0].containers[0].{exitCode:exitCode,reason:reason}' \
  --output json
```

Expect `"exitCode": 0`. Re-run migrations after pushing a new migrate image whenever `db/migrations/` changes.

## 3. Build and push the builder API image

Use the same region as `var.aws_region`. In **zsh**, always write `${REPO_URL}:tag` (not `$REPO_URL:tag`).

```bash
cd infra

REPO_URL=$(terraform output -json ecr_builder_api | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region

# Username is always "AWS". Password is a short-lived token from the CLI.
aws ecr get-login-password --region "$REGION" --profile orchex \
  | docker login --username AWS --password-stdin "$(echo "$REPO_URL" | cut -d/ -f1)"

# Build from the repository root, tagged for ECR (amd64 is pinned in the Dockerfile)
cd ..
docker build -t "${REPO_URL}:latest" .

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

Hit the API via the ALB:

```bash
curl "$(cd infra && terraform output -json alb | jq -r '.dns_name')/health/builder"
```

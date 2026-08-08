# Orchex infrastructure

Terraform for AWS resources. Provisions **ECR**, an **Application Load Balancer (ALB)**, a shared **ECS Fargate cluster**, and an **ECS service** for the workflow builder API (`orchex-builder-api`).

Terraform creates and wires the infrastructure. Building and pushing the container image is done with Docker + the AWS CLI.

## Layout

```text
infra/
  terraform.tf      # required Terraform / provider versions
  providers.tf      # AWS provider (region, profile, default tags)
  variables.tf      # shared variables (e.g. region)
  main.tf           # root modules (ECR, ALB, ECS, ECS service)
  outputs.tf        # root outputs
  modules/
    ecr/            # ECR repository for container images
    alb/            # internet-facing ALB + target group
    ecs/            # shared Fargate cluster
    ecs_service/    # Fargate service (task definition, SG, ALB attachment)
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

## Infrastructure architecture

Root `main.tf` composes four modules. Traffic flows from the ALB to Fargate tasks running in the default VPC.

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
│  • SG: only ALB SG may reach :8080  │
│  • registers task IPs with ALB TG   │
└──────────────┬──────────────────────┘
               │ runs on
               ▼
┌─────────────────────────────────────┐
│  ECS cluster (modules/ecs)          │
│  • Fargate-only capacity provider   │
│  • shared by services in this stack │
└─────────────────────────────────────┘

Image source: ECR (modules/ecr) → orchex-builder-api:latest
```

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
- Container Insights is disabled for early-stage cost; enable later if needed.

### `modules/ecs_service`

- Wraps [terraform-aws-modules/ecs/aws//modules/service](https://registry.terraform.io/modules/terraform-aws-modules/ecs/aws).
- Runs the builder API as a Fargate service (`orchex-builder-api`) on the shared cluster.
- Pulls the image from **ECR** (`orchex-builder-api:latest`).
- **Load balancer**: attaches the service to the ALB target group; ECS keeps registrations in sync as tasks start/stop.
- **Networking**: tasks get a public IP in default VPC subnets; ingress on the container port is restricted to the **ALB security group** (not open to the internet).
- Container runs with a read-only root filesystem (distroless image).

Wiring in root `main.tf`:

```hcl
module "ecs_builder_api" {
  # ...
  target_group_arn      = module.alb.target_groups["builder"].arn
  alb_security_group_id = module.alb.security_group_id
}
```

After `terraform apply`, the ALB DNS name is available via `terraform output -json alb`.

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
```

| Output            | Meaning                                                 |
| ----------------- | ------------------------------------------------------- |
| `ecr_builder_api` | ECR repository URL, name, ARN                         |
| `alb`             | ALB DNS name, target groups, security groups            |
| `ecs`             | Shared Fargate cluster ARN / name                       |
| `ecs_builder_api` | Builder service task definition, security group, etc. |

## 2. Build and push the builder API image

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

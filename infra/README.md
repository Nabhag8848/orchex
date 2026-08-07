# Orchex infrastructure

Terraform for AWS resources. Currently provisions **ECR** for the workflow builder API (`orchex-builder-api`).

Terraform creates the repository only. Building and pushing the image is done with Docker + the AWS CLI.

## Layout

```text
infra/
  terraform.tf      # required Terraform / provider versions
  providers.tf      # AWS provider (region, profile, default tags)
  variables.tf      # shared variables (e.g. region)
  main.tf           # root modules (ECR builder API)
  outputs.tf        # root outputs
  modules/
    ecr/            # reusable ECR module
```

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) matching `required_version` in `terraform.tf`
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) v2
- [Docker](https://docs.docker.com/get-docker/)
- A working AWS login (named profile, SSO, or `aws login`)
- [`jq`](https://jqlang.github.io/jq/) (optional, for reading JSON outputs)

Set the AWS profile name in `providers.tf` to match your local setup. Do not commit access keys or secrets.

Default region is `ap-south-1` (`var.aws_region`). Override when needed:

```bash
terraform apply -var='aws_region=YOUR_REGION'
```

## 1. Create the ECR repository

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
```

| Field                    | Meaning                                                 |
| ------------------------ | ------------------------------------------------------- |
| `repository_url`         | Registry host + repository path (use for `docker push`) |
| `repository_name`        | ECR repository name                                     |
| `repository_arn`         | Full ARN                                                |
| `repository_registry_id` | Registry / account ID                                   |

## 2. Build and push the builder API image

Use the same region as `var.aws_region`. In **zsh**, always write `${REPO_URL}:tag` (not `$REPO_URL:tag`).

```bash
cd infra

REPO_URL=$(terraform output -json ecr_builder_api | jq -r '.repository_url')
REGION=ap-south-1   # must match var.aws_region

# Username is always "AWS". Password is a short-lived token from the CLI.
aws ecr get-login-password --region "$REGION" \
  | docker login --username AWS --password-stdin "$(echo "$REPO_URL" | cut -d/ -f1)"

# Build from the repository root, tagged for ECR
cd ..
docker build -t "${REPO_URL}:latest" .

docker push "${REPO_URL}:latest"
```

Confirm the image landed in ECR:

```bash
aws ecr describe-images \
  --repository-name "$(cd infra && terraform output -json ecr_builder_api | jq -r '.repository_name')" \
  --region "$REGION"
```


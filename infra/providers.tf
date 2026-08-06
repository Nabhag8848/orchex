provider "aws" {
  region  = var.aws_region
  profile = "orchex"
  default_tags {
    tags = {
      Project     = "orchex"
      Environment = "production"
    }
  }
}
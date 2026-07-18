terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "pcp-terraform-state"
    key    = "production/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "pcp-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = var.environment != "production"

  tags = local.common_tags
}

# RDS PostgreSQL
resource "aws_db_instance" "pcp_postgres" {
  identifier     = "pcp-${var.environment}"
  engine         = "postgres"
  engine_version = "16"
  instance_class = var.db_instance_class

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_encrypted     = true

  db_name  = "pcp"
  username = "pcp"
  password = var.db_password

  vpc_security_group_ids = [aws_security_group.db.id]
  db_subnet_group_name   = aws_db_subnet_group.pcp.name

  backup_retention_period = 7
  multi_az               = var.environment == "production"
  skip_final_snapshot    = var.environment != "production"

  tags = local.common_tags
}

# ElastiCache Redis
resource "aws_elasticache_cluster" "pcp_redis" {
  cluster_id           = "pcp-${var.environment}"
  engine               = "redis"
  node_type            = var.redis_node_type
  num_cache_nodes      = 1
  port                 = 6379
  security_group_ids   = [aws_security_group.redis.id]
  subnet_group_name    = aws_elasticache_subnet_group.pcp.name

  tags = local.common_tags
}

# EKS Cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = "pcp-${var.environment}"
  cluster_version = "1.30"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    default = {
      instance_types = [var.eks_instance_type]
      min_size       = var.eks_min_nodes
      max_size       = var.eks_max_nodes
      desired_size   = var.eks_desired_nodes
    }
  }

  tags = local.common_tags
}

# Security Groups
resource "aws_security_group" "db" {
  name_prefix = "pcp-db-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [module.eks.cluster_security_group_id]
  }
}

resource "aws_security_group" "redis" {
  name_prefix = "pcp-redis-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [module.eks.cluster_security_group_id]
  }
}

resource "aws_db_subnet_group" "pcp" {
  name       = "pcp-${var.environment}"
  subnet_ids = module.vpc.private_subnets
}

resource "aws_elasticache_subnet_group" "pcp" {
  name       = "pcp-${var.environment}"
  subnet_ids = module.vpc.private_subnets
}

locals {
  common_tags = {
    Project     = "pcp"
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# Variables
variable "aws_region" {
  default = "us-east-1"
}

variable "environment" {
  default = "staging"
}

variable "db_password" {
  sensitive = true
}

variable "db_instance_class" {
  default = "db.t3.medium"
}

variable "redis_node_type" {
  default = "cache.t3.micro"
}

variable "eks_instance_type" {
  default = "t3.medium"
}

variable "eks_min_nodes" {
  default = 2
}

variable "eks_max_nodes" {
  default = 10
}

variable "eks_desired_nodes" {
  default = 3
}

# Outputs
output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "rds_endpoint" {
  value = aws_db_instance.pcp_postgres.endpoint
}

output "redis_endpoint" {
  value = aws_elasticache_cluster.pcp_redis.cache_nodes[0].address
}

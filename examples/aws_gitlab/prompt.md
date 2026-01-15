# AWS + GitLab Infrastructure Deployment

Generate a complete AWS infrastructure with GitLab CI/CD pipeline.

## AWS Requirements

Create CloudFormation templates for:
1. **VPC** - Multi-AZ VPC with public and private subnets
2. **EKS Cluster** - Managed Kubernetes cluster in the VPC
3. **RDS Database** - PostgreSQL database in private subnets

Each stack should export its key outputs for cross-stack references.

## GitLab Requirements

Create a GitLab CI/CD pipeline that:
1. Validates CloudFormation templates
2. Deploys stacks in dependency order (VPC → EKS → RDS)
3. Runs integration tests after deployment
4. Supports manual approval for production

## Cross-Domain Integration

The GitLab pipeline must reference AWS outputs:
- `${aws.vpc.outputs.vpc_id}` - VPC ID for EKS and RDS
- `${aws.eks.outputs.cluster_name}` - Cluster name for kubectl config
- `${aws.rds.outputs.endpoint}` - Database endpoint for app config

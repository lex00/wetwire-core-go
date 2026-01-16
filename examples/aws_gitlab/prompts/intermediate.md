# S3 Bucket with GitLab Deployment (Intermediate)

Create an S3 bucket with CloudFormation and a GitLab CI/CD pipeline for deployment.

## AWS CloudFormation

Create `cfn-templates/s3-bucket.yaml`:
- S3 bucket resource with versioning and SSE-S3 encryption
- PublicAccessBlockConfiguration to block all public access
- Outputs: BucketName, BucketArn

Use parameters for:
- Environment tag (dev/staging/prod)
- Bucket name suffix

## GitLab Pipeline

Create `.gitlab-ci.yml` with stages:
1. `validate` - Run `aws cloudformation validate-template`
2. `deploy` - Run `aws cloudformation deploy`
3. `verify` - Describe the stack and show outputs

Use GitLab CI variables for AWS credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_DEFAULT_REGION`).

## Cross-Domain Integration

The pipeline should reference CloudFormation outputs for verification:
- `${aws.s3.outputs.bucket_name}`
- `${aws.s3.outputs.bucket_arn}`

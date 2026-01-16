# S3 Bucket Template with GitLab Publishing (Intermediate)

Create an S3 bucket CloudFormation template and a GitLab CI/CD pipeline to publish it.

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
2. `publish` - Upload template to S3 templates bucket
3. `release` - Tag with version, create GitLab release

The pipeline publishes the template file - it does NOT execute `cloudformation deploy`.

Use GitLab CI variables for AWS credentials.

## Cross-Domain Integration

Pipeline references:
- `${aws.s3.outputs.bucket_name}` - Target bucket for template storage

# S3 Bucket with GitLab Deployment

Create an S3 bucket using CloudFormation and a GitLab CI/CD pipeline to deploy it.

## AWS Requirements

Create a CloudFormation template for an S3 bucket with:
- Versioning enabled
- Server-side encryption (AES256)
- Block public access

Export the bucket name and ARN as stack outputs.

## GitLab Requirements

Create a `.gitlab-ci.yml` pipeline that:
1. Validates the CloudFormation template
2. Deploys the stack to AWS
3. Outputs the deployed bucket name

## Cross-Domain Integration

The GitLab pipeline should reference the AWS outputs:
- `${aws.s3.outputs.bucket_name}` - Bucket name for verification
- `${aws.s3.outputs.bucket_arn}` - Bucket ARN for IAM policies

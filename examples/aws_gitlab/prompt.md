# S3 Bucket Template with GitLab Publishing

Create an S3 bucket CloudFormation template and a GitLab CI/CD pipeline to publish it.

## AWS Requirements

Create a CloudFormation template for an S3 bucket with:
- Versioning enabled
- Server-side encryption (AES256)
- Block public access

Export the bucket name and ARN as stack outputs.

## GitLab Requirements

Create a `.gitlab-ci.yml` pipeline that:
1. Validates the CloudFormation template syntax
2. Publishes the template to an S3 bucket for distribution
3. Tags the release with version info

The pipeline publishes the template - it does NOT execute it to create resources.

## Cross-Domain Integration

The GitLab pipeline should reference:
- `${aws.s3.outputs.bucket_name}` - Target bucket for publishing templates

# S3 + GitLab Deploy (Expert)

S3 bucket via CFN with GitLab CI/CD.

## AWS

`cfn-templates/s3-bucket.yaml`:
- S3 with versioning, SSE-S3, PublicAccessBlockConfiguration
- Outputs: BucketName, BucketArn
- Parameters: Environment, BucketSuffix

## GitLab

`.gitlab-ci.yml`:
- validate: `cfn validate-template`
- deploy: `cfn deploy --no-fail-on-empty-changeset`
- verify: describe-stacks

AWS creds via CI vars. Include rules for branch-based deployment.

Cross-refs: `${aws.s3.outputs.bucket_name}`, `${aws.s3.outputs.bucket_arn}`

# S3 Template + GitLab Publish (Expert)

S3 bucket CFN template with GitLab pipeline for publishing (not execution).

## AWS

`cfn-templates/s3-bucket.yaml`:
- S3 with versioning, SSE-S3, PublicAccessBlockConfiguration
- Outputs: BucketName, BucketArn
- Parameters: Environment, BucketSuffix

## GitLab

`.gitlab-ci.yml`:
- validate: `cfn validate-template`
- publish: `aws s3 cp` to templates bucket
- release: git tag, GitLab release

No `cloudformation deploy` - publish only.

Cross-ref: `${aws.s3.outputs.bucket_name}`

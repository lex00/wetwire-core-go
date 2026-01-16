# Comprehensive S3 Bucket Template with GitLab Publishing Pipeline (Verbose)

This document provides detailed requirements for creating an Amazon S3 bucket CloudFormation template, along with a GitLab CI/CD pipeline for validating and publishing the template.

## Background and Context

Amazon S3 (Simple Storage Service) is AWS's object storage service. CloudFormation allows us to define AWS resources in YAML or JSON templates that can be version-controlled and shared.

**Important Distinction**: This scenario creates a pipeline that **publishes** the CloudFormation template to a distribution location (S3 bucket). It does NOT execute the template to create resources. Template execution happens separately when consumers use the published template.

## AWS CloudFormation Requirements

### Template Location and Format

Create a CloudFormation template at `cfn-templates/s3-bucket.yaml` using YAML format.

### S3 Bucket Resource Configuration

The S3 bucket should be configured with the following properties:

#### Versioning Configuration
Enable versioning to maintain a history of all object versions.

```yaml
VersioningConfiguration:
  Status: Enabled
```

#### Server-Side Encryption
Configure default encryption using Amazon S3 managed keys (SSE-S3).

```yaml
BucketEncryption:
  ServerSideEncryptionConfiguration:
    - ServerSideEncryptionByDefault:
        SSEAlgorithm: AES256
```

#### Public Access Block
Block all public access to the bucket.

```yaml
PublicAccessBlockConfiguration:
  BlockPublicAcls: true
  BlockPublicPolicy: true
  IgnorePublicAcls: true
  RestrictPublicBuckets: true
```

### Template Parameters

Include the following parameters:

1. **Environment** (String)
   - Allowed values: dev, staging, prod
   - Default: dev

2. **BucketNameSuffix** (String)
   - Pattern: [a-z0-9-]+

### Stack Outputs

Export the following values:

1. **BucketName** - The logical bucket name
2. **BucketArn** - The bucket's Amazon Resource Name

## GitLab CI/CD Pipeline Requirements

### Pipeline Purpose

The pipeline validates and **publishes** the CloudFormation template. It does NOT execute `aws cloudformation deploy` or create any AWS resources.

### Pipeline File

Create `.gitlab-ci.yml` in the repository root.

### Required CI/CD Variables

- `AWS_ACCESS_KEY_ID` - AWS IAM access key
- `AWS_SECRET_ACCESS_KEY` - AWS IAM secret key
- `AWS_DEFAULT_REGION` - Target AWS region
- `TEMPLATES_BUCKET` - S3 bucket for storing published templates

### Pipeline Stages

#### Stage 1: Validate
- **Purpose**: Verify CloudFormation template syntax
- **Command**: `aws cloudformation validate-template --template-body file://cfn-templates/s3-bucket.yaml`
- **Failure behavior**: Block publishing if validation fails

#### Stage 2: Publish
- **Purpose**: Upload template to S3 for distribution
- **Commands**:
  - `aws s3 cp cfn-templates/s3-bucket.yaml s3://${TEMPLATES_BUCKET}/templates/s3-bucket-${CI_COMMIT_TAG}.yaml`
  - Also publish as `latest`: `aws s3 cp cfn-templates/s3-bucket.yaml s3://${TEMPLATES_BUCKET}/templates/s3-bucket-latest.yaml`

#### Stage 3: Release
- **Purpose**: Create versioned release
- **Actions**:
  - Create GitLab release with template artifact
  - Tag with semantic version

### Branch-Based Rules

- Run validation on all branches
- Publish only from tagged commits on `main` branch

## Cross-Domain Integration

The GitLab pipeline references:
- `${aws.s3.outputs.bucket_name}` - Target bucket for published templates

## Expected Output Structure

```
output/
├── cfn-templates/
│   └── s3-bucket.yaml      # CloudFormation template
└── .gitlab-ci.yml          # GitLab CI/CD pipeline (validate + publish)
```

## Success Criteria

1. CloudFormation template passes validation
2. Template follows AWS security best practices
3. GitLab pipeline validates and publishes (does NOT deploy)
4. Published templates are versioned

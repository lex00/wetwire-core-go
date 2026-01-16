# Comprehensive S3 Bucket Deployment with GitLab CI/CD Pipeline (Verbose)

This document provides detailed requirements for creating an Amazon S3 bucket using AWS CloudFormation infrastructure-as-code, along with a complete GitLab CI/CD pipeline configuration for automated deployment.

## Background and Context

Amazon S3 (Simple Storage Service) is AWS's object storage service. CloudFormation allows us to define AWS resources in YAML or JSON templates that can be version-controlled and deployed consistently across environments.

GitLab CI/CD provides continuous integration and deployment capabilities that we'll use to automate the validation and deployment of our CloudFormation stack.

## AWS CloudFormation Requirements

### Template Location and Format

Create a CloudFormation template at `cfn-templates/s3-bucket.yaml` using YAML format (preferred for readability).

### S3 Bucket Resource Configuration

The S3 bucket should be configured with the following properties:

#### Versioning Configuration
Enable versioning on the bucket to maintain a complete history of all object versions. This allows recovery from accidental deletions and overwrites.

```yaml
VersioningConfiguration:
  Status: Enabled
```

#### Server-Side Encryption
Configure default encryption using Amazon S3 managed keys (SSE-S3) with AES-256 encryption algorithm. All objects stored in the bucket will be automatically encrypted at rest.

```yaml
BucketEncryption:
  ServerSideEncryptionConfiguration:
    - ServerSideEncryptionByDefault:
        SSEAlgorithm: AES256
```

#### Public Access Block
Block all public access to the bucket. This is a security best practice that prevents accidental exposure of sensitive data.

```yaml
PublicAccessBlockConfiguration:
  BlockPublicAcls: true
  BlockPublicPolicy: true
  IgnorePublicAcls: true
  RestrictPublicBuckets: true
```

### Template Parameters

Include the following parameters to allow customization:

1. **Environment** (String)
   - Allowed values: dev, staging, prod
   - Default: dev
   - Description: Deployment environment for tagging

2. **BucketNameSuffix** (String)
   - Description: Suffix to append to bucket name for uniqueness
   - Pattern: [a-z0-9-]+

### Stack Outputs

Export the following values for use by other stacks and the CI/CD pipeline:

1. **BucketName**
   - Value: The logical bucket name
   - Export: `${AWS::StackName}-BucketName`

2. **BucketArn**
   - Value: The bucket's Amazon Resource Name
   - Export: `${AWS::StackName}-BucketArn`

## GitLab CI/CD Pipeline Requirements

### Pipeline File

Create `.gitlab-ci.yml` in the repository root.

### Required CI/CD Variables

The following variables must be configured in GitLab CI/CD settings:

- `AWS_ACCESS_KEY_ID` - AWS IAM access key
- `AWS_SECRET_ACCESS_KEY` - AWS IAM secret key
- `AWS_DEFAULT_REGION` - Target AWS region (e.g., us-east-1)

### Pipeline Stages

Configure three sequential stages:

#### Stage 1: Validate
- **Purpose**: Verify the CloudFormation template syntax is correct
- **Command**: `aws cloudformation validate-template --template-body file://cfn-templates/s3-bucket.yaml`
- **Failure behavior**: Block deployment if validation fails

#### Stage 2: Deploy
- **Purpose**: Create or update the CloudFormation stack
- **Command**: `aws cloudformation deploy --template-file cfn-templates/s3-bucket.yaml --stack-name s3-bucket-stack --parameter-overrides Environment=$CI_ENVIRONMENT_NAME --no-fail-on-empty-changeset`
- **Notes**: The `--no-fail-on-empty-changeset` flag prevents failures when no changes are detected

#### Stage 3: Verify
- **Purpose**: Confirm successful deployment and display outputs
- **Commands**:
  - `aws cloudformation describe-stacks --stack-name s3-bucket-stack`
  - Display stack outputs including bucket name and ARN

### Branch-Based Deployment Rules

Configure the pipeline to:
- Run validation on all branches
- Deploy to dev environment from feature branches
- Deploy to staging from the `develop` branch
- Deploy to production from the `main` branch with manual approval

## Cross-Domain Integration

The GitLab pipeline must correctly reference and validate the following AWS CloudFormation outputs:

- `${aws.s3.outputs.bucket_name}` - The deployed S3 bucket name, used for verification and downstream configuration
- `${aws.s3.outputs.bucket_arn}` - The bucket's ARN, used for IAM policy configuration

These references ensure that the GitLab pipeline is aware of and validates the actual deployed resources.

## Expected Output Structure

After successful execution, the scenario should produce:

```
output/
├── cfn-templates/
│   └── s3-bucket.yaml      # CloudFormation template
└── .gitlab-ci.yml          # GitLab CI/CD pipeline
```

## Success Criteria

1. CloudFormation template passes validation
2. Template follows AWS best practices for S3 security
3. GitLab pipeline is syntactically correct
4. Pipeline stages execute in correct order
5. Cross-domain references are properly configured

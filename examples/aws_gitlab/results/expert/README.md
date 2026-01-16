# S3 Bucket CloudFormation Template with GitLab Publishing Pipeline

Production-ready infrastructure for publishing CloudFormation templates to AWS S3 via GitLab CI/CD.

## Architecture

### CloudFormation Template (`cfn-templates/s3-bucket.yaml`)

**Features:**
- **Versioning**: All object versions retained for rollback capability
- **Encryption**: Server-side encryption with AES256 (S3-managed keys)
- **Security**:
  - Public access blocking on all four dimensions
  - Bucket policy denying unencrypted uploads
  - Bucket policy enforcing HTTPS/TLS connections
- **Parameters**: Environment (dev/staging/prod) and BucketSuffix for name uniqueness
- **Outputs**: BucketName, BucketArn, Region, VersioningStatus

**Parameters:**
- `Environment`: Deployment environment (default: dev)
- `BucketSuffix`: Suffix for bucket name uniqueness (alphanumeric + hyphens)

**Outputs:**
- `BucketName`: CloudFormation export for cross-stack references
- `BucketArn`: IAM policy resource identifier
- `BucketRegion`: Deployment region
- `VersioningStatus`: Versioning configuration state

### GitLab CI/CD Pipeline (`.gitlab-ci.yml`)

**Pipeline Stages:**

1. **Validate**
   - `validate:cfn` — CloudFormation template syntax validation via AWS API
   - `lint:cfn` — Best practices linting using cfn-lint
   - Both stages run on all branches and MRs

2. **Publish**
   - `publish:templates` — Upload validated templates to S3 with metadata
   - Runs only on `main`, `master`, and tag branches
   - Stores templates in versioned S3 path: `s3://bucket/cfn-templates/{branch}/{commit-sha}/{template}`
   - Includes object metadata: commit SHA, branch, pipeline ID
   - Uses STANDARD_IA storage class for cost optimization

3. **Release**
   - `release:create` — Create GitLab release tags with deployment details
   - Triggered on git tags only
   - Links release to S3 bucket location

4. **Maintenance**
   - `cleanup:old-versions` — Scheduled job to delete template versions >30 days old
   - Prevents S3 storage cost drift

## Usage

### Deploy Stack from Published Template

```bash
# Reference templates published by the pipeline
aws cloudformation create-stack \
  --stack-name my-bucket-stack \
  --template-url https://${TEMPLATES_BUCKET}.s3.amazonaws.com/cfn-templates/main/a1b2c3d4/s3-bucket.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=prod \
    ParameterKey=BucketSuffix,ParameterValue=myapp
```

### Cross-Stack Reference

Use CloudFormation exports in other templates:

```yaml
Resources:
  AppStorage:
    Type: AWS::S3::Bucket
    Properties:
      # Reference bucket from S3 stack
      NotificationConfiguration:
        # ... use ${aws.s3.outputs.bucket_name}
```

## CI/CD Workflow

### On Every Push to Branch
1. Validate template syntax
2. Lint for best practices
3. Skip S3 publish (templates not production-ready)

### On Merge to Main/Master
1. Validate template syntax
2. Lint for best practices
3. **Publish** templates to S3 with branch/commit metadata
4. Output artifact contains S3 path for deployment

### On Git Tag (Release)
1. Run full validation/lint/publish pipeline
2. **Create GitLab Release** with link to published templates
3. Release notes reference S3 bucket location

## Security & Best Practices

**Encryption:**
- Server-side encryption (AES256) enforced for all uploads
- Deny policy for unencrypted `PutObject` operations

**Access Control:**
- Complete public access blocking
- Deny policy for non-HTTPS connections
- Bucket policy requires explicit IAM permissions

**Versioning & Audit:**
- S3 object versioning enables rollback
- Pipeline metadata in S3 object tags for audit trail
- 30-day retention for old template versions (configurable)

**Cost Optimization:**
- STANDARD_IA storage class for infrequently-accessed templates
- Automated cleanup of versions >30 days old
- Scheduled maintenance job (configurable frequency)

## Configuration

### Customize Template Bucket

Update `.gitlab-ci.yml` variable:
```yaml
TEMPLATES_BUCKET: my-custom-templates-bucket
```

### Adjust Retention Policy

Modify cleanup job:
```yaml
cleanup:old-versions:
  script:
    - cutoff_date=$(date -d "60 days ago" +%Y-%m-%d)  # Change to 60 days
```

### Change Storage Class

Modify publish job:
```yaml
aws s3 cp $template $s3_path \
  --storage-class GLACIER  # For archival
```

## Testing

### Local CloudFormation Validation

```bash
aws cloudformation validate-template \
  --template-body file://cfn-templates/s3-bucket.yaml
```

### Local Linting

```bash
cfn-lint cfn-templates/s3-bucket.yaml
```

### Deploy to Dev Environment

```bash
aws cloudformation create-stack \
  --stack-name dev-s3-bucket \
  --template-body file://cfn-templates/s3-bucket.yaml \
  --parameters ParameterKey=Environment,ParameterValue=dev \
                ParameterKey=BucketSuffix,ParameterValue=dev-test
```

## Monitoring & Troubleshooting

### Check Published Templates

```bash
aws s3 ls s3://${TEMPLATES_BUCKET}/cfn-templates/ --recursive
```

### Verify Template Versions

```bash
aws s3api list-object-versions \
  --bucket ${TEMPLATES_BUCKET} \
  --prefix cfn-templates/
```

### View Pipeline Logs

Check GitLab CI/CD job logs for validation/publish errors:
- Validation failures indicate template syntax issues
- Lint failures suggest best practice violations
- Publish failures indicate S3 permissions or connectivity issues

## File Structure

```
results/expert/
├── cfn-templates/
│   └── s3-bucket.yaml          # S3 bucket CloudFormation template
├── .gitlab-ci.yml              # Pipeline definition
└── README.md                   # This file
```

## References

- [AWS CloudFormation S3 Bucket](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html)
- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
- [cfn-lint Rules](https://github.com/aws-cloudformation/cfn-lint)

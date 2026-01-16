# Secure S3 Bucket CloudFormation Template

This directory contains a production-ready CloudFormation template for managing secure S3 buckets with GitLab CI/CD pipeline integration.

## Files

- **s3-bucket-template.yaml** - CloudFormation template with security best practices
- **.gitlab-ci.yml** - GitLab CI/CD pipeline for validation and publishing

## CloudFormation Template Features

### Security Best Practices Implemented

✅ **Encryption**
- Server-side encryption (SSE-S3) enabled by default
- Can be upgraded to KMS for enhanced key management

✅ **Versioning**
- Optional object versioning for data protection
- Lifecycle policies to archive old versions to Glacier
- Automatic cleanup of non-current versions after 90 days

✅ **Access Control**
- Block all public access enabled
- Bucket policies deny:
  - Unencrypted uploads
  - Insecure transport (non-HTTPS)
  - Deletion of object versions (ransomware protection)

✅ **Logging**
- Optional access logging to separate bucket
- All logs encrypted and versioned
- Access logs prefixed for easy filtering

✅ **Lifecycle Management**
- Incomplete multipart uploads cleanup (7 days)
- Automatic archival of old versions (30 days → Glacier)
- Automatic deletion of ancient versions (90 days)

✅ **Monitoring & Compliance**
- CloudFormation tags for resource tracking
- Stack outputs for integration with other tools
- DeletionPolicy: Retain to prevent accidental deletion

## Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `BucketName` | String | Required | Globally unique bucket name (lowercase, hyphens allowed) |
| `Environment` | String | `dev` | Environment: dev, staging, or prod |
| `EnableVersioning` | String | `true` | Enable S3 versioning |
| `EnableLogging` | String | `true` | Enable access logging |

## Deployment

### Manual via CloudFormation Console

1. Navigate to AWS CloudFormation console
2. Create stack
3. Upload `s3-bucket-template.yaml`
4. Fill in parameters (BucketName is required)
5. Review security settings on the final page
6. Create stack

### Via AWS CLI

```bash
aws cloudformation create-stack \
  --stack-name my-s3-bucket \
  --template-body file://s3-bucket-template.yaml \
  --parameters \
    ParameterKey=BucketName,ParameterValue=my-unique-bucket-name-12345 \
    ParameterKey=Environment,ParameterValue=prod \
    ParameterKey=EnableVersioning,ParameterValue=true \
    ParameterKey=EnableLogging,ParameterValue=true \
  --region us-east-1
```

### Via GitLab CI/CD

**This pipeline is for publishing the template to S3 for use by other deployments.**

1. Push changes to merge request or main branch
2. Pipeline automatically validates and lints the template
3. Click "publish_template" job to manually publish to S3
4. Template URL is output to job logs and artifacts

## GitLab Pipeline Stages

### Validate Stage
- **validate_template** - CloudFormation template validation
- **lint_template** - Lint checks with cfn-lint
- **policy_check** - Security policy verification

### Publish Stage
- **publish_template** - Publish validated template to S3 (manual trigger)
- **publish_release** - Create release notes for tagged versions (manual trigger)

## Outputs

After stack creation, CloudFormation provides:

- `BucketName` - The S3 bucket name
- `BucketArn` - ARN for IAM policies
- `BucketDomainName` - DNS name for bucket access
- `LoggingBucketName` - Name of the logging bucket (if enabled)

## Cost Considerations

- **Versioning**: Stores multiple versions of objects (increased storage cost)
- **Logging**: Separate logging bucket creates additional storage
- **Lifecycle Policies**: Move old versions to Glacier for cost savings
- **Encryption**: No additional cost with SSE-S3; small cost with KMS

## Updating the Template

1. Edit `s3-bucket-template.yaml`
2. Commit and push to feature branch
3. Create merge request - pipeline automatically validates
4. After approval, merge to main
5. Pipeline will lint and validate on main branch
6. Manually trigger publish_template to deploy to S3

## Customization

### Enable KMS Encryption (Production)

```yaml
BucketEncryption:
  ServerSideEncryptionConfiguration:
    - ServerSideEncryptionByDefault:
        SSEAlgorithm: aws:kms
        KMSMasterKeyID: arn:aws:kms:region:account:key/key-id
      BucketKeyEnabled: true
```

### Add Object Lock (Compliance)

```yaml
ObjectLockEnabled: true
ObjectLockConfiguration:
  ObjectLockEnabled: Enabled
  Rule:
    DefaultRetention:
      Mode: GOVERNANCE
      Days: 30
```

### Add MFA Delete Protection

Requires root account credentials and manual AWS CLI configuration.

## Troubleshooting

### Template Validation Fails
- Check BucketName format: must be lowercase, hyphens allowed, 3-63 characters
- Ensure bucket name is globally unique across all AWS accounts

### Publish to S3 Fails
- Verify AWS credentials in GitLab CI/CD variables
- Ensure CloudFormation templates S3 bucket exists or create it first
- Check IAM permissions for S3 and CloudFormation

### Lifecycle Rules Not Working
- Ensure versioning is enabled (`EnableVersioning=true`)
- Lifecycle policies only apply to objects created after rule activation
- Allow 24 hours for lifecycle policy evaluation

## Security Notes

⚠️ **Important**: This template provides good baseline security but consider:
- **KMS Keys**: Use customer-managed keys in production for key rotation control
- **MFA Delete**: Enable MFA Delete for critical data (requires root account)
- **Bucket Policies**: Restrict further by IP, VPC, or specific roles
- **CloudTrail**: Enable CloudTrail logging for API audit trail
- **Compliance**: Add Object Lock for regulatory requirements (WORM)

## Support

For issues or questions:
1. Review CloudFormation stack events for errors
2. Check GitLab CI/CD pipeline logs
3. Validate template locally: `cfn-lint s3-bucket-template.yaml`
4. Test parameters before production deployment

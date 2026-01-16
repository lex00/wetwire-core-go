# Deployment Guide

This guide explains how to deploy the secure S3 bucket using the CloudFormation template and GitLab pipeline.

## Prerequisites

- AWS Account with appropriate IAM permissions
- GitLab project configured with AWS credentials
- GitLab Runner with AWS CLI access (for pipeline execution)

## GitLab CI/CD Setup

### 1. Configure AWS Credentials in GitLab

Add these variables to your GitLab project (Settings → CI/CD → Variables):

```
AWS_ACCESS_KEY_ID          # Your AWS access key
AWS_SECRET_ACCESS_KEY      # Your AWS secret access key
AWS_DEFAULT_REGION         # us-east-1 (or your region)
AWS_ACCOUNT_ID             # Your AWS account ID (12 digits)
```

Mark `AWS_SECRET_ACCESS_KEY` as **Protected** and **Masked**.

### 2. Create S3 Bucket for Template Storage (One-time)

The pipeline publishes templates to S3. Create the bucket first:

```bash
aws s3 mb s3://cloudformation-templates-${AWS_ACCOUNT_ID}-us-east-1 \
  --region us-east-1 \
  --create-bucket-configuration LocationConstraint=us-east-1
```

Enable versioning on this bucket:

```bash
aws s3api put-bucket-versioning \
  --bucket cloudformation-templates-${AWS_ACCOUNT_ID}-us-east-1 \
  --versioning-configuration Status=Enabled \
  --region us-east-1
```

### 3. Push to GitLab

```bash
git add s3-bucket-template.yaml .gitlab-ci.yml
git commit -m "Add secure S3 bucket CloudFormation template"
git push origin main
```

Pipeline will automatically:
- ✅ Validate the template
- ✅ Run lint checks
- ✅ Check security policies

## Deployment Methods

### Method 1: Via GitLab Pipeline (Recommended)

**Step 1: Validate Template**
- Push code to `main` or create merge request
- Pipeline automatically runs validation, lint, and policy checks
- All checks must pass (shown in merge request)

**Step 2: Publish Template to S3**
- After merge to `main`, click "Publish" stage
- Click "publish_template" job's play button
- Job outputs template URL to S3

**Step 3: Deploy Stack (Manual)**
- Use AWS Console or CLI with the published template:

```bash
aws cloudformation create-stack \
  --stack-name my-s3-bucket \
  --template-url https://cloudformation-templates-${AWS_ACCOUNT_ID}-us-east-1.s3.us-east-1.amazonaws.com/s3-bucket/abc1234/s3-bucket-template.yaml \
  --parameters file://parameters-example.json \
  --region us-east-1
```

### Method 2: AWS CloudFormation Console

1. Go to [AWS CloudFormation Console](https://console.aws.amazon.com/cloudformation/)
2. Click **Create Stack**
3. Select **Upload a template file**
4. Upload `s3-bucket-template.yaml`
5. Fill in parameters:
   - **BucketName**: `my-secure-bucket-001` (must be globally unique)
   - **Environment**: `dev` (or `staging`, `prod`)
   - **EnableVersioning**: `true`
   - **EnableLogging**: `true`
6. Click **Next** and review stack settings
7. Check acknowledgment for IAM resource creation
8. Click **Create Stack**

Monitor progress in the Events tab.

### Method 3: AWS CLI

```bash
# Create stack with parameters file
aws cloudformation create-stack \
  --stack-name my-s3-bucket \
  --template-body file://s3-bucket-template.yaml \
  --parameters file://parameters-example.json \
  --tags Key=Project,Value=wetwire Key=Environment,Value=dev \
  --region us-east-1

# Watch stack creation progress
aws cloudformation wait stack-create-complete \
  --stack-name my-s3-bucket \
  --region us-east-1

# Get stack outputs
aws cloudformation describe-stacks \
  --stack-name my-s3-bucket \
  --query 'Stacks[0].Outputs' \
  --region us-east-1
```

### Method 4: AWS CLI with Inline Parameters

```bash
aws cloudformation create-stack \
  --stack-name my-s3-bucket \
  --template-body file://s3-bucket-template.yaml \
  --parameters \
    ParameterKey=BucketName,ParameterValue=my-secure-bucket-001 \
    ParameterKey=Environment,ParameterValue=dev \
    ParameterKey=EnableVersioning,ParameterValue=true \
    ParameterKey=EnableLogging,ParameterValue=true \
  --region us-east-1
```

## Post-Deployment

### Verify Stack Creation

```bash
# Check stack status
aws cloudformation describe-stacks \
  --stack-name my-s3-bucket \
  --query 'Stacks[0].StackStatus' \
  --region us-east-1

# Should return: CREATE_COMPLETE
```

### Get Outputs

```bash
aws cloudformation describe-stacks \
  --stack-name my-s3-bucket \
  --query 'Stacks[0].Outputs[*].[OutputKey,OutputValue]' \
  --output table \
  --region us-east-1
```

### Test Bucket Access

```bash
# List bucket contents
aws s3 ls s3://my-secure-bucket-001/

# Upload a test file
echo "test data" > test.txt
aws s3 cp test.txt s3://my-secure-bucket-001/test.txt --sse AES256

# Verify encryption
aws s3api head-object \
  --bucket my-secure-bucket-001 \
  --key test.txt \
  --query 'ServerSideEncryption' \
  --region us-east-1

# Clean up
rm test.txt
aws s3 rm s3://my-secure-bucket-001/test.txt
```

### Enable MFA Delete (Optional - Requires Root)

```bash
# ⚠️ WARNING: This requires root account credentials
# 1. Enable versioning (already done by template)
# 2. Generate MFA device serial number
# 3. Run this command with MFA token:

aws s3api put-bucket-versioning \
  --bucket my-secure-bucket-001 \
  --versioning-configuration Status=Enabled,MFADelete=Enabled \
  --sse-customer-algorithm AES256 \
  --serial-number arn:aws:iam::ACCOUNT:mfa/root-device \
  --token-code 123456 \
  --region us-east-1
```

## Updating the Stack

### Minor Changes (Parameters)

```bash
aws cloudformation update-stack \
  --stack-name my-s3-bucket \
  --use-previous-template \
  --parameters ParameterKey=EnableVersioning,ParameterValue=false \
  --region us-east-1
```

### Template Changes

1. Edit `s3-bucket-template.yaml`
2. Commit and push (triggers pipeline validation)
3. Publish updated template to S3
4. Update stack with new template URL:

```bash
aws cloudformation update-stack \
  --stack-name my-s3-bucket \
  --template-url https://cloudformation-templates-${AWS_ACCOUNT_ID}-us-east-1.s3.us-east-1.amazonaws.com/s3-bucket/new-hash/s3-bucket-template.yaml \
  --region us-east-1
```

## Deletion

### Delete Stack (Removes S3 Bucket Resources)

```bash
aws cloudformation delete-stack \
  --stack-name my-s3-bucket \
  --region us-east-1

# Wait for completion
aws cloudformation wait stack-delete-complete \
  --stack-name my-s3-bucket \
  --region us-east-1
```

**Note**: S3 buckets have `DeletionPolicy: Retain` - they are NOT deleted when stack is deleted. This is intentional to prevent accidental data loss.

To delete the buckets manually:

```bash
# Empty and delete logging bucket
aws s3 rm s3://my-secure-bucket-001-logs-${AWS_ACCOUNT_ID} --recursive
aws s3api delete-bucket --bucket my-secure-bucket-001-logs-${AWS_ACCOUNT_ID}

# Empty and delete main bucket
aws s3 rm s3://my-secure-bucket-001 --recursive
aws s3api delete-bucket --bucket my-secure-bucket-001
```

## Troubleshooting

### Stack Creation Fails with "BucketAlreadyOwnedByYou"

The bucket name must be globally unique. Try a different name:

```bash
aws cloudformation update-stack-set \
  --stack-set-name my-s3-bucket \
  --parameters \
    ParameterKey=BucketName,ParameterValue=my-secure-bucket-$(date +%s)
```

### Stack Creation Fails with Insufficient Permissions

Verify IAM user has these permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*",
        "cloudformation:*"
      ],
      "Resource": "*"
    }
  ]
}
```

### Pipeline Fails: "Unable to assume role"

Verify AWS credentials in GitLab CI/CD variables:
- Settings → CI/CD → Variables
- Check AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are correct
- Ensure credentials have S3 and CloudFormation permissions

### Publish to S3 Fails

Verify template storage bucket exists:

```bash
aws s3api head-bucket \
  --bucket cloudformation-templates-${AWS_ACCOUNT_ID}-us-east-1 \
  --region us-east-1
```

If not found, create it (see Prerequisites section).

## Monitoring

### CloudWatch Metrics

Monitor bucket metrics in CloudWatch:

```bash
aws cloudwatch list-metrics \
  --namespace AWS/S3 \
  --dimensions Name=BucketName,Value=my-secure-bucket-001 \
  --region us-east-1
```

### Access Logs

View access logs from the logging bucket:

```bash
aws s3api list-objects-v2 \
  --bucket my-secure-bucket-001-logs-${AWS_ACCOUNT_ID} \
  --prefix access-logs/ \
  --region us-east-1
```

### Versioning Status

Check versioning and lifecycle policies:

```bash
aws s3api get-bucket-versioning \
  --bucket my-secure-bucket-001 \
  --region us-east-1

aws s3api get-bucket-lifecycle-configuration \
  --bucket my-secure-bucket-001 \
  --region us-east-1
```

## Next Steps

1. ✅ Deploy stack using one of the methods above
2. ✅ Verify bucket creation in S3 console
3. ✅ Test file upload with encryption
4. ✅ Configure bucket policies as needed
5. ✅ Set up CloudWatch alerts for bucket metrics
6. ✅ Document bucket usage in team wiki

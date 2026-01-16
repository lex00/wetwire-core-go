# S3 Bucket Template with GitLab Publishing (Beginner)

I'm new to AWS and GitLab CI/CD. Please help me create:

1. A CloudFormation template that defines an S3 bucket
2. A GitLab pipeline to validate and publish the template

## What I need for AWS

I want to create a CloudFormation template for an S3 bucket. This is the "infrastructure as code" approach where I define what I want in a YAML file.

For the bucket, I'd like:
- **Versioning** - So I can recover old versions of files
- **Encryption** - To keep my data secure
- **Private access only** - Don't want anything public

Please explain what each part of the template does.

## What I need for GitLab

I want a `.gitlab-ci.yml` file that:
1. Validates my CloudFormation template syntax is correct
2. Publishes the template file to S3 so others can use it
3. Creates a versioned release

**Important**: The pipeline should just publish the template file itself - it should NOT actually create the S3 bucket resources. That happens later when someone uses the template.

## Questions I have

- How do I connect GitLab to AWS for publishing?
- What's the difference between publishing a template vs deploying it?
- How do I version my templates?

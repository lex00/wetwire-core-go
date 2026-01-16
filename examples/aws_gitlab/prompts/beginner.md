# S3 Bucket with GitLab Deployment (Beginner)

I'm new to AWS and GitLab CI/CD. Please help me create:

1. An S3 bucket using CloudFormation
2. A GitLab pipeline to deploy it

## What I need for AWS

I want to create an S3 bucket. I've heard that CloudFormation is the "infrastructure as code" way to do this on AWS.

For the bucket, I'd like:
- **Versioning** - So I can recover old versions of files
- **Encryption** - To keep my data secure
- **Private access only** - Don't want anything public

Please explain what each part of the template does.

## What I need for GitLab

I want a `.gitlab-ci.yml` file that automatically deploys my CloudFormation template when I push code.

The pipeline should:
1. Check that my template is valid (I heard there's a validate command?)
2. Deploy the stack to AWS
3. Show me what got created

I don't know much about GitLab CI stages, so please explain the structure.

## Questions I have

- How do I connect GitLab to AWS? (I assume I need credentials?)
- What's a "stack output" and why do I need it?
- How do I know if the deployment worked?

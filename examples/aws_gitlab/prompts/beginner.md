I'm new to AWS and GitLab CI/CD. Please help me create:

For AWS I need:

1. A CloudFormation template that defines an S3 bucket
2. A GitLab pipeline to validate and publish the template

- I want to be able to recover old versions of files
- I want encryption (not sure how it works)
- It cannot have public access

Please explain what each part of the template does.


For gitlab I need a pipeline that:
1. Validates my CloudFormation template syntax is correct
2. Publishes my template file to S3 so others can use it
3. Creates a versioned release

**Important**: The pipeline should just publish the template file itself - it should NOT actually create the S3 bucket resources. That happens later when someone uses the template.

## Questions I have

- How do I connect GitLab to AWS for publishing?
- What's the difference between publishing a template vs deploying it?
- How do I version my templates?

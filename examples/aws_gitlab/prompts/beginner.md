I need a CloudFormation template for managing an S3 bucket.

- I want to be able to recover old versions of files
- I want encryption (not sure how it works)
- It cannot have public access

Please explain what each part of the template does.

I also need a GitLab pipeline to:
1. Validate my CloudFormation template syntax is correct
2. Publishe my template file to S3 so others can use it
3. Create a versioned release

**Important**: Only publish the template to s3 so that it is reusable, no execution.

## Questions I have

- How do I get the right values into my GitLab pipeline for AWS publishing?

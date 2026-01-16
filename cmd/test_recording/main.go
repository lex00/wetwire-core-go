package main

import (
	"fmt"
	"os"

	"github.com/lex00/wetwire-core-go/scenario"
)

// DemoSession implements scenario.SessionMessages for recording
type DemoSession struct {
	name     string
	messages []scenario.SessionMessage
}

func (s *DemoSession) Name() string                           { return s.name }
func (s *DemoSession) GetMessages() []scenario.SessionMessage { return s.messages }

func main() {
	outputDir := "/tmp/recordings"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Create a session with realistic conversation
	session := &DemoSession{
		name: "s3_bucket_demo",
		messages: []scenario.SessionMessage{
			{
				Role:    "developer",
				Content: "I need an S3 bucket for storing application logs",
			},
			{
				Role: "runner",
				Content: `I'll create an S3 bucket for your application logs.

Let me generate the CloudFormation template...

` + "```yaml" + `
AWSTemplateFormatVersion: '2010-09-09'
Description: S3 Bucket for Application Logs

Resources:
  LogsBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub '${AWS::StackName}-logs'
      VersioningConfiguration:
        Status: Enabled
      LifecycleConfiguration:
        Rules:
          - Id: ExpireOldLogs
            Status: Enabled
            ExpirationInDays: 90
      PublicAccessBlockConfiguration:
        BlockPublicAcls: true
        BlockPublicPolicy: true
        IgnorePublicAcls: true
        RestrictPublicBuckets: true

Outputs:
  BucketName:
    Value: !Ref LogsBucket
    Export:
      Name: !Sub '${AWS::StackName}-bucket-name'
  BucketArn:
    Value: !GetAtt LogsBucket.Arn
` + "```" + `

Created storage.yaml with:
- S3 bucket with versioning enabled
- Lifecycle rule to expire logs after 90 days
- Public access blocked for security
- Exported bucket name and ARN for other stacks`,
			},
			{
				Role:    "developer",
				Content: "Can you add encryption?",
			},
			{
				Role: "runner",
				Content: `Adding SSE-S3 encryption to the bucket...

` + "```yaml" + `
  LogsBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketEncryption:
        ServerSideEncryptionConfiguration:
          - ServerSideEncryptionByDefault:
              SSEAlgorithm: AES256
` + "```" + `

Updated! The bucket now has server-side encryption with AES-256.`,
			},
		},
	}

	err := scenario.RecordSession(session, scenario.SessionRecordOptions{
		OutputDir: outputDir,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== SVG saved to: %s/%s.svg ===\n", outputDir, session.Name())
}

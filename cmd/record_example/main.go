package main

import (
	"fmt"
	"os"

	"github.com/lex00/wetwire-core-go/scenario"
)

// ExampleSession implements scenario.SessionMessages
type ExampleSession struct {
	name     string
	messages []scenario.SessionMessage
}

func (s *ExampleSession) Name() string                           { return s.name }
func (s *ExampleSession) GetMessages() []scenario.SessionMessage { return s.messages }

func main() {
	outputDir := "./examples/aws_gitlab/recordings"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Realistic multi-domain conversation
	session := &ExampleSession{
		name: "aws_gitlab_demo",
		messages: []scenario.SessionMessage{
			{
				Role:    "developer",
				Content: "I need AWS infrastructure with a GitLab CI/CD pipeline for deployment",
			},
			{
				Role: "runner",
				Content: `I'll create AWS infrastructure and a GitLab pipeline.

Let me start with the VPC template...

` + "```yaml" + `
AWSTemplateFormatVersion: '2010-09-09'
Description: Multi-AZ VPC for EKS

Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsHostnames: true

  PublicSubnetA:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      AvailabilityZone: !Select [0, !GetAZs '']
      CidrBlock: 10.0.1.0/24
` + "```" + `

Created vpc.yaml with public/private subnets.`,
			},
			{
				Role:    "developer",
				Content: "Add EKS cluster and RDS database too",
			},
			{
				Role: "runner",
				Content: `Creating EKS and RDS templates...

` + "```yaml" + `
# eks.yaml
Resources:
  EKSCluster:
    Type: AWS::EKS::Cluster
    Properties:
      Name: !Sub '${AWS::StackName}-cluster'
      ResourcesVpcConfig:
        SubnetIds:
          - !ImportValue vpc-private-subnet-a
          - !ImportValue vpc-private-subnet-b

# rds.yaml
Resources:
  Database:
    Type: AWS::RDS::DBInstance
    Properties:
      Engine: postgres
      DBInstanceClass: db.t3.medium
      VPCSecurityGroups:
        - !ImportValue vpc-db-security-group
` + "```" + `

Now creating the GitLab pipeline...

` + "```yaml" + `
# .gitlab-ci.yml
stages:
  - validate
  - deploy
  - test

validate:
  stage: validate
  script:
    - aws cloudformation validate-template --template-body file://cfn-templates/vpc.yaml

deploy-vpc:
  stage: deploy
  script:
    - aws cloudformation deploy --template-file cfn-templates/vpc.yaml --stack-name vpc
  when: manual
` + "```" + `

Done! Created:
- cfn-templates/vpc.yaml
- cfn-templates/eks.yaml
- cfn-templates/rds.yaml
- .gitlab-ci.yml`,
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

	fmt.Printf("Recording saved to: %s/%s.svg\n", outputDir, session.Name())
}

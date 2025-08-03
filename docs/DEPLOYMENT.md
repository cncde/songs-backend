# Songs Backend API - Deployment Guide

This guide provides comprehensive instructions for deploying the Songs Backend
API to AWS. The application uses a containerized Lambda deployment model with
DynamoDB, S3, and Cognito integration.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Infrastructure Setup](#infrastructure-setup)
- [AWS Service Configuration](#aws-service-configuration)
- [Environment Variables](#environment-variables)
- [Container Build & Deployment](#container-build--deployment)
- [Monitoring & Logging](#monitoring--logging)
- [Production Considerations](#production-considerations)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

```bash
# AWS CLI v2
aws --version
# Should return: aws-cli/2.x.x or higher

# Docker
docker --version
# Should return: Docker version 20.x.x or higher

# Go 1.22+
go version
# Should return: go version go1.22.x

# Optional but recommended
terraform --version  # For Infrastructure as Code
```

### AWS Permissions

Ensure your AWS credentials have permissions for:

- Lambda (create, update, deploy functions)
- DynamoDB (create tables, manage items)
- S3 (create buckets, manage objects)
- Cognito (create user pools)
- IAM (create roles and policies)
- ECR (create repositories, push images)
- CloudWatch (create log groups, put metrics)

## Infrastructure Setup

### 1. Create AWS Resources

#### Option A: Using AWS CLI

```bash
# Set environment variables
export ENVIRONMENT="production"  # or "staging", "development"
export AWS_REGION="us-east-1"
export PROJECT_NAME="songs-backend"

# Create DynamoDB Tables
aws dynamodb create-table \
    --table-name "songs-${ENVIRONMENT}" \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=stage,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=UserIndex,KeySchema="[{AttributeName=user_id,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}]",Projection="{ProjectionType=ALL}",\
            ProvisionedThroughput="{ReadCapacityUnits=5,WriteCapacityUnits=5}" \
        IndexName=StageIndex,KeySchema="[{AttributeName=stage,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}]",Projection="{ProjectionType=ALL}",\
            ProvisionedThroughput="{ReadCapacityUnits=5,WriteCapacityUnits=5}" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region $AWS_REGION

aws dynamodb create-table \
    --table-name "comments-${ENVIRONMENT}" \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=song_id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=audio_url,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=SongIndex,KeySchema="[{AttributeName=song_id,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}]",Projection="{ProjectionType=ALL}",\
            ProvisionedThroughput="{ReadCapacityUnits=5,WriteCapacityUnits=5}" \
        IndexName=UserIndex,KeySchema="[{AttributeName=user_id,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}]",Projection="{ProjectionType=ALL}",\
            ProvisionedThroughput="{ReadCapacityUnits=5,WriteCapacityUnits=5}" \
        IndexName=AudioIndex,KeySchema="[{AttributeName=audio_url,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}]",Projection="{ProjectionType=ALL}",\
            ProvisionedThroughput="{ReadCapacityUnits=5,WriteCapacityUnits=5}" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region $AWS_REGION

# Create S3 Bucket
aws s3 mb "s3://songs-media-${ENVIRONMENT}" --region $AWS_REGION

# Configure S3 bucket CORS
cat > cors-config.json << EOF
{
    "CORSRules": [
        {
            "AllowedHeaders": ["*"],
            "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
            "AllowedOrigins": ["*"],
            "ExposeHeaders": ["ETag"]
        }
    ]
}
EOF

aws s3api put-bucket-cors \
    --bucket "songs-media-${ENVIRONMENT}" \
    --cors-configuration file://cors-config.json

# Enable S3 bucket versioning
aws s3api put-bucket-versioning \
    --bucket "songs-media-${ENVIRONMENT}" \
    --versioning-configuration Status=Enabled
```

#### Option B: Using CloudFormation Template

Create `infrastructure/cloudformation.yaml`:

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Songs Backend API Infrastructure'

Parameters:
  Environment:
    Type: String
    Default: production
    AllowedValues: [development, staging, production]
  
  ProjectName:
    Type: String
    Default: songs-backend

Resources:
  # DynamoDB Tables
  SongsTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: !Sub "${ProjectName}-songs-${Environment}"
      BillingMode: PAY_PER_REQUEST
      AttributeDefinitions:
        - AttributeName: id
          AttributeType: S
        - AttributeName: user_id
          AttributeType: S
        - AttributeName: stage
          AttributeType: S
        - AttributeName: created_at
          AttributeType: S
      KeySchema:
        - AttributeName: id
          KeyType: HASH
      GlobalSecondaryIndexes:
        - IndexName: UserIndex
          KeySchema:
            - AttributeName: user_id
              KeyType: HASH
            - AttributeName: created_at
              KeyType: RANGE
          Projection:
            ProjectionType: ALL
        - IndexName: StageIndex
          KeySchema:
            - AttributeName: stage
              KeyType: HASH
            - AttributeName: created_at
              KeyType: RANGE
          Projection:
            ProjectionType: ALL
      StreamSpecification:
        StreamViewType: NEW_AND_OLD_IMAGES
      PointInTimeRecoverySpecification:
        PointInTimeRecoveryEnabled: true

  CommentsTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: !Sub "${ProjectName}-comments-${Environment}"
      BillingMode: PAY_PER_REQUEST
      AttributeDefinitions:
        - AttributeName: id
          AttributeType: S
        - AttributeName: song_id
          AttributeType: S
        - AttributeName: user_id
          AttributeType: S
        - AttributeName: audio_url
          AttributeType: S
        - AttributeName: created_at
          AttributeType: S
      KeySchema:
        - AttributeName: id
          KeyType: HASH
      GlobalSecondaryIndexes:
        - IndexName: SongIndex
          KeySchema:
            - AttributeName: song_id
              KeyType: HASH
            - AttributeName: created_at
              KeyType: RANGE
          Projection:
            ProjectionType: ALL
        - IndexName: UserIndex
          KeySchema:
            - AttributeName: user_id
              KeyType: HASH
            - AttributeName: created_at
              KeyType: RANGE
          Projection:
            ProjectionType: ALL
        - IndexName: AudioIndex
          KeySchema:
            - AttributeName: audio_url
              KeyType: HASH
            - AttributeName: created_at
              KeyType: RANGE
          Projection:
            ProjectionType: ALL
      StreamSpecification:
        StreamViewType: NEW_AND_OLD_IMAGES
      PointInTimeRecoverySpecification:
        PointInTimeRecoveryEnabled: true

  # S3 Bucket
  MediaBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub "${ProjectName}-media-${Environment}"
      CorsConfiguration:
        CorsRules:
          - AllowedHeaders: ['*']
            AllowedMethods: [GET, PUT, POST, DELETE]
            AllowedOrigins: ['*']
            ExposedHeaders: [ETag]
      VersioningConfiguration:
        Status: Enabled
      LifecycleConfiguration:
        Rules:
          - Id: DeleteIncompleteMultipartUploads
            Status: Enabled
            AbortIncompleteMultipartUpload:
              DaysAfterInitiation: 1
      PublicAccessBlockConfiguration:
        BlockPublicAcls: true
        BlockPublicPolicy: true
        IgnorePublicAcls: true
        RestrictPublicBuckets: true

  # ECR Repository
  ECRRepository:
    Type: AWS::ECR::Repository
    Properties:
      RepositoryName: !Sub "${ProjectName}-${Environment}"
      ImageScanningConfiguration:
        ScanOnPush: true
      LifecyclePolicy:
        LifecyclePolicyText: |
          {
            "rules": [
              {
                "rulePriority": 1,
                "selection": {
                  "tagStatus": "untagged",
                  "countType": "sinceImagePushed",
                  "countUnit": "days",
                  "countNumber": 7
                },
                "action": {
                  "type": "expire"
                }
              }
            ]
          }

  # Lambda Execution Role
  LambdaExecutionRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: !Sub "${ProjectName}-lambda-execution-role-${Environment}"
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
      Policies:
        - PolicyName: DynamoDBAccess
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - dynamodb:GetItem
                  - dynamodb:PutItem
                  - dynamodb:UpdateItem
                  - dynamodb:DeleteItem
                  - dynamodb:Query
                  - dynamodb:Scan
                  - dynamodb:BatchGetItem
                  - dynamodb:BatchWriteItem
                Resource:
                  - !GetAtt SongsTable.Arn
                  - !Sub "${SongsTable.Arn}/index/*"
                  - !GetAtt CommentsTable.Arn
                  - !Sub "${CommentsTable.Arn}/index/*"
        - PolicyName: S3Access
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - s3:GetObject
                  - s3:PutObject
                  - s3:DeleteObject
                  - s3:ListBucket
                  - s3:PutObjectAcl
                  - s3:GetObjectAcl
                Resource:
                  - !GetAtt MediaBucket.Arn
                  - !Sub "${MediaBucket.Arn}/*"

Outputs:
  SongsTableName:
    Description: DynamoDB Songs Table Name
    Value: !Ref SongsTable
    Export:
      Name: !Sub "${ProjectName}-songs-table-${Environment}"
  
  CommentsTableName:
    Description: DynamoDB Comments Table Name
    Value: !Ref CommentsTable
    Export:
      Name: !Sub "${ProjectName}-comments-table-${Environment}"
  
  MediaBucketName:
    Description: S3 Media Bucket Name
    Value: !Ref MediaBucket
    Export:
      Name: !Sub "${ProjectName}-media-bucket-${Environment}"
  
  ECRRepositoryURI:
    Description: ECR Repository URI
    Value: !GetAtt ECRRepository.RepositoryUri
    Export:
      Name: !Sub "${ProjectName}-ecr-repository-${Environment}"
  
  LambdaExecutionRoleArn:
    Description: Lambda Execution Role ARN
    Value: !GetAtt LambdaExecutionRole.Arn
    Export:
      Name: !Sub "${ProjectName}-lambda-role-${Environment}"
```

Deploy the CloudFormation stack:

```bash
aws cloudformation deploy \
    --template-file infrastructure/cloudformation.yaml \
    --stack-name "songs-backend-${ENVIRONMENT}" \
    --parameter-overrides Environment=$ENVIRONMENT \
    --capabilities CAPABILITY_NAMED_IAM \
    --region $AWS_REGION
```

### 2. AWS Cognito User Pool Setup

```bash
# Create Cognito User Pool
aws cognito-idp create-user-pool \
    --pool-name "songs-backend-${ENVIRONMENT}" \
    --policies '{
        "PasswordPolicy": {
            "MinimumLength": 8,
            "RequireUppercase": true,
            "RequireLowercase": true,
            "RequireNumbers": true,
            "RequireSymbols": false
        }
    }' \
    --auto-verified-attributes email \
    --username-attributes email \
    --schema '[
        {
            "Name": "email",
            "AttributeDataType": "String",
            "Required": true,
            "Mutable": true
        },
        {
            "Name": "name",
            "AttributeDataType": "String",
            "Required": false,
            "Mutable": true
        }
    ]' \
    --region $AWS_REGION

# Note the UserPoolId from the response, then create a client
export USER_POOL_ID="us-east-1_XXXXXXXXX"  # Replace with actual ID

aws cognito-idp create-user-pool-client \
    --user-pool-id $USER_POOL_ID \
    --client-name "songs-backend-client-${ENVIRONMENT}" \
    --generate-secret \
    --explicit-auth-flows ALLOW_ADMIN_USER_PASSWORD_AUTH \
        ALLOW_USER_SRP_AUTH ALLOW_REFRESH_TOKEN_AUTH \
    --region $AWS_REGION
```

## AWS Service Configuration

### DynamoDB Table Setup

After creating tables, enable auto-scaling (recommended for production):

```bash
# Enable auto-scaling for Songs table
aws application-autoscaling register-scalable-target \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:ReadCapacityUnits \
    --resource-id "table/songs-${ENVIRONMENT}" \
    --min-capacity 5 \
    --max-capacity 100

aws application-autoscaling register-scalable-target \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:WriteCapacityUnits \
    --resource-id "table/songs-${ENVIRONMENT}" \
    --min-capacity 5 \
    --max-capacity 100

# Create scaling policies
aws application-autoscaling put-scaling-policy \
    --service-namespace dynamodb \
    --scalable-dimension dynamodb:table:ReadCapacityUnits \
    --resource-id "table/songs-${ENVIRONMENT}" \
    --policy-name "songs-read-scaling-policy" \
    --policy-type TargetTrackingScaling \
    --target-tracking-scaling-policy-configuration '{
        "TargetValue": 70.0,
        "ScaleInCooldown": 60,
        "ScaleOutCooldown": 60,
        "PredefinedMetricSpecification": {
            "PredefinedMetricType": "DynamoDBReadCapacityUtilization"
        }
    }'
```

### S3 Bucket Policies

Create a bucket policy for secure access:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowLambdaAccess",
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::ACCOUNT-ID:role/songs-backend-lambda-execution-role-production"
            },
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::songs-media-production",
                "arn:aws:s3:::songs-media-production/*"
            ]
        }
    ]
}
```

Apply the policy:

```bash
aws s3api put-bucket-policy \
    --bucket "songs-media-${ENVIRONMENT}" \
    --policy file://bucket-policy.json
```

## Environment Variables

### Required Environment Variables

Create a `.env.production` file (or use AWS Systems Manager Parameter Store):

```bash
# Application Settings
ENVIRONMENT=production
AWS_REGION=us-east-1
API_VERSION=v1
LOG_LEVEL=info
ENABLE_METRICS=true

# DynamoDB Configuration
SONGS_TABLE=songs-production
COMMENTS_TABLE=comments-production

# S3 Configuration
S3_BUCKET=songs-media-production

# Cognito Configuration
COGNITO_USER_POOL_ID=us-east-1_XXXXXXXXX
COGNITO_REGION=us-east-1
```

### Using AWS Systems Manager Parameter Store

Store sensitive configuration in AWS Systems Manager:

```bash
# Store configuration parameters
aws ssm put-parameter \
    --name "/songs-backend/production/cognito-user-pool-id" \
    --value "us-east-1_XXXXXXXXX" \
    --type SecureString \
    --region $AWS_REGION

aws ssm put-parameter \
    --name "/songs-backend/production/songs-table" \
    --value "songs-production" \
    --type String \
    --region $AWS_REGION

aws ssm put-parameter \
    --name "/songs-backend/production/comments-table" \
    --value "comments-production" \
    --type String \
    --region $AWS_REGION

aws ssm put-parameter \
    --name "/songs-backend/production/s3-bucket" \
    --value "songs-media-production" \
    --type String \
    --region $AWS_REGION
```

Update your Lambda function to read from Parameter Store:

```go
package config

import (
    "context"
    "fmt"
    
    "github.com/aws/aws-sdk-go-v2/service/ssm"
)

func LoadFromParameterStore(ssmClient *ssm.Client, \
    environment string) (*Config, error) {
    prefix := fmt.Sprintf("/songs-backend/%s/", environment)
    
    // Get parameters
    input := &ssm.GetParametersByPathInput{
        Path:           &prefix,
        Recursive:      true,
        WithDecryption: true,
    }
    
    result, err := ssmClient.GetParametersByPath(context.TODO(), input)
    if err != nil {
        return nil, err
    }
    
    // Parse parameters into config
    cfg := &Config{}
    for _, param := range result.Parameters {
        key := strings.TrimPrefix(*param.Name, prefix)
        value := *param.Value
        
        switch key {
        case "cognito-user-pool-id":
            cfg.CognitoUserPoolID = value
        case "songs-table":
            cfg.SongsTable = value
        // ... handle other parameters
        }
    }
    
    return cfg, nil
}
```

## Container Build & Deployment

### 1. Build the Container Image

```bash
# Build the Docker image
docker build -t songs-backend:latest .

# Get AWS account ID
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

# Create ECR repository (if not created via CloudFormation)
aws ecr create-repository \
    --repository-name songs-backend-${ENVIRONMENT} \
    --region $AWS_REGION

# Get ECR login token
aws ecr get-login-password --region $AWS_REGION | \
    docker login --username AWS --password-stdin $ECR_REGISTRY

# Tag the image for ECR
docker tag songs-backend:latest \
    $ECR_REGISTRY/songs-backend-${ENVIRONMENT}:latest

# Push the image to ECR
docker push $ECR_REGISTRY/songs-backend-${ENVIRONMENT}:latest
```

### 2. Create Lambda Function

```bash
# Create Lambda function with container image
aws lambda create-function \
    --function-name "songs-backend-${ENVIRONMENT}" \
    --role "arn:aws:iam::${AWS_ACCOUNT_ID}:role/\
        songs-backend-lambda-execution-role-${ENVIRONMENT}" \
    --code ImageUri="${ECR_REGISTRY}/songs-backend-${ENVIRONMENT}:latest" \
    --package-type Image \
    --timeout 30 \
    --memory-size 512 \
    --environment "Variables={
        ENVIRONMENT=${ENVIRONMENT},
        AWS_REGION=${AWS_REGION},
        SONGS_TABLE=songs-${ENVIRONMENT},
        COMMENTS_TABLE=comments-${ENVIRONMENT},
        S3_BUCKET=songs-media-${ENVIRONMENT},
        COGNITO_USER_POOL_ID=${USER_POOL_ID},
        API_VERSION=v1,
        LOG_LEVEL=info,
        ENABLE_METRICS=true
    }" \
    --region $AWS_REGION

# Configure Function URL (optional - for direct HTTPS access)
aws lambda create-function-url-config \
    --function-name "songs-backend-${ENVIRONMENT}" \
    --auth-type NONE \
    --cors '{
        "AllowCredentials": false,
        "AllowHeaders": ["content-type", "authorization"],
        "AllowMethods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        "AllowOrigins": ["*"],
        "ExposeHeaders": ["date", "keep-alive"],
        "MaxAge": 86400
    }' \
    --region $AWS_REGION
```

### 3. Update Lambda Function (for deployments)

```bash
# Update function code with new image
aws lambda update-function-code \
    --function-name "songs-backend-${ENVIRONMENT}" \
    --image-uri "${ECR_REGISTRY}/songs-backend-${ENVIRONMENT}:latest" \
    --region $AWS_REGION

# Wait for update to complete
aws lambda wait function-updated \
    --function-name "songs-backend-${ENVIRONMENT}" \
    --region $AWS_REGION

# Update environment variables (if needed)
aws lambda update-function-configuration \
    --function-name "songs-backend-${ENVIRONMENT}" \
    --environment "Variables={
        ENVIRONMENT=${ENVIRONMENT},
        AWS_REGION=${AWS_REGION},
        SONGS_TABLE=songs-${ENVIRONMENT},
        COMMENTS_TABLE=comments-${ENVIRONMENT},
        S3_BUCKET=songs-media-${ENVIRONMENT},
        COGNITO_USER_POOL_ID=${USER_POOL_ID},
        API_VERSION=v1,
        LOG_LEVEL=info,
        ENABLE_METRICS=true
    }" \
    --region $AWS_REGION
```

### 4. Deployment Automation Script

Create `scripts/deploy.sh`:

```bash
#!/bin/bash
set -e

# Configuration
ENVIRONMENT=${1:-production}
AWS_REGION=${AWS_REGION:-us-east-1}
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
FUNCTION_NAME="songs-backend-${ENVIRONMENT}"

echo "Deploying Songs Backend API to ${ENVIRONMENT}..."

# Build and push container image
echo "Building container image..."
docker build -t songs-backend:latest .

echo "Pushing to ECR..."
aws ecr get-login-password --region $AWS_REGION | \
    docker login --username AWS --password-stdin $ECR_REGISTRY

docker tag songs-backend:latest $ECR_REGISTRY/songs-backend-${ENVIRONMENT}:latest
docker push $ECR_REGISTRY/songs-backend-${ENVIRONMENT}:latest

# Update Lambda function
echo "Updating Lambda function..."
aws lambda update-function-code \
    --function-name $FUNCTION_NAME \
    --image-uri $ECR_REGISTRY/songs-backend-${ENVIRONMENT}:latest \
    --region $AWS_REGION

# Wait for deployment to complete
echo "Waiting for deployment to complete..."
aws lambda wait function-updated \
    --function-name $FUNCTION_NAME \
    --region $AWS_REGION

# Run health check
echo "Running health check..."
FUNCTION_URL=$(aws lambda get-function-url-config \
    --function-name $FUNCTION_NAME \
    --query FunctionUrl \
    --output text \
    --region $AWS_REGION 2>/dev/null || echo "")

if [ -n "$FUNCTION_URL" ]; then
    HEALTH_RESPONSE=$(curl -s "${FUNCTION_URL}health" || echo "")
    if echo "$HEALTH_RESPONSE" | grep -q '"status":"healthy"'; then
        echo "✅ Deployment successful! Health check passed."
        echo "Function URL: $FUNCTION_URL"
    else
        echo "❌ Health check failed. Response: $HEALTH_RESPONSE"
        exit 1
    fi
else
    echo "✅ Deployment completed. Function URL not configured."
fi

echo "Deployment to ${ENVIRONMENT} completed successfully!"
```

Make it executable and run:

```bash
chmod +x scripts/deploy.sh
./scripts/deploy.sh production
```

## Monitoring & Logging

### 1. CloudWatch Log Groups

```bash
# Create log group for Lambda function
aws logs create-log-group \
    --log-group-name "/aws/lambda/songs-backend-${ENVIRONMENT}" \
    --region $AWS_REGION

# Set retention period (30 days for production)
aws logs put-retention-policy \
    --log-group-name "/aws/lambda/songs-backend-${ENVIRONMENT}" \
    --retention-in-days 30 \
    --region $AWS_REGION
```

### 2. CloudWatch Alarms

```bash
# High error rate alarm
aws cloudwatch put-metric-alarm \
    --alarm-name "songs-backend-${ENVIRONMENT}-high-error-rate" \
    --alarm-description "High error rate for Songs Backend API" \
    --actions-enabled \
    --metric-name Errors \
    --namespace AWS/Lambda \
    --statistic Sum \
    --dimensions Name=FunctionName,Value=songs-backend-${ENVIRONMENT} \
    --period 300 \
    --evaluation-periods 2 \
    --threshold 10 \
    --comparison-operator GreaterThanThreshold \
    --region $AWS_REGION

# High duration alarm
aws cloudwatch put-metric-alarm \
    --alarm-name "songs-backend-${ENVIRONMENT}-high-duration" \
    --alarm-description "High duration for Songs Backend API" \
    --actions-enabled \
    --metric-name Duration \
    --namespace AWS/Lambda \
    --statistic Average \
    --dimensions Name=FunctionName,Value=songs-backend-${ENVIRONMENT} \
    --period 300 \
    --evaluation-periods 2 \
    --threshold 10000 \
    --comparison-operator GreaterThanThreshold \
    --region $AWS_REGION
```

### 3. Custom Metrics Dashboard

Create `monitoring/dashboard.json`:

```json
{
    "widgets": [
        {
            "type": "metric",
            "properties": {
                "metrics": [
                    ["AWS/Lambda", "Invocations", "FunctionName", "songs-backend-production"],
                    [".", "Errors", ".", "."],
                    [".", "Duration", ".", "."]
                ],
                "period": 300,
                "stat": "Sum",
                "region": "us-east-1",
                "title": "Lambda Metrics"
            }
        },
        {
            "type": "metric",
            "properties": {
                "metrics": [
                    ["AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "songs-production"],
                    [".", "ConsumedWriteCapacityUnits", ".", "."]
                ],
                "period": 300,
                "stat": "Sum",
                "region": "us-east-1",
                "title": "DynamoDB Capacity"
            }
        }
    ]
}
```

## Production Considerations

### 1. Security Hardening

```bash
# Enable GuardDuty for threat detection
aws guardduty create-detector \
    --enable \
    --region $AWS_REGION

# Enable CloudTrail for audit logging
aws cloudtrail create-trail \
    --name songs-backend-audit-trail \
    --s3-bucket-name songs-audit-logs-${ENVIRONMENT} \
    --include-global-service-events \
    --is-multi-region-trail \
    --region $AWS_REGION
```

### 2. Backup Strategy

```bash
# Enable DynamoDB Point-in-Time Recovery (if not in CloudFormation)
aws dynamodb update-continuous-backups \
    --table-name songs-${ENVIRONMENT} \
    --point-in-time-recovery-specification PointInTimeRecoveryEnabled=true \
    --region $AWS_REGION

aws dynamodb update-continuous-backups \
    --table-name comments-${ENVIRONMENT} \
    --point-in-time-recovery-specification PointInTimeRecoveryEnabled=true \
    --region $AWS_REGION

# S3 versioning is already enabled in CloudFormation template
# Consider adding lifecycle policies for cost optimization
```

### 3. Performance Optimization

```bash
# Configure Lambda reserved concurrency (prevents cold starts)
aws lambda put-reserved-concurrency-configuration \
    --function-name songs-backend-${ENVIRONMENT} \
    --reserved-concurrent-executions 50 \
    --region $AWS_REGION

# Configure DynamoDB auto-scaling (shown earlier)
# Consider using DynamoDB Global Tables for multi-region deployments
```

### 4. Cost Optimization

```yaml
# S3 lifecycle policy for old files
LifecycleConfiguration:
  Rules:
    - Id: TransitionToIA
      Status: Enabled
      Transitions:
        - Days: 30
          StorageClass: STANDARD_IA
        - Days: 90
          StorageClass: GLACIER
    - Id: DeleteOldVersions
      Status: Enabled
      NoncurrentVersionExpiration:
        NoncurrentDays: 90
```

## Troubleshooting

### Common Issues

1. **Lambda Cold Start Times**

```bash
# Solution: Use provisioned concurrency
aws lambda put-provisioned-concurrency-config \
    --function-name songs-backend-${ENVIRONMENT} \
    --qualifier '$LATEST' \
    --provisioned-concurrency-units 2
```

1. **DynamoDB Throttling**

```bash
# Check throttling metrics
aws cloudwatch get-metric-statistics \
    --namespace AWS/DynamoDB \
    --metric-name ThrottledRequests \
    --dimensions Name=TableName,Value=songs-${ENVIRONMENT} \
    --statistics Sum \
    --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 300
```

1. **JWT Validation Failures**

```bash
# Check Cognito configuration
aws cognito-idp describe-user-pool \
    --user-pool-id $USER_POOL_ID
    
# Verify JWKS URL accessibility
curl "https://cognito-idp.${AWS_REGION}.amazonaws.com/${USER_POOL_ID}/.well-known/jwks.json"
```

### Debugging Lambda Function

```bash
# View recent logs
aws logs tail "/aws/lambda/songs-backend-${ENVIRONMENT}" \
    --since 1h \
    --follow

# Check function configuration
aws lambda get-function-configuration \
    --function-name songs-backend-${ENVIRONMENT}

# Test function with sample event
aws lambda invoke \
    --function-name songs-backend-${ENVIRONMENT} \
    --payload '{"httpMethod":"GET","path":"/health","headers":{}}' \
    --region $AWS_REGION \
    response.json

cat response.json
```

### Health Check Endpoint

The API includes a health check endpoint at `/health`:

```bash
# Test health endpoint
curl https://your-lambda-function-url.lambda-url.us-east-1.on.aws/health

# Expected response:
# {
#   "status": "healthy",
#   "service": "songs-api",
#   "version": "v1"
# }
```

This deployment guide provides a comprehensive approach to deploying the Songs
Backend API to AWS with proper monitoring, security, and production
considerations. Adapt the configurations based on your specific requirements
and AWS account structure.

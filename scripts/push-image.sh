#!/usr/bin/env sh

set -e

AWS_DEFAULT_REGION=eu-west-1

# Get the app version
echo "Getting the app version"
APP_VERSION=$(go run main.go -version)

# Get the AWS account ID
echo "Getting the AWS account ID"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

REPOSITORY_URL="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_DEFAULT_REGION}.amazonaws.com/dumbbell"

# Login to AWS ECR
echo "Logging into AWS ECR"
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${REPOSITORY_URL}:${APP_VERSION}

# Build the docker image
echo "Building the docker image"
docker build --platform linux/amd64 -t dumbbell .

# Tag the docker image
echo "Tagging the docker image"
docker tag dumbbell:latest ${REPOSITORY_URL}:${APP_VERSION}

# Push the docker image
echo "Pushing the docker image"
docker push ${REPOSITORY_URL}:${APP_VERSION}
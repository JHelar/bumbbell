variable "AWS_REGION" {
  description = "The AWS region to deploy the infrastructure"
  default     = "eu-west-1"
}

variable "AWS_REPOSITORY_NAME" {
  description = "The name of the ECR repository"
  default     = "dumbbell"
}

variable "APP_PORT" {
  description = "The port the application will run on"
  default     = 8080
}

variable "APP_ENVIRONMENT" {
  description = "The environment the application will run in"
  default     = "production"
}

variable "APP_VERSION" {
  description = "The application version"
}

provider "aws" {
  region = var.AWS_REGION
}

resource "aws_ecr_repository" "dumbbell_docker_repository" {
  name = var.AWS_REPOSITORY_NAME
}

resource "aws_security_group" "ssh_http_https_access" {
  name        = "ssh-http-https-access"
  description = "Allow SSH, HTTP, and HTTPS inbound traffic"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Allow SSH access from anywhere (not recommended for production)
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Allow HTTP access from anywhere (not recommended for production)
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Allow HTTPS access from anywhere (not recommended for production)
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "ecr_instance_role" {
  name = "ecr-instance-role"
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : [
          "sts:AssumeRole"
        ],
        "Principal" : {
          "Service" : [
            "ec2.amazonaws.com"
          ]
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecr_policy_attachment" {
  role       = aws_iam_role.ecr_instance_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAppRunnerServicePolicyForECRAccess"
}

resource "aws_iam_instance_profile" "ecr_instance_profile" {
  name = "ecr-instance-profile"
  role = aws_iam_role.ecr_instance_role.name
}

resource "aws_instance" "dumbbell_web_server" {
  ami                  = "ami-0766b4b472db7e3b9"
  instance_type        = "t2.micro"
  key_name             = "Dumbbell - SSH"
  security_groups      = [aws_security_group.ssh_http_https_access.name]
  iam_instance_profile = aws_iam_instance_profile.ecr_instance_profile.name
  user_data            = <<-EOF
                        Content-Type: multipart/mixed; boundary="//"
                        MIME-Version: 1.0

                        --//
                        Content-Type: text/cloud-config; charset="us-ascii"
                        MIME-Version: 1.0
                        Content-Transfer-Encoding: 7bit
                        Content-Disposition: attachment; filename="cloud-config.txt"

                        #cloud-config
                        cloud_final_modules:
                        - [scripts-user, always]

                        --//
                        Content-Type: text/x-shellscript; charset="us-ascii"
                        MIME-Version: 1.0
                        Content-Transfer-Encoding: 7bit
                        Content-Disposition: attachment; filename="userdata.txt"

                        #!/bin/bash

                        set -e

                        echo "Install docker"
                        sudo yum update -y
                        sudo yum install -y docker
                        sudo service docker start

                        # Get ECR login password
                        echo "Get ECR login password"
                        AWS_ECR_PASSWORD=$(aws ecr get-login-password --region ${var.AWS_REGION})
                        
                        # Login docker to ECR
                        echo "Login docker to ECR"
                        sudo docker login --username AWS --password $AWS_ECR_PASSWORD ${aws_ecr_repository.dumbbell_docker_repository.repository_url}:${var.APP_VERSION}

                        # Pull Docker image from ECR
                        echo "Pull Docker image from ECR"
                        sudo docker pull ${aws_ecr_repository.dumbbell_docker_repository.repository_url}:${var.APP_VERSION}

                        # Stop and remove existing Docker container
                        echo "Stop and remove existing Docker container"
                        sudo docker stop dumbbell || true && sudo docker rm dumbbell || true

                        # Run Docker image
                        echo "Run Docker image"
                        sudo docker run --name dumbbell -p 80:${var.APP_PORT} -p 443:${var.APP_PORT} -e PORT=${var.APP_PORT} -e ENVIRONMENT=${var.APP_ENVIRONMENT} -d ${aws_ecr_repository.dumbbell_docker_repository.repository_url}:${var.APP_VERSION}
                        --//--
                    EOF


  tags = {
    Name = "dumbbell - Web Server"
  }
}

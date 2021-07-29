#!/bin/bash

# Start local test environment
docker-compose up -d

# Start sam local api
AWS_REGION=us-west-2 sam local start-api --docker-network lambda-local

# Create need tables
aws dynamodb create-table \
    --table-name ServerBoi-Server-List \
    --attribute-definitions \
        AttributeName=ServerID,AttributeType=S \
    --key-schema \
        AttributeName=ServerID,KeyType=HASH \
    --provisioned-throughput \
            ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000

aws dynamodb create-table \
    --table-name AWS-User-List \
    --attribute-definitions \
        AttributeName=UserID,AttributeType=S \
    --key-schema \
        AttributeName=UserID,KeyType=HASH \
    --provisioned-throughput \
            ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000

# Create Test Entries
aws dynamodb put-item \
    --table-name AWS-User-List \
    --item \
        '{"UserID": {"S": "155875705417236480"}, "AWSAccountID": {"S": "742762521158"}}' \
    --endpoint-url http://localhost:8000

# aws dynamodb put-item \
#     --table-name ServerBoi-Server-List \
#     --item \
#         '{"ServerID":{"S":"Test"},"Service":{"M":{"AccountID":{"S":"742762521158"},"InstanceID":{"S":"i-0933d3d5b3f92fc02"},"Name":{"S":"aws"},"Region":{"S":"us-west-2"}}}}' \
#     --endpoint-url http://localhost:8000



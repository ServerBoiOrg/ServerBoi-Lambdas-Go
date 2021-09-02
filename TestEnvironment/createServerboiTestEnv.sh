#!/bin/bash

# Start local test environment
docker-compose up -d

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

aws dynamodb create-table \
    --table-name Linode-User-List \
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
        '{"UserID": {"S": "0001"}, "AWSAccountID": {"S": "000000000000"}}' \
    --endpoint-url http://localhost:8000

aws dynamodb put-item \
    --table-name Linode-User-List \
    --item \
        '{"UserID": {"S": "0001"}, "AWSAccountID": {"S": "000000000000"}}' \
    --endpoint-url http://localhost:8000



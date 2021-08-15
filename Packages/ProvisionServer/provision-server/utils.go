package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func queryAWSAccountID(dynamo *dynamodb.Client, userID string) string {
	table := getEnvVar("AWS_TABLE")

	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]dynamotypes.AttributeValue{
			"UserID": &dynamotypes.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		log.Fatalf("Error retrieving item from dynamo: %v", err)
		panic(err)
	}
	var awsResponse AWSTableResponse
	err = attributevalue.UnmarshalMap(response.Item, &awsResponse)

	return awsResponse.AWSAccountID
}

func getCloudwatchClient() *cloudwatch.Client {
	cfg := getConfig()
	log.Printf("Getting cloudwatch client")
	cw := cloudwatch.NewFromConfig(cfg, func(options *cloudwatch.Options) {})

	return cw
}

func getS3Client() *s3.Client {
	cfg := getConfig()
	log.Printf("Getting cloudwatch client")
	s3 := s3.NewFromConfig(cfg)

	return s3
}

func getDynamo() *dynamodb.Client {
	cfg := getConfig()
	stage := getEnvVar("STAGE")
	log.Printf("Getting dynamo session")
	dynamo := dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		options.Region = "us-west-2"
		if stage == "Testing" {
			log.Printf("Testing environment. Setting dynamo endpoint to localhost:8000")
			dynamoHostname := getEnvVar("DYNAMO_CONTAINER")
			endpoint := fmt.Sprintf("http://%v:8000/", dynamoHostname)
			options.EndpointResolver = dynamodb.EndpointResolverFromURL(endpoint)
		}
	})

	return dynamo
}

func formBaseServerItem(
	ownerID string,
	owner string,
	application string,
	serverName string,
	port int,
) map[string]dynamotypes.AttributeValue {
	portString := strconv.Itoa(port)
	serverItem := map[string]dynamotypes.AttributeValue{
		"OwnerID":     &dynamotypes.AttributeValueMemberS{Value: ownerID},
		"Owner":       &dynamotypes.AttributeValueMemberS{Value: owner},
		"Application": &dynamotypes.AttributeValueMemberS{Value: application},
		"ServerName":  &dynamotypes.AttributeValueMemberS{Value: serverName},
		"Port":        &dynamotypes.AttributeValueMemberN{Value: portString},
	}

	return serverItem
}

func formServerID() string {
	uuidWithHyphen := uuid.New()
	uuidString := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return strings.ToUpper(uuidString[0:4])
}

func getEnvVar(key string) string {

	env := godotenv.Load(".env")

	if env == nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

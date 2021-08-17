package main

import (
	"context"
	"log"
	"strconv"
	"strings"

	gu "generalutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

func queryAWSAccountID(dynamo *dynamodb.Client, userID string) string {
	table := gu.GetEnvVar("AWS_TABLE")

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

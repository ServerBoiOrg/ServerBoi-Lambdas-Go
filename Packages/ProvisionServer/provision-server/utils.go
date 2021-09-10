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

func queryTable(userID string, table string) *dynamodb.GetItemOutput {
	dynamo := gu.GetDynamo()

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

	return response
}

func queryAWSAccountID(userID string) string {
	table := gu.GetEnvVar("OWNER_TABLE")

	response := queryTable(userID, table)
	var awsResponse AWSTableResponse
	attributevalue.UnmarshalMap(response.Item, &awsResponse)

	return awsResponse.AWSAccountID
}

func queryLinodeApiKey(userID string) string {
	table := gu.GetEnvVar("LINODE_TABLE")

	response := queryTable(userID, table)
	var linodeResponse LinodeTableResponse
	attributevalue.UnmarshalMap(response.Item, &linodeResponse)

	return linodeResponse.ApiKey
}

func formBaseServerItem(
	ownerID string,
	owner string,
	application string,
	serverName string,
	service string,
	port int,
	serverID string,
	authorized gu.Authorized,
) map[string]dynamotypes.AttributeValue {
	portString := strconv.Itoa(port)

	auth := map[string]dynamotypes.AttributeValue{
		"Users": &dynamotypes.AttributeValueMemberL{
			Value: []dynamotypes.AttributeValue{
				&dynamotypes.AttributeValueMemberS{Value: authorized.Users[0]}},
		},
		"Roles": &dynamotypes.AttributeValueMemberL{
			Value: []dynamotypes.AttributeValue{},
		},
	}

	serverItem := map[string]dynamotypes.AttributeValue{
		"OwnerID":     &dynamotypes.AttributeValueMemberS{Value: ownerID},
		"Owner":       &dynamotypes.AttributeValueMemberS{Value: owner},
		"Application": &dynamotypes.AttributeValueMemberS{Value: application},
		"ServerName":  &dynamotypes.AttributeValueMemberS{Value: serverName},
		"Service":     &dynamotypes.AttributeValueMemberS{Value: service},
		"ServerID":    &dynamotypes.AttributeValueMemberS{Value: serverID},
		"Port":        &dynamotypes.AttributeValueMemberN{Value: portString},
		"Authorized":  &dynamotypes.AttributeValueMemberM{Value: auth},
	}

	return serverItem
}

func formServerID() string {
	uuidWithHyphen := uuid.New()
	uuidString := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return strings.ToUpper(uuidString[0:4])
}

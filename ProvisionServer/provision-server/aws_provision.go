package provision

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

func provisionAWS(params ProvisonServerParameters) {

	dynamo := getDynamo()

	log.Printf("Querying aws account for %v item from Dynamo", params.Owner)
	accountID := queryAWSAccountID(dynamo, params.OwnerID)
}

type AWSTableResponse struct {
	UserID       string `json:UserID`
	AWSAccountID string `json:AWSAccountID`
}

func queryAWSAccountID(dynamo *dynamodb.Client, userID string) string {
	table := getEnvVar("AWS_TABLE")

	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"UserID": &types.AttributeValueMemberS{Value: userID},
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

func getConfig() aws.Config {
	log.Printf("Getting config")

	config, err := config.LoadDefaultConfig(context.TODO(), func(options *config.LoadOptions) error {
		options.Region = getEnvVar("AWS_REGION")

		return nil
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	return config
}

func getDynamo() *dynamodb.Client {
	cfg := getConfig()
	stage := getEnvVar("STAGE")
	log.Printf("Getting dynamo session")
	dynamo := dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		if stage == "Testing" {
			options.EndpointResolver = dynamodb.EndpointResolverFromURL("http://localhost:8000")
		}
	})

	return dynamo
}

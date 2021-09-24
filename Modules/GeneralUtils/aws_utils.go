package generalutils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func getConfig() aws.Config {
	config, err := config.LoadDefaultConfig(context.TODO(), func(options *config.LoadOptions) error {
		options.Region = GetEnvVar("AWS_REGION")

		return nil
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	return config
}

func GetDynamo() (dynamo *dynamodb.Client) {
	cfg := getConfig()
	stage := GetEnvVar("STAGE")
	dynamo = dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		if stage == "Testing" {
			options.EndpointResolver = dynamodb.EndpointResolverFromURL("http://localhost:8000")
		}
	})

	return dynamo
}

func GetCloudwatchClient() *cloudwatch.Client {
	cfg := getConfig()
	cw := cloudwatch.NewFromConfig(cfg, func(options *cloudwatch.Options) {})

	return cw
}

func CreateEC2Client(region string, accountID string) *ec2.Client {
	stage := GetEnvVar("STAGE")
	var options ec2.Options

	if stage == "Testing" {
		log.Printf("Testing environment. Setting ec2 endpoint to localhost container")
		localstackHostname := GetEnvVar("LOCALSTACK_CONTAINER")
		endpoint := fmt.Sprintf("http://%v:4566/", localstackHostname)
		options.EndpointResolver = ec2.EndpointResolverFromURL(endpoint)
	} else {
		creds := getRemoteCreds(region, accountID)

		options = ec2.Options{
			Region:      region,
			Credentials: creds,
		}
	}

	client := ec2.New(options)

	return client
}

func CreateSfnClient() *sfn.Client {
	stage := GetEnvVar("STAGE")

	cfg := getConfig()
	client := sfn.NewFromConfig(cfg, func(options *sfn.Options) {
		if stage == "Testing" {
			options.EndpointResolver = sfn.EndpointResolverFromURL("http://localhost:8083")
		}
	})

	return client
}

func StartSfnExecution(statemachineArn string, executionName string, input string) {
	client := CreateSfnClient()

	executionInput := sfn.StartExecutionInput{
		StateMachineArn: &statemachineArn,
		Name:            &executionName,
		Input:           &input,
	}
	_, err := client.StartExecution(context.TODO(), &executionInput)
	if err != nil {
		log.Fatalf("Error starting execution: %v", err)
	}
}

func GetS3Client() *s3.Client {
	cfg := getConfig()
	client := s3.NewFromConfig(cfg)

	return client
}

func GetPresignedS3Client() *s3.PresignClient {
	cfg := getConfig()
	client := s3.NewFromConfig(cfg)
	pre := s3.NewPresignClient(client)

	return pre
}

func getRemoteCreds(region string, accountID string) *aws.CredentialsCache {
	log.Printf("Getting credentials for account: %s", accountID)
	cfg := getConfig()
	roleSession := "ServerBoiGo-Provision-Session"
	roleArn := fmt.Sprintf("arn:aws:iam::%v:role/ServerBoi-Resource.Assumed-Role", accountID)

	log.Printf("RoleARN: %v", roleArn)

	stsClient := sts.NewFromConfig(cfg)

	input := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &roleSession,
	}

	newRole, err := stsClient.AssumeRole(context.Background(), input)
	if err != nil {
		panic(err)
	}

	accessKey := newRole.Credentials.AccessKeyId
	secretKey := newRole.Credentials.SecretAccessKey
	sessionToken := newRole.Credentials.SessionToken

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(*accessKey, *secretKey, *sessionToken))

	return creds
}

func GetWebhookFromGuildID(guildID string) *WebhookTableResponse {
	dynamo := GetDynamo()
	webhookTable := GetEnvVar("WEBHOOK_TABLE")

	log.Printf("Querying webhook for guild %v from Dynamo", guildID)
	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(webhookTable),
		Key: map[string]types.AttributeValue{
			"GuildID": &types.AttributeValueMemberS{Value: guildID},
		},
	})
	if err != nil {
		log.Fatalf("Error retrieving item from dynamo: %v", err)
	}

	var responseItem WebhookTableResponse
	attributevalue.UnmarshalMap(response.Item, &responseItem)

	return &responseItem
}

func GetChannelIDFromGuildID(guildID string) (channelID string, err error) {
	dynamo := GetDynamo()
	webhookTable := GetEnvVar("CHANNEL_TABLE")

	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(webhookTable),
		Key: map[string]types.AttributeValue{
			"GuildID": &types.AttributeValueMemberS{Value: guildID},
		},
	})
	if err != nil {
		return channelID, err
	}

	var responseItem ChannelTableResponse
	attributevalue.UnmarshalMap(response.Item, &responseItem)

	return responseItem.ChannelID, nil
}

type UpdateOwnerItemInput struct {
	OwnerID    string
	FieldName  string
	FieldValue string
}

func UpdateOwnerItem(input *UpdateOwnerItemInput) error {
	dynamo := GetDynamo()
	table := GetEnvVar("OWNER_TABLE")

	key := map[string]types.AttributeValue{
		"OwnerID": &types.AttributeValueMemberS{Value: input.OwnerID},
	}
	value := map[string]types.AttributeValue{
		":item": &types.AttributeValueMemberS{Value: input.FieldValue},
	}
	updateExpression := fmt.Sprintf("SET %v = :item", input.FieldName)

	_, err := dynamo.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName:                 aws.String(table),
		Key:                       key,
		UpdateExpression:          &updateExpression,
		ExpressionAttributeValues: value,
	})
	return err
}

func RemoveFieldFromOwnerItem(input *UpdateOwnerItemInput) error {
	dynamo := GetDynamo()
	table := GetEnvVar("OWNER_TABLE")

	key := map[string]types.AttributeValue{
		"OwnerID": &types.AttributeValueMemberS{Value: input.OwnerID},
	}
	updateExpression := fmt.Sprintf("REMOVE %v", input.FieldName)

	_, err := dynamo.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName:        aws.String(table),
		Key:              key,
		UpdateExpression: &updateExpression,
	})
	return err
}

func GetOwnerItem(ownerID string) (ownerItem *OwnerItem, err error) {
	dynamo := GetDynamo()
	ownerTable := GetEnvVar("OWNER_TABLE")

	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(ownerTable),
		Key: map[string]types.AttributeValue{
			"OwnerID": &types.AttributeValueMemberS{Value: ownerID},
		},
	})
	if err != nil {
		return ownerItem, err
	} else if len(response.Item) == 0 {
		err = errors.New("No items found.")
		return ownerItem, err
	}

	attributevalue.UnmarshalMap(response.Item, &ownerItem)
	return ownerItem, nil
}

func ServerIDExists(serverID string) (bool, error) {
	dynamo := GetDynamo()
	serverTable := GetEnvVar("SERVER_TABLE")

	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(serverTable),
		Key: map[string]types.AttributeValue{
			"ServerID": &types.AttributeValueMemberS{Value: serverID},
		},
	})
	if err != nil {
		return false, err
	}

	if len(response.Item) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func GetServerFromID(serverID string) (server Server, err error) {
	dynamo := GetDynamo()
	serverTable := GetEnvVar("SERVER_TABLE")

	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(serverTable),
		Key: map[string]types.AttributeValue{
			"ServerID": &types.AttributeValueMemberS{Value: serverID},
		},
	})
	if err != nil {
		log.Printf("Error retrieving item from dynamo: %v", err)
		return server, err
	} else if len(response.Item) == 0 {
		log.Printf("No item was found")
		err = errors.New("No items found.")
		return server, err
	}

	serviceRaw := response.Item["Service"]
	var service string
	log.Printf("Unmarshaling service attribute value.")
	attributevalue.Unmarshal(serviceRaw, &service)

	log.Printf("Service is %v", service)
	switch strings.ToLower(service) {
	case "aws":
		awsServer := AWSServer{}
		attributevalue.UnmarshalMap(response.Item, &awsServer)
		return awsServer, nil
	case "linode":
		linodeServer := LinodeServer{}
		err = attributevalue.UnmarshalMap(response.Item, &linodeServer)
		return linodeServer, nil
	default:
		panic("Unknown service")
	}
}

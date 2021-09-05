package generalutils

import (
	"context"
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

func CreateEC2Client(region string, accountID string) ec2.Client {
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

	return *client
}

func CreateSfnClient() sfn.Client {
	stage := GetEnvVar("STAGE")

	cfg := getConfig()
	client := sfn.NewFromConfig(cfg, func(options *sfn.Options) {
		if stage == "Testing" {
			options.EndpointResolver = sfn.EndpointResolverFromURL("http://localhost:8083")
		}
	})

	return *client
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
	log.Printf("Getting S3 client")
	s3 := s3.NewFromConfig(cfg)

	return s3
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

func (server AWSServer) Start() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	_, err = client.StartInstances(context.Background(), input)
	if err != nil {
		return data, err
	}
	formRespInput := FormResponseInput{
		"Content": "Starting server",
	}

	return FormResponseData(formRespInput), nil
}

func (server AWSServer) Stop() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	_, err = client.StopInstances(context.Background(), input)
	if err != nil {
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Stopping server",
	}

	return FormResponseData(formRespInput), nil
}

func (server AWSServer) GetService() string {
	return server.Service
}

func (server AWSServer) GetIPv4() (string, error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	log.Printf("Describing instance: %s", server.InstanceID)
	response, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", response.Reservations[0].Instances[0].PublicIpAddress), nil
}

func (server AWSServer) GetServerBoiRegion() ServerBoiRegion {
	return FormServerBoiRegion(server.Service, server.Region)
}

func (server AWSServer) GetBaseService() BaseServer {
	return BaseServer{
		ServerID:    server.ServerID,
		Application: server.Application,
		ServerName:  server.ServerName,
		Service:     server.Service,
		Owner:       server.Owner,
		OwnerID:     server.OwnerID,
		Port:        server.Port,
	}
}

func (server AWSServer) Restart() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	input := &ec2.RebootInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	_, err = client.RebootInstances(context.Background(), input)
	if err != nil {
		return data, err
	}
	formRespInput := FormResponseInput{
		"Content": "Restarting server",
	}

	return FormResponseData(formRespInput), nil
}

func (server AWSServer) Status() (status string, err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	log.Printf("Ec2 Client made in Target account")
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	log.Printf("Describing instance: %s", server.InstanceID)
	response, err := client.DescribeInstances(context.Background(), input)
	if err != nil {
		return status, err
	}

	return fmt.Sprintf("%v", response.Reservations[0].Instances[0].State.Name), nil
}

func GetWebhookFromGuildID(guildID string) WebhookTableResponse {
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

	return responseItem
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

func GetServerFromID(serverID string) (server Server, err error) {
	dynamo := GetDynamo()
	serverTable := GetEnvVar("SERVER_TABLE")

	log.Printf("Querying server %v item from Dynamo", serverID)
	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(serverTable),
		Key: map[string]types.AttributeValue{
			"ServerID": &types.AttributeValueMemberS{Value: serverID},
		},
	})
	if err != nil {
		log.Printf("Error retrieving item from dynamo: %v", err)
		return server, nil
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

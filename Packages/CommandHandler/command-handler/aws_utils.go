package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

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

func getDynamo() (dynamo *dynamodb.Client) {
	cfg := getConfig()
	stage := getEnvVar("STAGE")
	log.Printf("Getting dynamo session")
	dynamo = dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		if stage == "Testing" {
			options.EndpointResolver = dynamodb.EndpointResolverFromURL("http://localhost:8000")
		}
	})

	return dynamo
}

func createEC2Client(serverInfo AWSService) ec2.Client {
	stage := getEnvVar("STAGE")
	log.Printf("Making EC2 client in account: %v", serverInfo.AccountID)
	creds := getRemoteCreds(serverInfo.Region, serverInfo.AccountID)
	log.Printf("Got credentials for account")

	options := ec2.Options{
		Region:      serverInfo.Region,
		Credentials: creds,
	}
	if stage == "Testing" {
		options.EndpointResolver = ec2.EndpointResolverFromURL("http://localhost:4566")
	}

	client := ec2.New(options)
	log.Printf("EC2 Client created")

	return *client
}

func createSfnClient() sfn.Client {
	stage := getEnvVar("STAGE")

	cfg := getConfig()
	client := sfn.NewFromConfig(cfg, func(options *sfn.Options) {
		if stage == "Testing" {
			options.EndpointResolver = sfn.EndpointResolverFromURL("http://localhost:8083")
		}
	})

	return *client
}

func startSfnExecution(statemachineArn string, executionName string, input string) {
	client := createSfnClient()

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

func getRemoteCreds(region string, accountID string) *aws.CredentialsCache {
	log.Printf("Getting credentials for account: %s", accountID)
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Printf("unable to load SDK config, %v", err)
	}

	roleSession := fmt.Sprintf("ServerBoiGo-%v-%v-Session", accountID, region)
	roleArn := fmt.Sprintf("arn:aws:iam::%v:role/ServerBoi-Resource.Assumed-Role", accountID)

	stsClient := sts.NewFromConfig(cfg)

	input := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &roleSession,
	}

	newRole, err := stsClient.AssumeRole(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error assuming the role:")
		fmt.Println(err)
	}

	accessKey := newRole.Credentials.AccessKeyId
	secretKey := newRole.Credentials.SecretAccessKey
	sessionToken := newRole.Credentials.SessionToken

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(*accessKey, *secretKey, *sessionToken))

	return creds

}

func (server AWSServer) start() (data DiscordInteractionResponseData, err error) {
	client := createEC2Client(server.ServiceInfo)
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{
			server.ServiceInfo.InstanceID,
		},
	}
	_, err = client.StartInstances(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error retrieving starting EC2 instances:")
		fmt.Println(err)
		return data, err
	}
	formRespInput := FormResponseInput{
		"Content": "Starting server",
	}

	return formResponseData(formRespInput), nil
}

func (server AWSServer) stop() (data DiscordInteractionResponseData, err error) {
	client := createEC2Client(server.ServiceInfo)
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{
			server.ServiceInfo.InstanceID,
		},
	}
	_, err = client.StopInstances(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error retrieving starting EC2 instances:")
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Stopping server",
	}

	return formResponseData(formRespInput), nil
}

func (server AWSServer) restart() (data DiscordInteractionResponseData, err error) {
	client := createEC2Client(server.ServiceInfo)
	input := &ec2.RebootInstancesInput{
		InstanceIds: []string{
			server.ServiceInfo.InstanceID,
		},
	}
	_, err = client.RebootInstances(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error retrieving starting EC2 instances:")
		fmt.Println(err)
		return data, err
	}
	formRespInput := FormResponseInput{
		"Content": "Restarting server",
	}

	return formResponseData(formRespInput), nil
}

func (server AWSServer) status() (data DiscordInteractionResponseData, err error) {
	client := createEC2Client(server.ServiceInfo)
	log.Printf("Ec2 Client made in Target account")
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			server.ServiceInfo.InstanceID,
		},
	}
	log.Printf("Describing instance: %s", server.ServiceInfo.InstanceID)
	response, err := client.DescribeInstances(context.Background(), input)
	if err != nil {
		fmt.Println(err)
		return data, err
	}
	for _, r := range response.Reservations {
		for _, i := range r.Instances {
			formRespInput := FormResponseInput{
				"Content": fmt.Sprintf("Server status: %s", i.State.Name),
			}

			return formResponseData(formRespInput), nil
		}
	}
	return
}

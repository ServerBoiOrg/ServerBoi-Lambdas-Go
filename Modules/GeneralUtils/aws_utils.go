package generalutils

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func getConfig() aws.Config {
	log.Printf("Getting config")

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
	log.Printf("Getting dynamo session")
	dynamo = dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		if stage == "Testing" {
			options.EndpointResolver = dynamodb.EndpointResolverFromURL("http://localhost:8000")
		}
	})

	return dynamo
}

func GetCloudwatchClient() *cloudwatch.Client {
	cfg := getConfig()
	log.Printf("Getting cloudwatch client")
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
		log.Printf("Making EC2 client in account: %v", accountID)
		creds := getRemoteCreds(region, accountID)
		log.Printf("Got credentials for account.")

		options = ec2.Options{
			Region:      region,
			Credentials: creds,
		}
	}

	client := ec2.New(options)
	log.Printf("EC2 Client created")

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
	log.Printf("Getting cloudwatch client")
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
		fmt.Println("Got an error assuming the role:")
		panic(err)
	}

	accessKey := newRole.Credentials.AccessKeyId
	secretKey := newRole.Credentials.SecretAccessKey
	sessionToken := newRole.Credentials.SessionToken

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(*accessKey, *secretKey, *sessionToken))

	return creds
}

func (server AWSServer) Start() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.ServiceInfo.Region, server.ServiceInfo.AccountID)
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

	return FormResponseData(formRespInput), nil
}

func (server AWSServer) Stop() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.ServiceInfo.Region, server.ServiceInfo.AccountID)
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

	return FormResponseData(formRespInput), nil
}

func (server AWSServer) Restart() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.ServiceInfo.Region, server.ServiceInfo.AccountID)
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

	return FormResponseData(formRespInput), nil
}

func (server AWSServer) Status() (data DiscordInteractionResponseData, err error) {
	client := CreateEC2Client(server.ServiceInfo.Region, server.ServiceInfo.AccountID)
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

			return FormResponseData(formRespInput), nil
		}
	}
	return
}

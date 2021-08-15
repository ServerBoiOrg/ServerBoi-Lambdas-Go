package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ServerCommandResponse struct {
	Status string
	Result bool
}

func routeServerCommand(command DiscordInteractionApplicationCommand) (response DiscordInteractionResponseData, err error) {
	serverCommand := command.Data.Options[0].Options[0].Name
	log.Printf("Server Commmad Option: %v", serverCommand)

	serverID := command.Data.Options[0].Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := getServerFromID(serverID)
	if err != nil {
		log.Fatalf("Unable to get server object. Error: %s", err)
	}
	log.Printf("Server Object: %s", server)
	log.Printf("Running %s on server %s", serverCommand, serverID)

	var data DiscordInteractionResponseData
	switch {
	//Server Actions
	case serverCommand == "status":
		data, err = server.status()
	case serverCommand == "start":
		data, err = server.start()
	case serverCommand == "stop":
		data, err = server.stop()
	case serverCommand == "restart":
		data, err = server.restart()
	//Workflows
	// case serverCommand == "add":
	case serverCommand == "terminate":
		input := ServerTerminateInput{
			Token:         command.Token,
			InteractionID: command.ID,
			ApplicationID: command.ApplicationID,
			ServerID:      command.Data.Options[0].Options[0].Name,
		}
		data, err = serverTerminate(input)
	}
	if err != nil {
		log.Fatalf("Error performing command: %v", err)
		return response, err
	}

	return data, nil
}

type ServerTerminateInput struct {
	Token         string
	InteractionID string
	ApplicationID string
	ServerID      string
}

type TerminateWorkflow struct {
	Token         string
	ApplicationID string
	ServerID      string
	ExecutionName string
}

func serverTerminate(input ServerTerminateInput) (data DiscordInteractionResponseData, err error) {
	executionName := generateWorkflowUUID("Terminate")
	terminateArn := getEnvVar("TERMINATE_ARN")

	terminationWorkflowInput := TerminateWorkflow{
		ServerID:      input.ServerID,
		Token:         input.Token,
		ApplicationID: input.ApplicationID,
		ExecutionName: executionName,
	}
	inputJson, err := json.Marshal(terminationWorkflowInput)
	if err != nil {
		log.Println(err)
	}
	inputString := fmt.Sprintf(string(inputJson))

	startSfnExecution(terminateArn, executionName, inputString)

	embedInput := FormWorkflowEmbedInput{
		Name:        "Terminate-Workflow",
		Description: fmt.Sprintf("WorkflowID: %s", executionName),
		Status:      "⏳ Pending",
		Stage:       "Starting...",
		Color:       Greyple,
	}
	workflowEmbed := formWorkflowEmbed(embedInput)

	formRespInput := FormResponseInput{
		"Embeds": workflowEmbed,
	}

	return formResponseData(formRespInput), nil
}

type ServerActionInput struct {
	ServerID string
}

func getServerFromID(serverID string) (server Server, err error) {
	dynamo := getDynamo()
	serverTable := getEnvVar("SERVER_TABLE")

	log.Printf("Querying server %v item from Dynamo", serverID)
	response, err := dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(serverTable),
		Key: map[string]types.AttributeValue{
			"ServerID": &types.AttributeValueMemberS{Value: serverID},
		},
	})
	if err != nil {
		log.Fatalf("Error retrieving item from dynamo: %v", err)
		return server, err
	}

	var serverInfo ServerBoiServer
	err = attributevalue.UnmarshalMap(response.Item, &serverInfo)

	log.Printf("Server Item: %v", serverInfo)

	service := serverInfo.Service["Name"]

	log.Printf("Service of server: %v", service)
	switch {
	case strings.ToLower(service) == "aws":
		var awsService AWSService
		jsonedService, _ := json.Marshal(serverInfo.Service)
		json.Unmarshal(jsonedService, &awsService)
		log.Printf("Service Item: %v", awsService)

		server := AWSServer{
			ServiceInfo: awsService,
		}
		return server, nil
	case strings.ToLower(service) == "linode":
		var service LinodeService
		jsonedService, _ := json.Marshal(serverInfo.Service)
		json.Unmarshal(jsonedService, &service)
		log.Printf("Service Item: %v", service)
		return server, nil
	}

	return server, err
}

type Server interface {
	start() (data DiscordInteractionResponseData, err error)
	stop() (data DiscordInteractionResponseData, err error)
	restart() (data DiscordInteractionResponseData, err error)
	status() (data DiscordInteractionResponseData, err error)
}

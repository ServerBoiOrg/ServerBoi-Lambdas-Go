package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type CreateServerWorkflowInput struct {
	ExecutionName    string
	Application      string
	Service          string
	OwnerID          string
	Owner            string
	InteractionID    string
	InteractionToken string
	ApplicationID    string
	GuildID          string
	Url              string
	ServerName       string
	CreationOptions  map[string]string
}

func createServer(command DiscordInteractionApplicationCommand) (response DiscordInteractionResponseData, err error) {
	application := command.Data.Options[0].Options[0].Name
	optionsSlice := command.Data.Options[0].Options[0].Options
	creationOptions := make(map[string]string)
	for _, option := range optionsSlice {
		creationOptions[option.Name] = option.Value
	}
	executionName := generateWorkflowUUID("Provision")
	service := creationOptions["service"]
	delete(creationOptions, service)

	errors := verifyCreateServerParams(creationOptions)
	if len(errors) > 0 {
		return formInvalidParametersResponse(errors), nil
	}

	log.Printf("Application to create: %v", application)
	log.Printf("Service provider: %v", service)

	var name string
	_, ok := creationOptions["name"]
	if ok {
		name = creationOptions["name"]
		delete(creationOptions, name)
	} else {
		name = fmt.Sprintf("ServerBoi-%v", application)
	}

	executionInput := CreateServerWorkflowInput{
		ExecutionName:    executionName,
		Application:      application,
		Service:          service,
		OwnerID:          command.User.ID,
		Owner:            command.User.Username,
		InteractionID:    command.ID,
		InteractionToken: command.Token,
		ApplicationID:    command.ApplicationID,
		GuildID:          command.GuildID,
		Url:              getEnvVar("URL"),
		ServerName:       name,
		CreationOptions:  creationOptions,
	}
	log.Printf("Provision Workflow Input: %v", executionInput)

	log.Printf("Converting input to string for submission.")
	inputJson, err := json.Marshal(executionInput)
	if err != nil {
		log.Println(err)
	}
	inputString := fmt.Sprintf(string(inputJson))

	provisionArn := getEnvVar("PROVISION_ARN")

	log.Printf("Submitting workflow")
	startSfnExecution(provisionArn, executionName, inputString)

	log.Printf("Forming workflow embed")
	embedInput := FormWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", executionName),
		Status:      "⏳ Pending",
		Stage:       "Starting...",
		Color:       Greyple,
	}
	workflowEmbed := formWorkflowEmbed(embedInput)

	log.Printf("Prepping response data")
	formRespInput := FormResponseInput{
		"Embeds": workflowEmbed,
	}

	return formResponseData(formRespInput), nil
}

func verifyCreateServerParams(options map[string]string) []string {
	errors := []string{}
	serviceErr := verifyService(options["service"])
	if serviceErr != nil {
		errors = append(errors, serviceErr.Error())
	}
	regionErr := verifyRegion(options["service"], options["region"])
	if regionErr != nil {
		errors = append(errors, regionErr.Error())
	}
	return errors
}

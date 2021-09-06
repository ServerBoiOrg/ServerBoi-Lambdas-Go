package main

import (
	"encoding/json"
	"fmt"
	"log"

	gu "generalutils"
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

func createServer(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	application := command.Data.Options[0].Options[0].Name
	optionsSlice := command.Data.Options[0].Options[0].Options
	creationOptions := make(map[string]string)
	for _, option := range optionsSlice {
		creationOptions[option.Name] = option.Value
	}
	executionName := gu.GenerateWorkflowUUID("Provision")
	service := creationOptions["service"]
	delete(creationOptions, service)

	errors := verifyCreateServerParams(creationOptions)
	if len(errors) > 0 {
		return gu.FormInvalidParametersResponse(errors)
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
		OwnerID:          command.Member.User.ID,
		Owner:            command.Member.User.Username,
		InteractionID:    command.ID,
		InteractionToken: command.Token,
		ApplicationID:    command.ApplicationID,
		GuildID:          command.GuildID,
		Url:              gu.GetEnvVar("API_URL"),
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

	provisionArn := gu.GetEnvVar("PROVISION_ARN")

	log.Printf("Submitting workflow")
	gu.StartSfnExecution(provisionArn, executionName, inputString)

	log.Printf("Forming workflow embed")
	embedInput := gu.FormWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", executionName),
		Status:      "⏳ Pending",
		Stage:       "Starting...",
		Color:       gu.Greyple,
	}
	workflowEmbed := gu.FormWorkflowEmbed(embedInput)

	log.Printf("Prepping response data")
	formRespInput := gu.FormResponseInput{
		"Embeds": workflowEmbed,
	}

	return gu.FormResponseData(formRespInput)
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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	gu "generalutils"
)

type CreateServerWorkflowInput struct {
	ExecutionName    string
	Application      string
	OwnerID          string
	Owner            string
	InteractionID    string
	InteractionToken string
	ApplicationID    string
	GuildID          string
	Url              string
	CreationOptions  map[string]string `json:"CreationOptions,omitempty"`
	Service          string
	Name             string
	Region           string
	HardwareType     string `json:"HardwareType,omitempty"`
	Private          bool
}

type GenericCreationOptions struct {
	Service          string `json:"service"`
	Region           string `json:"region"`
	Name             string `json:"name,omitempty"`
	RegionOverride   string `json:"override-region,omitempty"`
	HardwareOverride string `json:"override-hardware,omitempty"`
	Private          bool   `json:"private,omitempty"`
}

type CreateOptions struct {
	Service          string
	Name             string
	Region           string
	HardwareType     string
	Private          bool
	OptionalCommands map[string]string
}

func verifyCreationOptions(creationOptions map[string]interface{}) (output CreateOptions, errors []string) {
	output.Service = creationOptions["service"].(string)
	delete(creationOptions, "service")

	log.Printf("Service to create server on: %v", output.Service)
	//Check Name
	if name, ok := creationOptions["name"]; ok {
		name := name.(string)
		name, err := verifyName(name)
		if err != nil {
			errors = append(errors, "Provided name not permitted")
		} else {
			output.Name = name
		}
	}
	delete(creationOptions, "name")
	log.Printf("Name verified: %v", output.Name)

	//Set Region
	if creationOptions["region"] == "override" {
		if overrideRegion, ok := creationOptions["override-region"]; ok {
			overrideRegion := overrideRegion.(string)
			err := verifyRegion(overrideRegion, output.Service)
			if err != nil {
				errors = append(errors, fmt.Sprintf(
					"Region %v is not valid for service %v.",
					overrideRegion,
					creationOptions["service"],
				))
			} else {
				delete(creationOptions, "override-region")
				output.Region = overrideRegion
			}
		} else {
			errors = append(errors, "No region provided for override.")
		}
	} else {
		output.Region = convertRegion(creationOptions["region"].(string), output.Service)
	}
	delete(creationOptions, "region")
	log.Printf("Region verified: %v", output.Region)

	//Check if Hardware Type override
	if hardwareType, ok := creationOptions["override-hardware"]; ok {
		err := verifyHardwareType(hardwareType.(string), output.Service)
		if err != nil {
			errors = append(errors, "Hardware type %v is not valid for service %v", hardwareType.(string), output.Service)
		} else {
			output.HardwareType = hardwareType.(string)
			delete(creationOptions, "override-hardware")
			log.Printf("Hardware Override verified: %v", output.HardwareType)
		}
	}

	//Check if private
	if private, ok := creationOptions["private"]; ok {
		output.Private = private.(bool)
		delete(creationOptions, "private")
	}
	log.Printf("Private: %v", output.Private)

	tmp := make(map[string]string)
	for key, value := range creationOptions {
		tmp[key] = value.(string)
	}
	output.OptionalCommands = tmp
	log.Printf("Creation Options: %v", output.OptionalCommands)

	return output, nil
}

func verifyName(name string) (string, error) {
	return name, nil
}

func verifyRegion(region string, service string) (err error) {
	switch service {
	case "aws":
		log.Printf("Checking AWS region %v", region)
		for _, awsRegion := range gu.AWSRegions {
			if region == awsRegion {
				return nil
			}
		}
	case "linode":
		log.Printf("Checking Linode region %v", region)
		for _, linodeRegion := range gu.AWSRegions {
			if region == linodeRegion {
				return nil
			}
		}
	}
	return errors.New("Bad region")
}

func verifyHardwareType(hardwareType string, service string) (err error) {
	return nil
}

func convertRegion(serverboiRegion string, service string) (region string) {
	rand.Seed(time.Now().Unix())

	switch serverboiRegion {
	case "us-west":
		switch service {
		case "aws":
			list := []string{"us-west-2", "us-west-1"}
			region = list[rand.Intn(len(list))]
		case "linode":
			region = "us-west"
		}
	case "us-east":
		switch service {
		case "aws":
			list := []string{"us-east-2", "us-east-1"}
			region = list[rand.Intn(len(list))]
		case "linode":
			region = "us-east"
		}
	case "us-central":
		switch service {
		case "aws":
			region = "us-east-2"
		case "linode":
			region = "us-central"
		}
	case "us-south":
		switch service {
		case "aws":
			region = "us-east-2"
		case "linode":
			region = "us-southeast"
		}
	}
	return region
}

func createServer(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	log.Printf("Running create command")

	application := command.Data.Options[0].Name
	log.Printf("Application: %v", application)

	optionsMap := make(map[string]interface{})
	for _, option := range command.Data.Options[0].Options {
		optionsMap[option.Name] = option.Value
	}
	executionName := gu.GenerateWorkflowUUID("Provision")

	verifiedParams, errors := verifyCreationOptions(optionsMap)
	if len(errors) != 0 {
		message := "Couldn't verify provided parameters. The following problems were present.\n"
		for _, e := range errors {
			message = fmt.Sprintf("%v* %v", message, e)
		}
		return gu.DiscordInteractionResponseData{
			Content: message,
		}
	}
	if verifiedParams.Name == "" {
		verifiedParams.Name = fmt.Sprintf("ServerBoi-%v", strings.ToUpper(application))
	}

	log.Printf("Application to create: %v", application)

	executionInput := CreateServerWorkflowInput{
		ExecutionName:    executionName,
		Application:      application,
		OwnerID:          command.Member.User.ID,
		Owner:            command.Member.User.Username,
		InteractionID:    command.ID,
		InteractionToken: command.Token,
		ApplicationID:    command.ApplicationID,
		GuildID:          command.GuildID,
		Url:              gu.GetEnvVar("API_URL"),
		CreationOptions:  verifiedParams.OptionalCommands,
		Service:          verifiedParams.Service,
		Name:             verifiedParams.Name,
		Region:           verifiedParams.Region,
		HardwareType:     verifiedParams.HardwareType,
		Private:          verifiedParams.Private,
	}

	log.Printf("Converting input to string for submission.")
	inputJson, err := json.Marshal(executionInput)
	if err != nil {
		log.Println(err)
	}
	inputString := fmt.Sprintf(string(inputJson))
	log.Printf("Provision Workflow Input: %v", inputString)

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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	gu "generalutils"
	ru "responseutils"

	dt "github.com/awlsring/discordtypes"
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
	ClientPort       int
	QueryPort        int
	CreationOptions  map[string]string `json:"CreationOptions,omitempty"`
	Service          string
	Name             string
	Region           string
	HardwareType     string `json:"HardwareType,omitempty"`
	Visible          bool
	IsRole           bool
}

type CreateOptions struct {
	Service          string
	Name             string
	Region           string
	ProfileID        string
	ProfileName      string
	HardwareType     string
	Visible          bool
	OptionalCommands map[string]string
	ClientPort       int
	QueryPort        int
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

	//Check Profile
	if profile, ok := creationOptions["profile"]; ok {
		output.ProfileID = profile.(string)
		delete(creationOptions, "profile")
	}

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
	if visible, ok := creationOptions["visible"]; ok {
		output.Visible = visible.(bool)
		delete(creationOptions, "visible")
	} else {
		output.Visible = true
	}
	log.Printf("Visible: %v", output.Visible)

	//Check if client port
	if clientPort, ok := creationOptions["clientPort"]; ok {
		p := clientPort.(string)
		pI, err := strconv.Atoi(p)
		if err != nil {
			errors = append(errors, "Client port must be a number between 1-65353")
		} else {
			port := verifyPort(pI)
			if err != nil || port == 0 {
				errors = append(errors, "Client port must be a number between 1-65353")
			} else {
				output.ClientPort = port
			}
		}
		delete(creationOptions, "clientPort")
	}

	//Check if query port
	if queryPort, ok := creationOptions["queryPort"]; ok {
		p := queryPort.(string)
		pI, err := strconv.Atoi(p)
		if err != nil {
			errors = append(errors, "Query port must be a number between 1-65353")
		} else {
			port := verifyPort(pI)
			if err != nil || port == 0 {
				errors = append(errors, "Query port must be a number between 1-65353")
			} else {
				output.QueryPort = port
			}
		}
		delete(creationOptions, "queryPort")
	}

	log.Printf("Visible: %v", output.Visible)

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

func verifyPort(port int) int {
	if port != 0 && port < 65535 {
		return port
	} else {
		return 0
	}
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

func isUserVerifiedForProfile(roleID string, roles []string) bool {
	for _, role := range roles {
		if role == roleID {
			return true
		}
	}
	return false
}

func ownerHasAccountForService(service string, ownerId string) bool {
	ownerItem, err := gu.GetOwnerItem(ownerId)
	if err == nil {
		log.Printf("Checking if owner has account for service %v", service)
		switch service {
		case "aws":
			if ownerItem.AWSAccountID != "" {
				log.Printf("Has AWS Account")
				return true
			}
		case "linode":
			if ownerItem.LinodeApiKey != "" {
				log.Printf("Has Linode Account")
				return true
			}
		}
	} else {
		log.Printf("Error getting Owner item: %v", err)
	}
	return false
}

func getResolvedRoleName(commandOption *dt.InteractionData) (string, error) {
	for _, role := range commandOption.Resolved.Roles {
		return role.Name, nil
	}
	return "", errors.New("No role")
}

func createServer(command *dt.Interaction) (response *dt.InteractionCallbackData) {
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
		return &dt.InteractionCallbackData{
			Content: message,
		}
	}
	if verifiedParams.Name == "" {
		verifiedParams.Name = fmt.Sprintf("ServerBoi-%v", strings.ToUpper(application))
	}

	var (
		ownerName  string
		ownerID    string
		isRole     bool
		authorized bool
	)
	if verifiedParams.ProfileID != "" {
		log.Printf("Create command with profile")
		roleName, err := getResolvedRoleName(command.Data)
		if err == nil {
			ownerID = verifiedParams.ProfileID
			ownerName = roleName
			isRole = true
			authorized = isUserVerifiedForProfile(ownerID, command.Member.Roles)
			log.Printf("Authorized: %v", authorized)
		}
	} else {
		log.Printf("Create command as personal")
		ownerID = command.Member.User.ID
		ownerName = command.Member.User.Username
		isRole = false
		authorized = true
	}

	hasAccount := ownerHasAccountForService(verifiedParams.Service, ownerID)
	log.Printf("Authorized: %v | HasAccount: %v", authorized, hasAccount)
	if authorized && hasAccount {
		log.Printf("Application to create: %v", application)

		executionInput := CreateServerWorkflowInput{
			ExecutionName:    executionName,
			Application:      application,
			OwnerID:          ownerID,
			Owner:            ownerName,
			InteractionID:    command.ID,
			InteractionToken: command.Token,
			ApplicationID:    command.ApplicationID,
			GuildID:          command.GuildID,
			ClientPort:       verifiedParams.ClientPort,
			QueryPort:        verifiedParams.QueryPort,
			Url:              gu.GetEnvVar("API_URL"),
			CreationOptions:  verifiedParams.OptionalCommands,
			Service:          verifiedParams.Service,
			Name:             verifiedParams.Name,
			Region:           verifiedParams.Region,
			HardwareType:     verifiedParams.HardwareType,
			Visible:          verifiedParams.Visible,
			IsRole:           isRole,
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
		workflowEmbed := ru.CreateWorkflowEmbed(&ru.CreateWorkflowEmbedInput{
			Name:        "Provision-Server",
			Description: fmt.Sprintf("WorkflowID: %s", executionName),
			Status:      "⏳ Pending",
			Stage:       "Starting...",
			Color:       ru.Greyple,
		})
		log.Printf("Prepping response data")
		response = &dt.InteractionCallbackData{
			Embeds: []*dt.Embed{workflowEmbed},
		}
	} else {
		var message string
		if hasAccount {
			message = fmt.Sprintf("You are not authorized to use the role %v.", ownerName)
		} else {
			message = fmt.Sprintf("No account registered for chosen service.")
		}
		response = &dt.InteractionCallbackData{
			Content: message,
		}
	}

	return response
}

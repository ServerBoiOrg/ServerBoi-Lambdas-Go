package main

import (
	"encoding/json"
	"fmt"
	gu "generalutils"
	"log"
)

type ServerCommandResponse struct {
	Status string
	Result bool
}

func routeServerCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	serverCommand := command.Data.Options[0].Options[0].Name
	log.Printf("Server Commmad Option: %v", serverCommand)

	serverID := command.Data.Options[0].Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := gu.GetServerFromID(serverID)
	if err != nil {
		log.Fatalf("Unable to get server object. Error: %s", err)
	}
	log.Printf("Server Object: %s", server)
	log.Printf("Running %s on server %s", serverCommand, serverID)

	var data gu.DiscordInteractionResponseData
	switch {
	//Server Actions
	case serverCommand == "status":
		status, err := server.Status()
		var message string
		if err != nil {
			message = "Error getting server status"
		} else {
			message = fmt.Sprintf("Server status: %v", status)
		}
		data = gu.FormResponseData(gu.FormResponseInput{
			"Content": message,
		})
	case serverCommand == "start":
		err = server.Start()
		data = gu.FormResponseData(gu.FormResponseInput{
			"Content": "Starting server",
		})
	case serverCommand == "stop":
		err = server.Stop()
		data = gu.FormResponseData(gu.FormResponseInput{
			"Content": "Stopping server",
		})
	case serverCommand == "restart":
		err = server.Restart()
		data = gu.FormResponseData(gu.FormResponseInput{
			"Content": "Restarting server",
		})
	case serverCommand == "terminate":
		input := ServerTerminateInput{
			Token:         command.Token,
			InteractionID: command.ID,
			ApplicationID: command.ApplicationID,
			ServerID:      serverID,
		}
		data, err = serverTerminate(input)
	default:
		formRespInput := gu.FormResponseInput{
			"Content": fmt.Sprintf("Server command `%v` is unknown.", serverCommand),
		}
		data = gu.FormResponseData(formRespInput)
	}
	if err != nil {
		log.Printf("Error performing command: %v", err)
		return response
	}
	return data
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

func serverTerminate(input ServerTerminateInput) (data gu.DiscordInteractionResponseData, err error) {
	executionName := gu.GenerateWorkflowUUID("Terminate")
	terminateArn := gu.GetEnvVar("TERMINATE_ARN")

	terminationWorkflowInput := TerminateWorkflow{
		ServerID:      input.ServerID,
		Token:         input.Token,
		ApplicationID: input.ApplicationID,
		ExecutionName: executionName,
	}
	inputJson, err := json.Marshal(terminationWorkflowInput)
	if err != nil {
		log.Fatalf("Error marshalling data: %v", err)
	}
	inputString := fmt.Sprintf(string(inputJson))

	gu.StartSfnExecution(terminateArn, executionName, inputString)

	embedInput := gu.FormWorkflowEmbedInput{
		Name:        "Terminate-Workflow",
		Description: fmt.Sprintf("WorkflowID: %s", executionName),
		Status:      "⏳ Pending",
		Stage:       "Starting...",
		Color:       gu.Greyple,
	}
	workflowEmbed := gu.FormWorkflowEmbed(embedInput)

	formRespInput := gu.FormResponseInput{
		"Embeds": workflowEmbed,
	}

	return gu.FormResponseData(formRespInput), nil
}

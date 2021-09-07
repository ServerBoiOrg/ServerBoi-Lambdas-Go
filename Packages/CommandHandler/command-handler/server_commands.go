package main

import (
	"encoding/json"
	"fmt"
	gu "generalutils"
	"log"
	"strings"
)

type ServerCommandResponse struct {
	Status string
	Result bool
}

func routeServerCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	serverCommand := command.Data.Options[0].Name
	log.Printf("Server Commmad Option: %v", serverCommand)

	serverID := command.Data.Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := gu.GetServerFromID(serverID)
	if err != nil {
		return gu.FormResponseData(gu.FormResponseInput{
			"Content": fmt.Sprintf("Server %v can't be found.", serverID),
		})
	}
	log.Printf("Server Object: %s", server)
	log.Printf("Running %s on server %s", serverCommand, serverID)

	var message string
	switch serverCommand {
	//Server Actions
	case "status":
		status, err := server.Status()
		if err != nil {
			message = "Error getting server status"
		} else {
			message = fmt.Sprintf("Server status: %v", status)
		}
	case "start":
		err = server.Start()
		message = "Starting server"
	case "stop":
		err = server.Stop()
		message = "Stopping server"
	case "reboot":
		err = server.Restart()
		message = "Restarting server"
	case "relist":
		status := gu.GetStatus(server)
		var running bool
		if strings.Contains(status, "Running") {
			running = true
		} else {
			running = false
		}
		embed := gu.CreateServerEmbedFromServer(server)
		log.Printf("Getting Channel for Guild")
		channelID, err := gu.GetChannelIDFromGuildID(command.GuildID)
		if err != nil {
			log.Printf("Error getting channelID from dynamo: %v", err)
			message = fmt.Sprintf("Error getting channelID from dynamo: %v", err)
		} else {
			client := gu.CreateDiscordClient(gu.CreateDiscordClientInput{
				BotToken:   gu.GetEnvVar("DISCORD_TOKEN"),
				ApiVersion: "v9",
			})
			log.Printf("Posting message")
			resp, err := client.CreateMessage(
				channelID,
				gu.FormServerEmbedResponseData(gu.FormServerEmbedResponseDataInput{
					ServerEmbed: embed,
					Running:     running,
				}),
			)
			if err != nil {
				log.Printf("Error getting creating message in Channel: %v", err)
				message = fmt.Sprintf("Error getting creating message in Channel: %v", err)
			} else {
				log.Printf("Response form server embed post: %v", resp)
				message = fmt.Sprintf("Server %v embed posted in server channel", serverID)
			}
		}
	case "terminate":
		input := ServerTerminateInput{
			Token:         command.Token,
			InteractionID: command.ID,
			ApplicationID: command.ApplicationID,
			ServerID:      serverID,
		}
		data, err := serverTerminate(input)
		if err == nil {
			return data
		}
	default:
		message = fmt.Sprintf("Server command `%v` is unknown.", serverCommand)
	}
	if err != nil {
		message = fmt.Sprintf("Error performing command: %v", err)
	}
	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
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

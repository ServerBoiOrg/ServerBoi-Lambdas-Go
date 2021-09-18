package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"

	dt "github.com/awlsring/discordtypes"
)

type ServerCommandResponse struct {
	Status string
	Result bool
}

func routeServerCommand(command *dt.Interaction) (response *dt.InteractionCallbackData) {
	serverCommand := command.Data.Options[0].Name
	log.Printf("Server Commmad Option: %v", serverCommand)

	serverID := command.Data.Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := gu.GetServerFromID(serverID)
	if err != nil {
		return &dt.InteractionCallbackData{
			Content: fmt.Sprintf("Server %v can't be found.", serverID),
		}
	}
	log.Printf("Server Object: %s", server)
	log.Printf("Running %s on server %s", serverCommand, serverID)

	var authorized bool
	for _, user := range server.AuthorizedUsers() {
		if user == command.Member.User.ID {
			authorized = true
		}
	}
	for _, role := range server.AuthorizedRoles() {
		for _, userRole := range command.Member.Roles {
			if role == userRole {
				authorized = true
			}
		}
	}

	var message string
	if authorized {
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
	} else {
		message = "You do not have authorization to run commands on this server."
	}
	return &dt.InteractionCallbackData{
		Content: message,
	}
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

func serverRelist(server gu.Server, guildID string) (message string) {
	status, err := server.GetStatus()
	if err != nil {
		return "Error getting server status."
	}
	state, _, err := ru.TranslateState(
		server.GetBaseService().Service,
		status,
	)
	if err != nil {
		return "Error getting server status."
	}
	var running bool
	if strings.Contains(state, "Running") {
		running = true
	} else {
		running = false
	}
	serverInfo := server.GetBaseService()
	ip, err := server.GetIPv4()
	if err != nil {
		return "Error getting server status."
	}
	embed := ru.CreateServerEmbed(ru.GetServerData(&ru.GetServerDataInput{
		Name:        serverInfo.ServerName,
		ID:          serverInfo.ServerID,
		IP:          ip,
		Status:      status,
		Region:      serverInfo.Region,
		Port:        serverInfo.Port,
		Application: serverInfo.Application,
		Owner:       serverInfo.Owner,
		Service:     serverInfo.Service,
	}))
	log.Printf("Getting Channel for Guild")
	channelID, err := gu.GetChannelIDFromGuildID(guildID)
	if err != nil {
		log.Printf("Error getting channelID from dynamo: %v", err)
		message = fmt.Sprintf("Error getting channelID from dynamo: %v", err)
	} else {
		log.Printf("Posting message")
		var (
			resp    *dt.Message
			headers *dc.DiscordHeaders
			err     error
		)
		for {
			resp, headers, err = client.CreateMessage(&dc.CreateMessageInput{
				ChannelID: channelID,
				Data: &dt.CreateMessageData{
					Embeds:     []*dt.Embed{embed},
					Components: ru.ServerEmbedComponents(running),
				},
			})
			if headers.StatusCode == 429 {
				log.Printf("Thottled, waiting")
				time.Sleep(time.Duration(headers.ResetAfter*1000) * time.Millisecond)
			}
			break
		}
		if err != nil {
			log.Printf("Error getting creating message in Channel: %v", err)
			message = fmt.Sprintf("Error getting creating message in Channel: %v", err)
		} else {
			log.Printf("Response form server embed post: %v", resp)
			message = fmt.Sprintf("Server %v embed posted in server channel", serverInfo.ServerID)
		}
	}
	return message
}

func serverTerminate(input ServerTerminateInput) (data *dt.InteractionCallbackData, err error) {
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

	workflowEmbed := ru.CreateWorkflowEmbed(&ru.CreateWorkflowEmbedInput{
		Name:        "Terminate-Workflow",
		Description: fmt.Sprintf("WorkflowID: %s", executionName),
		Status:      "⏳ Pending",
		Stage:       "Starting...",
		Color:       ru.Greyple,
	})

	return &dt.InteractionCallbackData{
		Embeds: []*dt.Embed{workflowEmbed},
	}, nil
}

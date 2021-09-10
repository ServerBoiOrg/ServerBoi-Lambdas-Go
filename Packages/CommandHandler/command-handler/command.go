package main

import (
	"encoding/json"
	"log"

	gu "generalutils"
)

func command(eventBody string) (output InteractionOutput) {
	log.Printf("Command: %v", eventBody)

	//Unmarshal into Interaction Type
	var command gu.DiscordInteractionApplicationCommand
	json.Unmarshal([]byte(eventBody), &command)

	log.Printf("Sending temporary response to Discord")
	gu.SendTempResponse(command.ID, command.Token)

	log.Printf("Command Option: %v", command.Data.Name)
	var response gu.DiscordInteractionResponseData
	switch command.Data.Name {
	case "create":
		response = createServer(command)
	case "set":
		response = routeSetCommand(command)
	case "remove":
		response = routeRemoveCommand(command)
	case "server":
		response = routeServerCommand(command)
	case "authorize":
		response = routeAuthorizeCommand(command)
	case "deauthorize":
		response = routeDeauthorizeCommand(command)
	}
	log.Printf("Response from %v command: %v", command.Data.Name, response)

	return InteractionOutput{
		ApplicationID:    command.ApplicationID,
		InteractionToken: command.Token,
		Response:         response,
	}
}

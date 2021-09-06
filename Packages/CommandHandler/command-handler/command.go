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

	commandOption := command.Data.Options[0].Name
	log.Printf("Command Option: %v", commandOption)
	var response gu.DiscordInteractionResponseData
	switch {
	case commandOption == "create":
		response = createServer(command)
	case commandOption == "onboard":
		response = routeOnboardCommand(command)
	case commandOption == "server":
		response = routeServerCommand(command)
	}
	log.Printf("Response from %v command: %v", commandOption, response)

	return InteractionOutput{
		ApplicationID:    command.ApplicationID,
		InteractionToken: command.Token,
		Response:         response,
	}
}

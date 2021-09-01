package main

import (
	"encoding/json"
	"log"

	gu "generalutils"
)

func command(eventBody string) (applicationID string, interactionToken string, response gu.DiscordInteractionResponseData, err error) {
	log.Printf("Command: %v", eventBody)

	//Unmarshal into Interaction Type
	var command gu.DiscordInteractionApplicationCommand
	json.Unmarshal([]byte(eventBody), &command)

	log.Printf("Sending temporary response to Discord")
	gu.SendTempResponse(command.ID, command.Token)

	commandOption := command.Data.Options[0].Name
	log.Printf("Command Option: %v", commandOption)
	switch {
	case commandOption == "create":
		response, err = createServer(command)
	case commandOption == "onboard":
		response, err = routeOnboardCommand(command)
	case commandOption == "server":
		response, err = routeServerCommand(command)
	}

	if err != nil {
		log.Fatalf("Error performing server command: %v", err)
		return "", "", response, err
	}
	log.Printf("Response from %v command: %v", commandOption, response)

	return command.ApplicationID, command.Token, response, nil
}

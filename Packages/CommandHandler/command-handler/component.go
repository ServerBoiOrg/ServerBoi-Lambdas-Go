package main

import "encoding/json"

func component(eventBody string) (string, string, DiscordInteractionResponseData) {
	var response DiscordInteractionResponseData

	var command DiscordInteractionComponentCommand
	json.Unmarshal([]byte(eventBody), &command)

	return command.ApplicationID, command.Token, response
}

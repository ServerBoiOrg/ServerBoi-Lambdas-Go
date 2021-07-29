package main

import "encoding/json"

func component(eventBody string) (string, string, DiscordInteractionResponse) {
	var response DiscordInteractionResponse

	var command DiscordInteractionComponentCommand
	json.Unmarshal([]byte(eventBody), &command)

	return command.ApplicationID, command.Token, response
}

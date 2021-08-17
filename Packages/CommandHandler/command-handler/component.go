package main

import (
	"encoding/json"
	gu "generalutils"
)

func component(eventBody string) (string, string, gu.DiscordInteractionResponseData) {
	var response gu.DiscordInteractionResponseData

	var command gu.DiscordInteractionComponentCommand
	json.Unmarshal([]byte(eventBody), &command)

	return command.ApplicationID, command.Token, response
}

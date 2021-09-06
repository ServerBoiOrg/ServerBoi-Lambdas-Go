package main

import (
	"encoding/json"
	gu "generalutils"
	"log"
	"strings"
)

func component(eventBody string) (output InteractionOutput) {
	log.Printf("Component: %v", eventBody)

	//Unmarshal into ComponentInteraction Type
	var component gu.DiscordComponentInteraction
	json.Unmarshal([]byte(eventBody), &component)

	customSplit := strings.Split(component.Data.CustomID, ":")
	componentType := customSplit[0]

	log.Printf("Sending temporary response to Discord")
	gu.SendTempResponse(component.ID, component.Token)

	var response gu.DiscordInteractionResponseData
	var err error
	switch componentType {
	case "server":
		response = routeServerAction(component)
	}
	if err != nil {
		log.Printf("Error performing server command: %v", err)
		return InteractionOutput{}
	}
	log.Printf("Response from %v command: %v", componentType, response)

	return InteractionOutput{
		ApplicationID:    component.ApplicationID,
		InteractionToken: component.Token,
		Response:         response,
	}
}

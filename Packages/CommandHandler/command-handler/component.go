package main

import (
	dc "discordhttpclient"
	"encoding/json"
	"log"
	"strings"

	dt "github.com/awlsring/discordtypes"
)

func component(eventBody string) (output *dc.InteractionFollowupInput) {
	log.Printf("Component: %v", eventBody)

	//Unmarshal into ComponentInteraction Type
	var component *dt.Interaction
	json.Unmarshal([]byte(eventBody), &component)

	customSplit := strings.Split(component.Data.CustomID, ":")
	componentType := customSplit[0]

	log.Printf("Sending temporary response to Discord")
	client.TemporaryResponse(&dc.InteractionCallbackInput{
		InteractionID:    component.ID,
		InteractionToken: component.Token,
	})

	var response *dt.InteractionCallbackData
	var err error
	switch componentType {
	case "server":
		response = routeServerAction(component)
	}
	if err != nil {
		log.Printf("Error performing server command: %v", err)
		return &dc.InteractionFollowupInput{}
	}
	log.Printf("Response from %v command: %v", componentType, response)

	return &dc.InteractionFollowupInput{
		ApplicationID:    component.ApplicationID,
		InteractionToken: component.Token,
		Data:             response,
	}
}

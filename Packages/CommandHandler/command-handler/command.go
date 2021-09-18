package main

import (
	"encoding/json"
	"log"
	"time"

	dc "discordhttpclient"

	dt "github.com/awlsring/discordtypes"
)

func command(eventBody string) (output *dc.InteractionFollowupInput) {
	log.Printf("Command: %v", eventBody)

	//Unmarshal into Interaction Type
	var command *dt.Interaction
	json.Unmarshal([]byte(eventBody), &command)

	log.Printf("Sending temporary response to Discord")
	for {
		headers, _ := client.TemporaryResponse(&dc.InteractionCallbackInput{
			InteractionID:    command.ID,
			InteractionToken: command.Token,
		})
		if headers.StatusCode == 429 {
			log.Printf("Thottled, waiting")
			time.Sleep(time.Duration(headers.ResetAfter*1000) * time.Millisecond)
		}
		break
	}

	log.Printf("Command Option: %v", command.Data.Name)
	var response *dt.InteractionCallbackData
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

	return &dc.InteractionFollowupInput{
		ApplicationID:    command.ApplicationID,
		InteractionToken: command.Token,
		Data:             response,
	}
}

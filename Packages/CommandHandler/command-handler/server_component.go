package main

import (
	"fmt"
	gu "generalutils"
	"log"
	"strings"
)

func routeServerAction(component gu.DiscordComponentInteraction) (data gu.DiscordInteractionResponseData) {
	embed := component.Message.Embeds[0]
	serverID := strings.Trim(embed.Title[len(embed.Title)-6:], "()")
	customSplit := strings.Split(component.Data.CustomID, ":")
	action := customSplit[1]

	server, err := gu.GetServerFromID(serverID)
	if err != nil {

	}

	var message string
	switch action {
	case "start":
		err = server.Start()
		if err == nil {
			message = "Starting server"
		}
	case "stop":
		err = server.Stop()
		if err == nil {
			message = "Stopping server"
		}
	case "reboot":
		err = server.Restart()
		if err == nil {
			message = "Rebooting server"
		}
	}
	if err != nil {
		log.Printf("Error performing command: %v", err)
		message = fmt.Sprintf("Error running %v on server", action)
	}
	data = gu.DiscordInteractionResponseData{
		Content: message,
		Flags:   1 << 6,
	}

	return data
}

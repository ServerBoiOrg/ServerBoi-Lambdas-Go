package main

import (
	"fmt"
	gu "generalutils"
	"log"
	"strings"

	dt "github.com/awlsring/discordtypes"
)

func routeServerAction(component *dt.Interaction) (data *dt.InteractionCallbackData) {
	embed := component.Message.Embeds[0]
	serverID := strings.Trim(embed.Title[len(embed.Title)-6:], "()")
	customSplit := strings.Split(component.Data.CustomID, ":")
	action := customSplit[1]

	server, err := gu.GetServerFromID(serverID)
	if err != nil {

	}

	var authorized bool
	for _, user := range server.AuthorizedUsers() {
		if user == component.Member.User.ID {
			authorized = true
		}
	}
	for _, role := range server.AuthorizedRoles() {
		for _, userRole := range component.Member.Roles {
			if role == userRole {
				authorized = true
			}
		}
	}

	var message string
	if authorized {
		switch action {
		case "start":
			log.Printf("Starting server")
			err = server.Start()
			if err == nil {
				message = "Starting server"
			}
		case "stop":
			log.Printf("Sopping server")
			err = server.Stop()
			if err == nil {
				message = "Stopping server"
			}
		case "reboot":
			log.Printf("Rebooting server")
			err = server.Restart()
			if err == nil {
				message = "Rebooting server"
			}
		}
		if err != nil {
			log.Printf("Error performing command: %v", err)
			message = fmt.Sprintf("Error running %v on server", action)
		}
	} else {
		message = "You do not have authorization to run commands on this server."
	}
	log.Printf("Message to send to discord")
	return &dt.InteractionCallbackData{
		Content: message,
		Flags:   1 << 6,
	}
}

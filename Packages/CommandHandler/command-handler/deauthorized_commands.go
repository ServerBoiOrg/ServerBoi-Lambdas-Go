package main

import (
	"errors"
	"fmt"
	gu "generalutils"
	"log"
)

func routeDeauthorizeCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	deauthorizeCommand := command.Data.Options[0].Name
	deauthOptions := command.Data.Options[0].Options
	log.Printf("Deauthorize Commmad Option: %v", deauthorizeCommand)

	serverID := command.Data.Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := gu.GetServerFromID(serverID)
	if err != nil {
		return gu.FormResponseData(gu.FormResponseInput{
			"Content": fmt.Sprintf("Server %v can't be found.", serverID),
		})
	}
	log.Printf("Server Object: %s", server)
	log.Printf("Running %s on server %s", deauthorizeCommand, serverID)

	var authorized bool
	// Check if user is the owner
	if server.GetBaseService().OwnerID == command.Member.User.ID {
		authorized = true
	}
	// Check if user is in role
	for _, role := range command.Member.Roles {
		if role == server.GetBaseService().OwnerID {
			authorized = true
		}
	}
	var message string
	if authorized {
		switch deauthorizeCommand {
		case "user":
			log.Printf("Updating user auth for server")
			message = removeItemFromAuth(RemoveFromAuthInput{
				Type:          6,
				Server:        server,
				DeauthOptions: deauthOptions,
			})
		case "role":
			log.Printf("Updating user auth for server")
			message = removeItemFromAuth(RemoveFromAuthInput{
				Type:          8,
				Server:        server,
				DeauthOptions: deauthOptions,
			})
		default:
			message = fmt.Sprintf("Deauthorize command `%v` is unknown.", deauthorizeCommand)
		}
		if err != nil {
			message = fmt.Sprintf("Error performing command: %v", err)
		}
	} else {
		message = "Only owners can deauthorize others for this server."
	}
	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func removeFromList(list []string, item string) (items []string, err error) {
	for i, v := range list {
		if v == item {
			return append(list[:i], list[i+1:]...), nil
		}
	}
	return items, errors.New("Nothing removed")
}

type RemoveFromAuthInput struct {
	Type          int
	Server        gu.Server
	DeauthOptions []gu.DiscordApplicationCommandOption
}

func removeItemFromAuth(input RemoveFromAuthInput) string {
	var (
		id       string
		message  string
		typeName string
	)
	switch input.Type {
	case 6:
		typeName = "user"
	case 8:
		typeName = "role"
	}

	for _, option := range input.DeauthOptions {
		if option.Type == input.Type {
			id = option.Value
		}
	}
	if id == input.Server.GetBaseService().OwnerID {
		message = "The owner of a server can't be deauthorized."
	} else {
		roles, err := removeFromList(input.Server.AuthorizedRoles(), id)
		if err != nil {
			message = fmt.Sprintf("Specified %v isn't authorized on this server.", typeName)
		} else {
			updateAuthorization(input.Server.AuthorizedUsers(), roles, input.Server.GetBaseService().ServerID)
			message = "Authorization updated."
		}
	}
	return message
}

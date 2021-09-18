package main

import (
	"errors"
	"fmt"
	gu "generalutils"
	"log"

	dt "github.com/awlsring/discordtypes"
)

func routeDeauthorizeCommand(command *dt.Interaction) (response *dt.InteractionCallbackData) {
	deauthorizeCommand := command.Data.Options[0].Name
	deauthOptions := command.Data.Options[0].Options
	log.Printf("Deauthorize Commmad Option: %v", deauthorizeCommand)

	serverID := command.Data.Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := gu.GetServerFromID(serverID)
	if err != nil {
		return &dt.InteractionCallbackData{
			Content: fmt.Sprintf("Server %v can't be found.", serverID),
		}
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
	return &dt.InteractionCallbackData{
		Content: message,
	}
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
	DeauthOptions []*dt.ApplicationCommandInteractionDataOption
}

func removeItemFromAuth(input RemoveFromAuthInput) string {
	var (
		id       string
		message  string
		typeName string
		users    []string
		roles    []string
		err      error
	)
	for _, option := range input.DeauthOptions {
		if option.Type == input.Type {
			id = option.Value
		}
	}

	users = input.Server.AuthorizedUsers()
	roles = input.Server.AuthorizedRoles()

	switch input.Type {
	case 6:
		typeName = "user"
		users, err = removeFromList(users, id)
	case 8:
		typeName = "role"
		roles, err = removeFromList(roles, id)
	}

	if id == input.Server.GetBaseService().OwnerID {
		message = "The owner of a server can't be deauthorized."
	} else {
		if err != nil {
			message = fmt.Sprintf("Specified %v isn't authorized on this server.", typeName)
		} else {
			updateAuthorization(users, roles, input.Server.GetBaseService().ServerID)
			message = "Authorization updated."
		}
	}
	return message
}

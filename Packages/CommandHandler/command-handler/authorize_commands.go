package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func routeAuthorizeCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	authorizeCommand := command.Data.Options[0].Name
	authOptions := command.Data.Options[0].Options
	log.Printf("Authorize Commmad Option: %v", authorizeCommand)

	serverID := command.Data.Options[0].Options[0].Value
	log.Printf("Target Server: %v", serverID)
	server, err := gu.GetServerFromID(serverID)
	if err != nil {
		return gu.FormResponseData(gu.FormResponseInput{
			"Content": fmt.Sprintf("Server %v can't be found.", serverID),
		})
	}
	log.Printf("Server Object: %s", server)
	log.Printf("Running %s on server %s", authorizeCommand, serverID)

	var authorized bool
	for _, user := range server.AuthorizedUsers() {
		if user == command.Member.User.ID {
			authorized = true
		}
	}
	for _, role := range server.AuthorizedRoles() {
		for _, userRole := range command.Member.Roles {
			if role == userRole {
				authorized = true
			}
		}
	}

	var message string
	if authorized {
		switch authorizeCommand {
		//Server Actions
		case "user":
			log.Printf("Updating user auth for server")
			var userID string
			for _, option := range authOptions {
				if option.Type == 6 {
					userID = option.Value
				}
			}
			log.Printf("UserID to update %v", userID)
			users := server.AuthorizedUsers()
			users = append(users, userID)

			updateAuthorization(users, server.AuthorizedRoles(), serverID)

			message = "Authorization updated."
		case "role":
			var roleID string
			for _, option := range authOptions {
				if option.Type == 8 {
					roleID = option.Value
				}
			}
			roles := server.AuthorizedRoles()
			roles = append(roles, roleID)

			updateAuthorization(server.AuthorizedUsers(), roles, serverID)

			message = "Authorization updated."
		default:
			message = fmt.Sprintf("Server command `%v` is unknown.", authorizeCommand)
		}
		if err != nil {
			message = fmt.Sprintf("Error performing command: %v", err)
		}
	} else {
		message = "You do not have authorization to authorize others for this server."
	}
	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func updateAuthorization(users []string, roles []string, serverID string) {
	dynamo := gu.GetDynamo()
	table := gu.GetEnvVar("SERVER_TABLE")
	log.Printf("Updating authorization in server item in table %v", table)

	auth := map[string]dynamotypes.AttributeValue{
		"Users": &dynamotypes.AttributeValueMemberL{
			Value: buildAuthUsers(users),
		},
		"Roles": &dynamotypes.AttributeValueMemberL{
			Value: buildAuthRoles(roles),
		},
	}

	log.Printf("Updating item")
	resp, err := dynamo.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key: map[string]dynamotypes.AttributeValue{
			"ServerID": &types.AttributeValueMemberS{Value: serverID},
		},
		UpdateExpression: aws.String("set Authorized = :auth"),
		ExpressionAttributeValues: map[string]dynamotypes.AttributeValue{
			":auth": &dynamotypes.AttributeValueMemberM{Value: auth},
		},
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("%v", resp.ResultMetadata)
}

func buildAuthUsers(users []string) []dynamotypes.AttributeValue {
	log.Printf("Length of users: %v", len(users))
	var userValues []dynamotypes.AttributeValue
	for _, user := range users {
		item := &dynamotypes.AttributeValueMemberS{Value: user}
		userValues = append(userValues, item)
	}

	return userValues
}

func buildAuthRoles(roles []string) []dynamotypes.AttributeValue {
	log.Printf("Length of roles: %v", len(roles))
	var roleValues []dynamotypes.AttributeValue
	for _, role := range roles {
		item := &dynamotypes.AttributeValueMemberS{Value: role}
		roleValues = append(roleValues, item)
	}

	return roleValues
}

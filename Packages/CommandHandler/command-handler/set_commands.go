package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"log"

	"github.com/linode/linodego"
)

func routeSetCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	setCommand := command.Data.Options[0].Name
	setOptions := command.Data.Options[0].Options[0]
	log.Printf("Set Commmad Option: %v", setCommand)

	var message string
	switch setCommand {
	case "personal":
		personalCommand(setOptions, command.Member.User.ID)
	case "profile":
		profileCommand(setOptions, command.Member.Roles)
	default:
		message = fmt.Sprintf("Set command `%v` is unknown.", setCommand)
	}

	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func personalCommand(command gu.DiscordApplicationCommandOption, ownerID string) (response gu.DiscordInteractionResponseData) {
	personalCommand := command.Name
	personalOptions := command.Options
	log.Printf("Personal Commmad Option: %v", personalCommand)

	var message string
	switch personalCommand {
	case "aws":
		log.Printf("Service: AWS")
		accountID := personalOptions[0].Value
		log.Printf("Account to add: %v", accountID)
		message = setAWSItem(ownerID, accountID)
	case "linode":
		log.Printf("Service: Linode")
		apiKey := personalOptions[0].Value
		log.Printf("Adding Api Key")
		message = setLinodeItem(ownerID, apiKey)
	default:
		message = fmt.Sprintf("Personal command `%v` is unknown.", personalCommand)
	}

	log.Printf("Message to respond with: %v", message)
	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func profileCommand(command gu.DiscordApplicationCommandOption, roles []string) (response gu.DiscordInteractionResponseData) {
	profileCommand := command.Name
	profileOptions := command.Options

	var message string
	switch profileCommand {
	case "aws":
		log.Printf("Service: AWS")
		accountID, role := sortProfileOptionFields(profileOptions)
		if checkRoleIdInRoles(role, roles) {
			log.Printf("Account to add: %v", accountID)
			message = setAWSItem(role, accountID)
		} else {
			message = "You must be a member of the role to update it."
		}
	case "linode":
		log.Printf("Service: Linode")
		apiKey, role := sortProfileOptionFields(profileOptions)
		if checkRoleIdInRoles(role, roles) {
			log.Printf("Adding Api Key")
			message = setLinodeItem(role, apiKey)
		} else {
			message = "You must be a member of the role to update it."
		}
	default:
		message = fmt.Sprintf("Set command `%v` is unknown.", profileCommand)
	}

	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func sortProfileOptionFields(setOptions []gu.DiscordApplicationCommandOption) (accountItem string, role string) {
	for _, option := range setOptions {
		switch option.Type {
		case 3:
			accountItem = option.Value
		case 8:
			role = option.Value
		}
	}
	return accountItem, role
}

func setAWSItem(ownerID string, accountID string) string {
	log.Printf("Setting AWS Account for Owner %v", ownerID)
	err := gu.UpdateOwnerItem(gu.UpdateOwnerItemInput{
		OwnerID:    ownerID,
		FieldName:  "AWSAccountID",
		FieldValue: accountID,
	})
	if err != nil {
		log.Printf("Error putting Owner Item: %v", err)
		return "AWS account set for Profile."
	} else {
		return formAWSOnboardMessage(accountID)
	}
}

func setLinodeItem(ownerID string, apiKey string) string {
	err := testLinodeKey(apiKey)
	if err != nil {
		return "Unable to validate Api Key. Check the key's scopes and ensure the key has valid permissions."
	} else {
		err = gu.UpdateOwnerItem(gu.UpdateOwnerItemInput{
			OwnerID:    ownerID,
			FieldName:  "LinodeApiKey",
			FieldValue: apiKey,
		})
		if err != nil {
			log.Printf("Error putting Owner Item: %v", err)
			return "Unable to set Linode information."
		} else {
			return "Linode Api Key validated. You're good to go!"
		}
	}
}

func formAWSOnboardMessage(accountID string) string {
	objectUrl := "https://serverboi-resources-bucket.s3-us-west-2.amazonaws.com/onboardingCloudformation.json"
	url := fmt.Sprintf("https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/review?templateURL=%v&stackName=ServerBoiOnboardingRole", objectUrl)
	return fmt.Sprintf("To onboard your AWS Account: %v to ServerBoi, the proper resources must be created in your AWS Account\n\nUse the following link to perform a One-Click deployment.\n\n%v", accountID, url)
}

func testLinodeKey(apikey string) error {
	client := gu.CreateLinodeClient(apikey)
	_, err := client.ListRegions(context.Background(), &linodego.ListOptions{})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func checkRoleIdInRoles(roleID string, roles []string) bool {
	log.Printf("Checking if %v in roles", roleID)
	for _, role := range roles {
		if roleID == role {
			log.Printf("Role in roles, returning role")
			return true
		}
	}
	return false
}

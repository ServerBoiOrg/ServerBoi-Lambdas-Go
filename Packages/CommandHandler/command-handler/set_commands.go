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
		profileCommand(setOptions)
	default:
		message = fmt.Sprintf("Profile command `%v` is unknown.", setCommand)
	}

	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func personalCommand(command gu.DiscordApplicationCommandOption, ownerID string) (response gu.DiscordInteractionResponseData) {
	personalCommand := command.Options[0].Name
	personalOptions := command.Options[0].Options

	var message string
	switch personalCommand {
	case "aws":
		accountID := personalOptions[0].Value
		message = setAWSItem(ownerID, accountID)
	case "linode":
		apiKey := personalOptions[0].Value
		message = setLinodeItem(ownerID, apiKey)
	default:
		message = fmt.Sprintf("Set command `%v` is unknown.", personalCommand)
	}

	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

func profileCommand(command gu.DiscordApplicationCommandOption) (response gu.DiscordInteractionResponseData) {
	profileCommand := command.Options[0].Name
	profileOptions := command.Options[0].Options

	var message string
	switch profileCommand {
	case "aws":
		accountId, role := sortProfileOptionFields(profileOptions)
		message = setAWSItem(accountId, role)
	case "linode":
		apiKey, role := sortProfileOptionFields(profileOptions)
		message = setLinodeItem(role, apiKey)
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
		case 6:
			role = option.Value
		}
	}
	return accountItem, role
}

func setAWSItem(ownerID string, accountID string) string {
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

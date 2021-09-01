package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/linode/linodego"
)

func routeOnboardCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData, err error) {
	onboardCommand := command.Data.Options[0].Options[0].Name
	log.Printf("Onboard Commmad Option: %v", onboardCommand)

	var data gu.DiscordInteractionResponseData
	switch {
	//Server Actions
	case onboardCommand == "aws":
		data, err = onboardAWS(OnboardAWSInput{
			AccountID: command.Data.Options[0].Options[0].Options[0].Value,
			UserID:    command.Member.User.ID,
		})
	case onboardCommand == "linode":
		data, err = onboardLinode(OnboardLinodeInput{
			ApiKey: command.Data.Options[0].Options[0].Options[0].Value,
			UserID: command.Member.User.ID,
		})
	}
	return data, nil

}

func putOnboardItem(item map[string]types.AttributeValue, table string) error {
	dynamo := gu.GetDynamo()
	_, err := dynamo.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	return err
}

func onboardAWS(input OnboardAWSInput) (data gu.DiscordInteractionResponseData, err error) {
	var response gu.FormResponseInput
	table := gu.GetEnvVar("AWS_TABLE")

	item := map[string]types.AttributeValue{
		"UserID":       &types.AttributeValueMemberS{Value: input.UserID},
		"AWSAccountID": &types.AttributeValueMemberS{Value: input.AccountID},
	}
	putItemErr := putOnboardItem(item, table)
	if putItemErr != nil {
		log.Printf("Error putting user Item: %v", putItemErr)
		response = gu.FormResponseInput{
			"Content": "You already have an AWS account onboarded.",
		}
	}

	objectUrl := "https://serverboi-resources-bucket.s3-us-west-2.amazonaws.com/onboardingCloudformation.json"
	url := fmt.Sprintf("https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/review?templateURL=%v&stackName=ServerBoiOnboardingRole", objectUrl)
	message := fmt.Sprintf("To onboard your AWS Account: %v to ServerBoi, the proper resources must be created in your AWS Account\n\nUse the following link to perform a One-Click deployment.\n\n%v", input.AccountID, url)
	response = gu.FormResponseInput{
		"Content": message,
	}

	return gu.FormResponseData(response), nil
}

type OnboardAWSInput struct {
	AccountID string
	UserID    string
}

func onboardLinode(input OnboardLinodeInput) (data gu.DiscordInteractionResponseData, err error) {
	var message string
	table := gu.GetEnvVar("LINODE_TABLE")

	item := map[string]types.AttributeValue{
		"UserID": &types.AttributeValueMemberS{Value: input.UserID},
		"ApiKey": &types.AttributeValueMemberS{Value: input.ApiKey},
	}
	putItemErr := putOnboardItem(item, table)
	if putItemErr != nil {
		log.Printf("Error putting user Item: %v", putItemErr)
		message = "You already have an Api Key in use."
		return gu.FormResponseData(gu.FormResponseInput{
			"Content": message,
		}), nil
	}

	client := gu.CreateLinodeClient(input.ApiKey)
	_, getRegionErr := client.ListRegions(context.Background(), &linodego.ListOptions{})
	if getRegionErr != nil {
		message = "Unable to validate Api Key. Check the key's scopes and ensure the key has valid permissions."
	}

	message = "Api Key validated. You're good to go!"

	return gu.FormResponseData(gu.FormResponseInput{
		"Content": message,
	}), nil
}

type OnboardLinodeInput struct {
	ApiKey string
	UserID string
}

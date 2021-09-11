package main

import (
	"fmt"
	gu "generalutils"
	"log"
)

func routeRemoveCommand(command gu.DiscordInteractionApplicationCommand) (response gu.DiscordInteractionResponseData) {
	removeCommand := command.Data.Options[0].Name
	removeOptions := command.Data.Options[0].Options
	log.Printf("Remove Commmad Option: %v", removeCommand)

	var (
		err     error
		message string
		service string
	)
	switch removeCommand {
	case "personal":
		service = removeOptions[0].Value
		switch service {
		case "aws":
			err = gu.RemoveFieldFromOwnerItem(gu.UpdateOwnerItemInput{
				OwnerID:   command.Member.User.ID,
				FieldName: "AWSAccountID",
			})
		case "linode":
			err = gu.RemoveFieldFromOwnerItem(gu.UpdateOwnerItemInput{
				OwnerID:   command.Member.User.ID,
				FieldName: "LinodeApiKey",
			})
		default:
			message = fmt.Sprintf("Option %v is not supported", service)
		}
	case "profile":
		var roleID string
		for _, field := range removeOptions {
			switch field.Type {
			case 3:
				service = field.Value
			case 8:
				roleID = field.Value
			}
		}
		if checkRoleIdInRoles(roleID, command.Member.Roles) {
			switch service {
			case "aws":
				err = gu.RemoveFieldFromOwnerItem(gu.UpdateOwnerItemInput{
					OwnerID:   roleID,
					FieldName: "AWSAccountID",
				})
			case "linode":
				err = gu.RemoveFieldFromOwnerItem(gu.UpdateOwnerItemInput{
					OwnerID:   roleID,
					FieldName: "LinodeApiKey",
				})
			default:
				message = fmt.Sprintf("Option %v is not supported", service)
			}
		} else {
			message = "You must be a member of the role to update it."
		}
	default:
		message = fmt.Sprintf("Remove command `%v` is unknown.", removeCommand)
	}
	if err != nil {
		message = "Error updating item"
	} else {
		message = fmt.Sprintf("Field for %v removed from OwnerItem", service)
	}

	formRespInput := gu.FormResponseInput{
		"Content": message,
	}
	return gu.FormResponseData(formRespInput)
}

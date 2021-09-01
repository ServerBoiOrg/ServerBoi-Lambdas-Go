package main

import (
	gu "generalutils"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func provisionLinode(params ProvisonServerParameters) map[string]dynamotypes.AttributeValue {

	server := gu.LinodeServer{
		OwnerID:     params.OwnerID,
		Owner:       params.Owner,
		Application: params.Application,
		ServerName:  params.ServerName,
	}

	return formLinodeServerItem(server)
}

func formLinodeServerItem(server gu.LinodeServer) map[string]dynamotypes.AttributeValue {
	serverItem := formBaseServerItem(
		server.OwnerID, server.Owner, server.Application, server.ServerName, server.Port,
	)

	serverItem["Location"] = &dynamotypes.AttributeValueMemberS{Value: server.Location}
	serverItem["ApiKey"] = &dynamotypes.AttributeValueMemberS{Value: server.ApiKey}
	serverItem["LinodeID"] = &dynamotypes.AttributeValueMemberS{Value: server.LinodeID}
	serverItem["LinodeType"] = &dynamotypes.AttributeValueMemberS{Value: server.LinodeType}

	return serverItem
}

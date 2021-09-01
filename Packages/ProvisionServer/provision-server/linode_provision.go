package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"log"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/linode/linodego"
)

func provisionLinode(params ProvisonServerParameters) map[string]dynamotypes.AttributeValue {
	log.Printf("Querying aws account for %v item from Dynamo", params.Owner)
	apiKey := queryLinodeApiKey(params.OwnerID)

	// Generics actions for each server
	region := params.CreationOptions["region"]
	architecture := getArchitecture(params.CreationOptions)
	log.Printf("Getting build info")
	buildInfo := getBuildInfo(params.Application)
	container := getContainer(buildInfo, architecture)
	serverID := formServerID()
	log.Printf("Generating bootscript")
	bootscript := formBootscript(
		FormDockerCommandInput{
			Application:      params.Application,
			Url:              params.Url,
			InteractionToken: params.InteractionToken,
			InteractionID:    params.InteractionID,
			ApplicationID:    params.ApplicationID,
			ExecutionName:    params.ExecutionName,
			ServerName:       params.ServerName,
			ServerID:         serverID,
			GuildID:          params.GuildID,
			Container:        container,
			EnvVar:           params.CreationOptions,
		},
		buildInfo.DockerCommands,
	)

	linodeType := getLinodeType(buildInfo, architecture)
	image := "linode/debian11"

	client := gu.CreateLinodeClient(apiKey)

	// This creates a lot over time, find a way to clean these up
	response, _ := client.CreateStackscript(context.Background(), linodego.StackscriptCreateOptions{
		Label:  fmt.Sprintf("ServerBoi-%v", params.Application),
		Images: []string{image},
		Script: bootscript,
	})
	scriptID := response.ID

	client.CreateInstance(context.Background(), linodego.InstanceCreateOptions{
		Region:        region,
		Type:          linodeType,
		Image:         image,
		Label:         params.ServerName,
		StackScriptID: scriptID,
	})

	server := gu.LinodeServer{
		OwnerID:     params.OwnerID,
		Owner:       params.Owner,
		Application: params.Application,
		ServerName:  params.ServerName,
	}

	return formLinodeServerItem(server)
}

func getLinodeType(buildInfo BuildInfo, architecture string) string {
	var archInfo ArchitectureInfo
	var defaultType string
	switch architecture {
	case "x86":
		archInfo = buildInfo.X86
		defaultType = "g6-standard-2"
	default:
		panic("Unknown architecture")
	}

	if buildLinodeType, ok := archInfo.InstanceType["linode"]; ok {
		client := gu.CreateAuthlessLinodeClient()
		response, _ := client.ListTypes(context.Background(), &linodego.ListOptions{})
		for _, linodeType := range response {
			if linodeType.ID == buildLinodeType {
				return buildLinodeType
			}
		}
		panic("Unable to find linode type")
	} else {
		log.Printf("Linode Type: %v", defaultType)
		return defaultType
	}
}

func formLinodeServerItem(server gu.LinodeServer) map[string]dynamotypes.AttributeValue {
	serverItem := formBaseServerItem(
		server.OwnerID, server.Owner, server.Application, server.ServerName, server.Port, server.ServerID,
	)

	serverItem["Location"] = &dynamotypes.AttributeValueMemberS{Value: server.Location}
	serverItem["ApiKey"] = &dynamotypes.AttributeValueMemberS{Value: server.ApiKey}
	serverItem["LinodeID"] = &dynamotypes.AttributeValueMemberN{Value: string(server.LinodeID)}
	serverItem["LinodeType"] = &dynamotypes.AttributeValueMemberS{Value: server.LinodeType}

	return serverItem
}

type LinodeTableResponse struct {
	UserID string `json:"UserID"`
	ApiKey string `json:"ApiKey"`
}

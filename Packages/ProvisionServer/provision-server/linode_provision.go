package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/linode/linodego"
)

func provisionLinode(params *ProvisonServerParameters) *ProvisionOutput {
	log.Printf("Querying aws account for %v item from Dynamo", params.Owner)
	ownerItem, err := gu.GetOwnerItem(params.OwnerID)
	if err != nil {
		log.Fatalf("Unable to get owner item")
	}
	apiKey := ownerItem.LinodeApiKey

	output := genericProvision(params)

	image := "linode/debian11"
	client := gu.CreateLinodeClient(apiKey)

	log.Printf("Creating Stackscript")
	// This creates a lot over time, find a way to clean these up
	response, stackErr := client.CreateStackscript(context.Background(), linodego.StackscriptCreateOptions{
		Label:  fmt.Sprintf("ServerBoi-%v", params.Application),
		Images: []string{image},
		Script: output.Bootscript,
	})
	if stackErr != nil {
		log.Fatalf("Error creating Stackscript: %v", stackErr)
	}
	scriptID := response.ID

	createResp, createErr := client.CreateInstance(context.Background(), linodego.InstanceCreateOptions{
		Region:         params.Region,
		Type:           output.HardwareType,
		Image:          image,
		Label:          output.ServerID,
		StackScriptID:  scriptID,
		AuthorizedKeys: []string{strings.TrimSpace(output.PublicKey)},
		RootPass:       generateUselessPassword(),
	})
	if createErr != nil {
		log.Fatalf("Error creating Linode: %v", createErr)
	}

	var authorized *gu.Authorized
	if params.IsRole {
		authorized = &gu.Authorized{
			Roles: []string{params.OwnerID},
		}
	} else {
		authorized = &gu.Authorized{
			Users: []string{params.OwnerID},
		}
	}

	server := gu.LinodeServer{
		OwnerID:     params.OwnerID,
		Owner:       params.Owner,
		Application: params.Application,
		ServerName:  params.Name,
		Port:        output.Configuration.ClientPort,
		QueryPort:   output.Configuration.QueryPort,
		QueryType:   output.Configuration.QueryType,
		Service:     "linode",
		ServerID:    output.ServerID,
		LinodeID:    createResp.ID,
		ApiKey:      apiKey,
		LinodeType:  output.HardwareType,
		Location:    params.Region,
		Authorized:  authorized,
	}

	return &ProvisionOutput{
		ServerID:         output.ServerID,
		PrivateKeyObject: output.PrivateKeyObject,
		ServerItem:       formLinodeServerItem(server),
	}
}

func getLinodeType(override string, configuration *ApplicationConfiguration, architecture string) string {
	log.Printf("Getting Linode Type")
	var archInfo *ArchitectureConfiguration
	var defaultType string
	switch architecture {
	case "x86":
		archInfo = configuration.X86
		defaultType = "g6-standard-2"
	default:
		panic("Unknown architecture")
	}
	client := gu.CreateAuthlessLinodeClient()
	response, _ := client.ListTypes(context.Background(), &linodego.ListOptions{})

	if override != "" {
		for _, linodeType := range response {
			if linodeType.ID == override {
				log.Printf("Linode Type: %v", override)
				return override
			}
		}
	}

	if buildLinodeType, ok := archInfo.InstanceType["linode"]; ok {
		for _, linodeType := range response {
			if linodeType.ID == buildLinodeType {
				log.Printf("Linode Type: %v", buildLinodeType)
				return buildLinodeType
			}
		}
		panic("Unable to find linode type")
	} else {
		log.Printf("Linode Type: %v", defaultType)
		return defaultType
	}
}

func generateUselessPassword() string {
	//Generate and forget since Linode insists on a root password
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 20)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func formLinodeServerItem(server gu.LinodeServer) map[string]dynamotypes.AttributeValue {
	serverItem := formBaseServerItem(
		server.OwnerID,
		server.Owner,
		server.Application,
		server.ServerName,
		server.Service,
		server.Port,
		server.QueryPort,
		server.QueryType,
		server.ServerID,
		server.Authorized,
	)
	serverItem["Location"] = &dynamotypes.AttributeValueMemberS{Value: server.Location}
	serverItem["ApiKey"] = &dynamotypes.AttributeValueMemberS{Value: server.ApiKey}
	serverItem["LinodeID"] = &dynamotypes.AttributeValueMemberN{Value: strconv.Itoa(server.LinodeID)}
	serverItem["LinodeType"] = &dynamotypes.AttributeValueMemberS{Value: server.LinodeType}

	return serverItem
}

type LinodeTableResponse struct {
	UserID string `json:"UserID"`
	ApiKey string `json:"ApiKey"`
}

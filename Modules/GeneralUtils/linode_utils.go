package generalutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func CreateLinodeClient(apiKey string) linodego.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	return linodego.NewClient(oauth2Client)
}

func CreateAuthlessLinodeClient() linodego.Client {
	return linodego.NewClient(&http.Client{})
}

func (server LinodeServer) Start() (data DiscordInteractionResponseData, err error) {
	client := CreateLinodeClient(server.ApiKey)

	err = client.BootInstance(context.Background(), server.LinodeID, 0)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Starting server",
	}

	return FormResponseData(formRespInput), nil
}

func (server LinodeServer) Stop() (data DiscordInteractionResponseData, err error) {
	client := CreateLinodeClient(server.ApiKey)

	err = client.ShutdownInstance(context.Background(), server.LinodeID)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Stopping server",
	}

	return FormResponseData(formRespInput), nil
}

func (server LinodeServer) Restart() (data DiscordInteractionResponseData, err error) {
	client := CreateLinodeClient(server.ApiKey)

	err = client.RebootInstance(context.Background(), server.LinodeID, 0)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Restarting server",
	}

	return FormResponseData(formRespInput), nil
}

func (server LinodeServer) Status() (data DiscordInteractionResponseData, err error) {
	client := CreateLinodeClient(server.ApiKey)
	instance, err := client.GetInstance(context.Background(), server.LinodeID)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": fmt.Sprintf("Server status: %v", instance.Status),
	}

	return FormResponseData(formRespInput), nil
}

func (server LinodeServer) GetService() string {
	return server.Service
}

func (server LinodeServer) GetBaseService() BaseServer {
	return BaseServer{
		ServerID:    server.ServerID,
		Application: server.Application,
		ServerName:  server.ServerName,
		Service:     server.Service,
		Owner:       server.Owner,
		OwnerID:     server.OwnerID,
		Port:        server.Port,
	}
}

func (server LinodeServer) GetServerBoiRegion() ServerBoiRegion {
	return FormServerBoiRegion(server.Service, server.Location)
}

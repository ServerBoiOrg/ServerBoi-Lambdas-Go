package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func createLinodeClient(apiKey string) linodego.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	return linodego.NewClient(oauth2Client)
}

func (server LinodeServer) start() (data DiscordInteractionResponseData, err error) {
	client := createLinodeClient(server.ServiceInfo.ApiKey)

	err = client.BootInstance(context.Background(), server.ServiceInfo.LinodeID, 0)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Starting server",
	}

	return formResponseData(formRespInput), nil
}

func (server LinodeServer) stop() (data DiscordInteractionResponseData, err error) {
	client := createLinodeClient(server.ServiceInfo.ApiKey)

	err = client.ShutdownInstance(context.Background(), server.ServiceInfo.LinodeID)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Stopping server",
	}

	return formResponseData(formRespInput), nil
}

func (server LinodeServer) restart() (data DiscordInteractionResponseData, err error) {
	client := createLinodeClient(server.ServiceInfo.ApiKey)

	err = client.RebootInstance(context.Background(), server.ServiceInfo.LinodeID, 0)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": "Restarting server",
	}

	return formResponseData(formRespInput), nil
}

func (server LinodeServer) status() (data DiscordInteractionResponseData, err error) {
	client := createLinodeClient(server.ServiceInfo.ApiKey)
	instance, err := client.GetInstance(context.Background(), server.ServiceInfo.LinodeID)
	if err != nil {
		fmt.Println(err)
		return data, err
	}

	formRespInput := FormResponseInput{
		"Content": fmt.Sprintf("Server status: %v", instance.Status),
	}

	return formResponseData(formRespInput), nil
}

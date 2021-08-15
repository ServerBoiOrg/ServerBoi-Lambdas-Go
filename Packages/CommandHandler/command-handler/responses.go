package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	embed "github.com/clinet/discordgo-embed"
)

func sendTempResponse(interactionID string, interactionToken string) {
	responseUrl := fmt.Sprintf("https://discord.com/api/v8/interactions/%s/%s/callback", interactionID, interactionToken)

	tempResponse := DiscordInteractionResponse{
		Type: 5,
		Data: DiscordInteractionResponseData{
			Flags: 1 << 6,
		},
	}

	responseBody, _ := json.Marshal(tempResponse)
	log.Printf("Temp response: %v", string(responseBody))
	bytes := bytes.NewBuffer(responseBody)

	http.Post(responseUrl, "application/json", bytes)
}

func editResponse(applicationID string, interactionToken string, data DiscordInteractionResponseData) {
	responseUrl := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s/messages/@original", applicationID, interactionToken)
	log.Printf("URL to Patch: %v", responseUrl)

	log.Printf("Editing response with: %v", data)
	responseBody, _ := json.Marshal(data)
	bytes := bytes.NewBuffer(responseBody)

	request, err := http.NewRequest(http.MethodPatch, responseUrl, bytes)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from discord: %v", resp)

	defer resp.Body.Close()
}

type FormResponseInput map[string]interface{}

func formResponseData(input FormResponseInput) (data DiscordInteractionResponseData) {
	log.Printf("Forming interaction response data")
	data = DiscordInteractionResponseData{
		Flags: 1 << 6,
	}

	if content, ok := input["Content"]; ok {
		log.Printf("Adding content to data")
		data.Content = content.(string)
	}

	if embeds, ok := input["Embeds"]; ok {
		log.Printf("Adding embeds to data")

		e := embeds.(*embed.Embed)
		data.Embeds = []embed.Embed{*e}
	}

	if components, ok := input["Components"]; ok {
		log.Printf("Adding components to data")
		data.Components = components.([]DiscordComponentData)
	}

	log.Printf("Formed Response Data: %v", data)
	return data
}

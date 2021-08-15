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
	bytes := bytes.NewBuffer(responseBody)

	http.Post(responseUrl, "application/json", bytes)
}

func editResponse(applicationID string, interactionToken string, data DiscordInteractionResponse) {
	responseUrl := fmt.Sprintf("https://discord.com/api/v8/interactions/%s/%s/messages/@original", applicationID, interactionToken)

	responseBody, _ := json.Marshal(data)
	bytes := bytes.NewBuffer(responseBody)

	request, err := http.NewRequest(http.MethodPatch, responseUrl, bytes)
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
}

type FormResponseInput map[string]interface{}

func formResponseData(input FormResponseInput) (data *DiscordInteractionResponseData) {

	data = &DiscordInteractionResponseData{
		Flags: 1 << 6,
	}

	if content, ok := input["Content"]; ok {
		data.Content = content.(string)
	}

	if embeds, ok := input["Embeds"]; ok {
		data.Embeds = embeds.([]embed.Embed)
	}

	if components, ok := input["Embeds"]; ok {
		data.Components = components.([]DiscordComponentData)
	}

	return data
}

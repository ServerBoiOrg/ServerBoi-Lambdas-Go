package generalutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type DiscordClient struct {
	BotToken   string
	ApiVersion string
	baseUrl    string
}

type CreateDiscordClientInput struct {
	BotToken   string
	ApiVersion string
}

func CreateDiscordClient(input CreateDiscordClientInput) DiscordClient {
	url := fmt.Sprintf("https://discord.com/api/%s", input.ApiVersion)
	return DiscordClient{
		BotToken:   input.BotToken,
		ApiVersion: input.ApiVersion,
		baseUrl:    url,
	}
}

func (client DiscordClient) EditMessage(
	channeldID string,
	messageID string,
	data DiscordInteractionResponseData,
) (message discordgo.Message, err error) {
	messageUrl := fmt.Sprintf("%s/channels/%s/messages/%s", client.baseUrl, channeldID, messageID)

	responseBody, _ := json.Marshal(data)
	bytes := bytes.NewBuffer(responseBody)
	log.Printf("Editing response with: %v", bytes)

	req, err := http.NewRequest("PATCH", messageUrl, bytes)
	if err != nil {
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
	}
	log.Printf("Request made: %v", resp)
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&message)
		return message, nil
	}
	return message, err

}

func (client DiscordClient) CreateMessage(channelID string, data DiscordInteractionResponseData) (message discordgo.Message, err error) {
	url := fmt.Sprintf("%s/channels/%s/messages", client.baseUrl, channelID)
	responseBody, _ := json.Marshal(data)
	log.Printf(string(responseBody))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(responseBody))
	if err != nil {
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&message)
		return message, nil
	}
	return message, err
}

func (client DiscordClient) GetChannelMessages(channeldID string) (messages []discordgo.Message, err error) {
	messageUrl := fmt.Sprintf("%s/channels/%s/messages", client.baseUrl, channeldID)

	req, err := http.NewRequest("GET", messageUrl, nil)
	if err != nil {
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&messages)
		return messages, nil
	}
	return messages, err

}

func (client DiscordClient) DeleteMessage(channelID string, messageID string) (err error) {
	messageUrl := fmt.Sprintf("%s/channels/%s/messages/%s", client.baseUrl, channelID, messageID)

	req, err := http.NewRequest("DELETE", messageUrl, nil)
	if err != nil {
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return err

}

package discordhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	dt "github.com/awlsring/discordtypes"
)

func CreateClient(input *CreateClientInput) *Client {
	base := fmt.Sprintf("https://discord.com/api")
	url := fmt.Sprintf("%s/%s", base, input.ApiVersion)
	webhookUrl := fmt.Sprintf("%s/webhooks", base)
	interactionUrl := fmt.Sprintf("%s/%s/interactions", base, input.ApiVersion)
	return &Client{
		BotToken:       input.BotToken,
		ApiVersion:     input.ApiVersion,
		url:            url,
		webhookUrl:     webhookUrl,
		interactionUrl: interactionUrl,
		http:           &http.Client{},
	}
}

func (client *Client) EditMessage(input *EditInteractionMessageInput) (message *dt.Message, headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s", client.url, input.ChannelID, input.MessageID)
	req, err := http.NewRequest("PATCH", url, convertData(input.Data))
	if err != nil {
		return message, headers, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.http.Do(req)
	if err != nil {
		return message, headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&message)
		return message, headers, nil
	}
	return message, headers, err
}

func (client *Client) CreateMessage(input *CreateMessageInput) (message *dt.Message, headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/channels/%s/messages", client.url, input.ChannelID)
	req, err := http.NewRequest("POST", url, convertData(input.Data))
	if err != nil {
		return message, headers, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return message, headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&message)
		return message, headers, nil
	}
	return message, headers, err
}

func (client *Client) GetChannelMessages(channeldID string) (messages []*dt.Message, headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/channels/%s/messages", client.url, channeldID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return messages, headers, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return messages, headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&messages)
		return messages, headers, nil
	}
	return messages, headers, err

}

func (client *Client) DeleteMessage(input *DeleteMessageInput) (headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s", client.url, input.ChannelID, input.MessageID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return headers, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", client.BotToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		return headers, nil
	}
	return headers, err
}

func (client *Client) TemporaryResponse(input *InteractionCallbackInput) (headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/%s/%s/callback", client.interactionUrl, input.InteractionID, input.InteractionToken)
	response := dt.InteractionResponse{
		Type: 5,
		Data: &dt.InteractionCallbackData{
			Flags: 1 << 6,
		},
	}

	req, err := http.NewRequest("POST", url, convertData(response))
	if err != nil {
		return headers, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		return headers, nil
	}
	return headers, err
}

func (client *Client) EditInteractionResponse(input *InteractionFollowupInput) (message *dt.Message, headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/%s/%s/messages/@original", client.webhookUrl, input.ApplicationID, input.InteractionToken)
	req, err := http.NewRequest("PATCH", url, convertData(input.Data))
	if err != nil {
		return message, headers, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return message, headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&message)
		return message, headers, nil
	}
	return message, headers, err
}

func (client *Client) PostInteractionFollowUp(input *InteractionFollowupInput) (message *dt.Message, headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/%s/%s", client.webhookUrl, input.ApplicationID, input.InteractionToken)
	req, err := http.NewRequest("POST", url, convertData(input.Data))
	if err != nil {
		return message, headers, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return message, headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&message)
		return message, headers, nil
	}
	return message, headers, err
}

func (client *Client) DeleteInteractionResponse(input *InteractionFollowupInput) (headers *DiscordHeaders, err error) {
	url := fmt.Sprintf("%s/%s/%s/messages/@original", client.webhookUrl, input.ApplicationID, input.InteractionToken)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return headers, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return headers, err
	}
	defer resp.Body.Close()
	headers = formDiscordHeaders(resp)
	if resp.StatusCode == http.StatusOK {
		return headers, nil
	}
	return headers, err
}

func formDiscordHeaders(resp *http.Response) *DiscordHeaders {
	h := DiscordHeaders{
		StatusCode: resp.StatusCode,
	}
	for header, value := range resp.Header {
		switch header {
		case "X-Ratelimit-Limit":
			h.Limit, _ = strconv.Atoi(value[0])
		case "X-Ratelimit-Remaining":
			h.Remaining, _ = strconv.Atoi(value[0])
		case "X-Ratelimit-Reset":
			reset, _ := strconv.ParseInt(value[0][:10], 10, 64)
			h.Reset = reset + 1
		case "X-Ratelimit-Reset-After":
			h.ResetAfter, _ = strconv.ParseFloat(value[0], 8)
		case "X-Ratelimit-Bucket":
			h.Bucket = value[0]
		}
	}
	return &h
}

func convertData(data interface{}) *bytes.Buffer {
	body, _ := json.Marshal(data)
	return bytes.NewBuffer(body)
}

package main

import (
	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"
	"strings"

	dt "github.com/awlsring/discordtypes"
)

func editServerMessage(channelID string, messageID string, data *dt.EditMessageData) (err error) {
	for {
		_, headers, err := client.EditMessage(&dc.EditInteractionMessageInput{
			ChannelID: channelID,
			MessageID: messageID,
			Data:      data,
		})
		if err != nil {
			return err
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			break
		}
	}
	return err
}

func updateEmbed(server gu.Server) (embed *dt.Embed, components []*dt.Component, err error) {
	status, err := server.GetStatus()
	if err != nil {
		return embed, components, err
	}
	state, _, err := ru.TranslateState(
		server.GetBaseService().Service,
		status,
	)
	if err != nil {
		return embed, components, err
	}
	var running bool
	if strings.Contains(state, "Running") {
		running = true
	} else {
		running = false
	}
	serverInfo := server.GetBaseService()
	ip, err := server.GetIPv4()
	if err != nil {
		return embed, components, err
	}

	embed = ru.CreateServerEmbed(ru.GetServerData(&ru.GetServerDataInput{
		Name:        serverInfo.ServerName,
		ID:          serverInfo.ServerID,
		IP:          ip,
		Status:      status,
		Region:      serverInfo.Region,
		Port:        serverInfo.Port,
		Application: serverInfo.Application,
		Owner:       serverInfo.Owner,
		Service:     serverInfo.Service,
	}))
	components = ru.ServerEmbedComponents(running)

	return embed, components, nil

}

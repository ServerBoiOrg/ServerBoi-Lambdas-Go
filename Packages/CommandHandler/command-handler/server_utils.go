package main

import (
	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"
	sq "serverquery"
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
	state, emoji, err := ru.GetStatus(&ru.GetStatusInput{
		Service: server.GetBaseService().Service,
		Status:  status,
	})
	if err != nil {
		return embed, components, err
	}
	var running bool
	if strings.Contains(state, "Running") {
		running = true
	} else {
		running = false
	}
	ip, err := server.GetIPv4()
	if err != nil {
		return embed, components, err
	}
	info, err := sq.ServerDataQuery(ip)
	if err != nil {
		return embed, components, err
	}

	embed = ru.CreateServerEmbed(ru.FormEmbedData(&ru.FormEmbedDataInput{
		Name:        info.General.Name,
		ID:          info.General.ID,
		IP:          ip,
		Port:        info.General.ClientPort,
		Status:      status,
		StatusEmoji: emoji,
		Region:      info.ServiceInfo.Region,
		Application: info.General.Application,
		Owner:       info.General.OwnerName,
		Service:     info.ServiceInfo.Provider,
	}))
	components = ru.ServerEmbedComponents(running)

	return embed, components, nil

}

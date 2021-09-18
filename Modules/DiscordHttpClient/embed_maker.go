package discordhttpclient

import (
	dt "github.com/awlsring/discordtypes"
)

/*
Inspired by https://github.com/Clinet/discordgo-embed
A lot of functionality from Clinet, but allows you to expose the Embed to use in the embed type.
Also updated with Embed functionality as of 10/2021
*/
type EmbedMaker struct {
	Embed *dt.Embed
}

func NewEmbedMaker() *EmbedMaker {
	return &EmbedMaker{Embed: &dt.Embed{}}
}

func (embed *EmbedMaker) SetTitle(name string) *EmbedMaker {
	embed.Embed.Title = name
	return embed
}

func (embed *EmbedMaker) SetDescription(description string) *EmbedMaker {
	if len(description) > 2048 {
		description = description[:2048]
	}
	embed.Embed.Description = description
	return embed
}

func (embed *EmbedMaker) SetAuthor(name string, url string, iconUrl string, proxyIconUrl string) *EmbedMaker {
	embed.Embed.Author = &dt.EmbedAuthor{
		Name:         name,
		URL:          url,
		IconURL:      iconUrl,
		ProxyIconURL: proxyIconUrl,
	}
	return embed
}

func (embed *EmbedMaker) SetFooter(text string, iconUrl string, proxyIconUrl string) *EmbedMaker {
	embed.Embed.Footer = &dt.EmbedFooter{
		IconURL:      iconUrl,
		Text:         text,
		ProxyIconURL: proxyIconUrl,
	}
	return embed
}

func (embed *EmbedMaker) SetColor(color int) *EmbedMaker {
	embed.Embed.Color = color
	return embed
}

func (embed *EmbedMaker) SetTimestamp(timestamp string) *EmbedMaker {
	embed.Embed.Timestamp = timestamp
	return embed
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-types
func (embed *EmbedMaker) SetType(embedType string) *EmbedMaker {
	embed.Embed.Type = embedType
	return embed
}

func (embed *EmbedMaker) SetUrl(url string) *EmbedMaker {
	embed.Embed.URL = url
	return embed
}

func (embed *EmbedMaker) SetProvider(name string, url string) *EmbedMaker {
	embed.Embed.Provider = &dt.EmbedProvider{
		Name: name,
		URL:  url,
	}
	return embed
}

func (embed *EmbedMaker) SetImage(url string, proxyUrl string, height int, width int) *EmbedMaker {
	embed.Embed.Image = &dt.EmbedImage{
		URL:      url,
		ProxyURL: proxyUrl,
		Height:   height,
		Width:    width,
	}
	return embed
}

func (embed *EmbedMaker) SetVideo(url string, proxyUrl string, height int, width int) *EmbedMaker {
	embed.Embed.Video = &dt.EmbedVideo{
		URL:      url,
		ProxyURL: proxyUrl,
		Width:    width,
		Height:   height,
	}
	return embed
}

func (embed *EmbedMaker) SetThumbnail(url string, proxyUrl string, height int, width int) *EmbedMaker {
	embed.Embed.Thumbnail = &dt.EmbedThumbnail{
		URL:      url,
		ProxyURL: proxyUrl,
		Width:    width,
		Height:   height,
	}
	return embed
}

func (embed *EmbedMaker) AddField(name string, value string, inline bool) *EmbedMaker {
	fields := make([]*dt.EmbedField, 0)

	if len(name) > 256 {
		name = name[:256]
	}

	if len(value) > 1024 {
		i := 1024
		extended := false
		for i = 1024; i < len(value); {
			if i != 1024 && extended == false {
				name += " (extended)"
				extended = true
			}
			if value[i] == []byte(" ")[0] || value[i] == []byte("\n")[0] || value[i] == []byte("-")[0] {
				fields = append(fields, &dt.EmbedField{
					Name:   name,
					Value:  value[i-1024 : i],
					Inline: inline,
				})
			} else {
				fields = append(fields, &dt.EmbedField{
					Name:   name,
					Value:  value[i-1024:i-1] + "-",
					Inline: inline,
				})
				i--
			}

			if (i + 1024) < len(value) {
				i += 1024
			} else {
				break
			}
		}
		if i < len(value) {
			name += " (extended)"
			fields = append(fields, &dt.EmbedField{
				Name:   name,
				Value:  value[i:],
				Inline: inline,
			})
		}
	} else {
		fields = append(fields, &dt.EmbedField{
			Name:   name,
			Value:  value,
			Inline: inline,
		})
	}
	embed.Embed.Fields = append(embed.Embed.Fields, fields...)
	return embed
}

package generalutils

import (
	"encoding/json"
	"time"

	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

type Unpacker struct {
	Data interface{}
}

func (u *Unpacker) UnmarshalJSON(bytes []byte) error {
	unmarshalData := &DiscordInteractionApplicationCommand{}
	err := json.Unmarshal(bytes, unmarshalData)

	// no error, but we also need to make sure we unmarshaled something
	if err == nil && unmarshalData.Type == 2 {
		u.Data = unmarshalData
		return nil
	}

	// abort if we have an error other than the wrong type
	if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
		return err
	}

	unmarshalData2 := &DiscordInteractionComponentCommand{}
	err = json.Unmarshal(bytes, unmarshalData2)
	if err != nil {
		return err
	}

	u.Data = unmarshalData2
	return nil
}

type DiscordMessageRequestData struct {
	Content         string                        `json:"content,omitempty"`
	Embeds          []embed.Embed                 `json:"embeds,omitempty"`
	Flags           int                           `json:"flags,omitempty"`
	File            []byte                        `json:"file,omitempty"`
	PayloadJSON     string                        `json:"payload_json,omitempty"`
	AllowedMentions discordgo.AllowedMentionType  `json:"allowed_mentions,omitempty"`
	Attachments     []discordgo.MessageAttachment `json:"attachments,omitempty`
	Components      []DiscordComponentData        `json:"components,omitempty"`
}

type DiscordInteractionApplicationCommand struct {
	ID            string                        `json:"id"`
	ApplicationID string                        `json:"application_id"`
	Type          int                           `json:"type"`
	Data          DiscordApplicationCommandData `json:"data"`
	GuildID       string                        `json:"guild_id"`
	ChannelID     string                        `json:"channel_id"`
	Member        DiscordMember                 `json:"member"`
	Token         string                        `json:"token"`
	Version       int                           `json:"version"`
}

type DiscordApplicationCommandData struct {
	ID      string                            `json:"id,omitempty"`
	Name    string                            `json:"name"`
	Options []DiscordApplicationCommandOption `json:"options,omitempty"`
}

type DiscordApplicationCommandOption struct {
	Type        int                                     `json:"type"`
	Name        string                                  `json:"name"`
	Description string                                  `json:"description,omitempty"`
	Required    bool                                    `json:"required,omitempty"`
	Value       string                                  `json:"value,omitempty"`
	Choices     []DiscordApplicationCommandOptionChoice `json:"choices,omitempty"`
	Options     []DiscordApplicationCommandOption       `json:"options,omitempty"`
}

type DiscordApplicationCommandOptionChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DiscordInteractionComponentCommand struct {
	ID            string               `json:"id"`
	ApplicationID string               `json:"application_id"`
	Type          int                  `json:"type"`
	Data          DiscordComponentData `json:"data"`
	GuildID       string               `json:"guild_id"`
	ChannelID     string               `json:"channel_id"`
	Member        DiscordMember        `json:"member"`
	Token         string               `json:"token"`
	Version       int                  `json:"version"`
}

type DiscordMember struct {
	User         discordgo.User `json:"user"`
	Nick         string         `json:"nick"`
	Roles        []string       `json:"roles"`
	JoinedAt     time.Time      `json:"joined_at"`
	PremiumSince time.Time      `json:"premium_since"`
	Deaf         bool           `json:"deaf"`
	Mute         bool           `json:"mute"`
	Pending      bool           `json:"pending"`
	Permissions  string         `json:"permissions"`
}

type DiscordComponentData struct {
	Type        int                        `json:"type"`
	CustomID    string                     `json:"custom_id,omitempty"`
	Disabled    string                     `json:"disabled,omitempty"`
	Style       int                        `json:"style,omitempty"`
	Label       string                     `json:"label,omitempty"`
	Emoji       DiscordEmoji               `json:"emoji,omitempty"`
	Url         string                     `json:"url,omitempty"`
	Options     []DiscordSelectMenuOptions `json:"options,omitempty"`
	Placeholder string                     `json:"placeholder,omitempty"`
	MinValues   int                        `json:"min_values,omitempty"`
	MaxValues   int                        `json:"max_values,omitempty"`
	Components  []DiscordComponentData     `json:"components,omitempty"`
}

type DiscordInteractionResponse struct {
	Type int                            `json:"type"`
	Data DiscordInteractionResponseData `json:"data"`
}

type DiscordInteractionResponseData struct {
	TTS             bool                         `json:"tts,omitempty"`
	Content         string                       `json:"content,omitempty"`
	Embeds          []embed.Embed                `json:"embeds,omitempty"`
	AllowedMentions discordgo.AllowedMentionType `json:"allowed_mentions,omitempty"`
	Flags           int                          `json:"flags,omitempty"`
	Components      []DiscordComponentData       `json:"components,omitempty"`
}

type DiscordSelectMenuOptions struct {
	Label       string       `json:"label"`
	Value       string       `json:"value"`
	Description string       `json:"description,omitempty"`
	Emoji       DiscordEmoji `json:"emoji,omitempty"`
	Default     bool         `json:"default,omitempty"`
}

type DiscordEmoji struct {
	// Component will have name, id, animated
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Roles         []discordgo.Role `json:"roles,omitempty"`
	User          discordgo.User   `json:"user,omitempty"`
	RequireColons bool             `json:"require_colons,omitempty"`
	Managed       bool             `json:"managed,omitempty"`
	Animated      bool             `json:"animated,omitempty"`
	Available     bool             `json:"available,omitempty"`
}

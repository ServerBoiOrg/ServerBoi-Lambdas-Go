package generalutils

func ServerEmbedComponents(running bool) []DiscordComponentData {
	startComponent := DiscordComponentData{
		Type:     2,
		Label:    "Start",
		Style:    1,
		Disabled: running,
		CustomID: "server:start",
		Emoji: DiscordEmoji{
			Name: "🟢",
		},
	}

	stopComponent := DiscordComponentData{
		Type:     2,
		Label:    "Stop",
		Style:    1,
		Disabled: !running,
		CustomID: "server:stop",
		Emoji: DiscordEmoji{
			Name: "🔴",
		},
	}

	rebootComponent := DiscordComponentData{
		Type:     2,
		Label:    "Reboot",
		Style:    1,
		Disabled: !running,
		CustomID: "server:reboot",
		Emoji: DiscordEmoji{
			Name: "🔁",
		},
	}

	componentData := DiscordComponentData{
		Type: 1,
		Components: []DiscordComponentData{
			startComponent, stopComponent, rebootComponent,
		},
	}

	return []DiscordComponentData{componentData}
}

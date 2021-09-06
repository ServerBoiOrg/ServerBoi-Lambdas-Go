package generalutils

import (
	"fmt"
)

func ServerEmbedComponents(serverID string, running bool) []DiscordComponentData {
	startComponent := DiscordComponentData{
		Type:     2,
		Label:    "Start",
		Style:    1,
		Disabled: running,
		CustomID: fmt.Sprintf("%v.start", serverID),
		Emoji: DiscordEmoji{
			Name: "🟢",
		},
	}

	stopComponent := DiscordComponentData{
		Type:     2,
		Label:    "Stop",
		Style:    1,
		Disabled: !running,
		CustomID: fmt.Sprintf("%v.stop", serverID),
		Emoji: DiscordEmoji{
			Name: "🔴",
		},
	}

	rebootComponent := DiscordComponentData{
		Type:     2,
		Label:    "Reboot",
		Style:    1,
		Disabled: !running,
		CustomID: fmt.Sprintf("%v.reboot", serverID),
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

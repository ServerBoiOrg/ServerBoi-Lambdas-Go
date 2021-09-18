package responseutils

import dt "github.com/awlsring/discordtypes"

func ServerEmbedComponents(running bool) []*dt.Component {
	startComponent := &dt.Component{
		Type:     2,
		Label:    "Start",
		Style:    1,
		Disabled: running,
		CustomID: "server:start",
		Emoji: &dt.Emoji{
			Name: "🟢",
		},
	}

	stopComponent := &dt.Component{
		Type:     2,
		Label:    "Stop",
		Style:    1,
		Disabled: !running,
		CustomID: "server:stop",
		Emoji: &dt.Emoji{
			Name: "🔴",
		},
	}

	rebootComponent := &dt.Component{
		Type:     2,
		Label:    "Reboot",
		Style:    1,
		Disabled: !running,
		CustomID: "server:reboot",
		Emoji: &dt.Emoji{
			Name: "🔁",
		},
	}

	componentData := &dt.Component{
		Type: 1,
		Components: []*dt.Component{
			startComponent, stopComponent, rebootComponent,
		},
	}

	return []*dt.Component{componentData}
}

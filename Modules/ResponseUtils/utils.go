package responseutils

import (
	"errors"
	"fmt"
	"log"
	"time"

	dt "github.com/awlsring/discordtypes"
	"github.com/rumblefrog/go-a2s"
)

func MakeTimestamp() string {
	t := time.Now().UTC()
	return fmt.Sprintf("⏱️ Last updated: %02d:%02d:%02d UTC", t.Hour(), t.Minute(), t.Second())
}

func CallServer(ip string, port int) (a2s *a2s.ServerInfo, err error) {
	for i := 0; ; i++ {
		a2s, err = A2SQuery(ip, (port + i))
		if err == nil {
			return a2s, nil
		}
		if i == 3 {
			return a2s, err
		}
	}
}

func A2SQuery(ip string, port int) (info *a2s.ServerInfo, err error) {
	clientString := fmt.Sprintf("%v:%v", ip, port)
	client, err := a2s.NewClient(clientString)
	if err != nil {
	}
	defer client.Close()
	info, err = client.QueryInfo()
	if err != nil {
	}
	client.Close()
	return info, err
}

func CreateLinkButton(url string) []*dt.Component {
	button := &dt.Component{
		Type:  2,
		Style: 5,
		Label: "Download SSH Key",
		Url:   url,
	}

	componentData := &dt.Component{
		Type: 1,
		Components: []*dt.Component{
			button,
		},
	}

	log.Printf("Component Data: %v", componentData)
	return []*dt.Component{componentData}
}

func FormFooter(owner string, service string, region string) string {
	t := time.Now().UTC()
	timestamp := fmt.Sprintf("⏱️ Last updated: %02d:%02d:%02d UTC", t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf(
		"Owner: %v | 🌎 Hosted on %v in region %v | %v",
		owner,
		service,
		region,
		timestamp,
	)
}

type GetStatusInput struct {
	Service string
	Status  string
	Running bool
}

func GetStatus(input *GetStatusInput) (state string, emoji string, err error) {
	if input.Running {
		state = "Running"
		emoji = "🟢"
	} else {
		switch input.Service {
		case "aws":
			state, emoji = TranslateAwsState(input.Status)
		case "linode":
			state, emoji = TranslateLinodeState(input.Status)
		default:
			return "", "", errors.New("Unsupported service")
		}
	}
	return state, emoji, nil
}

func TranslateAwsState(status string) (state string, stateEmoji string) {
	switch status {
	case "running":
		state = "Running"
		stateEmoji = "🟢"
	case "pending":
		state = "Starting"
		stateEmoji = "🟡"
	case "shutting-down":
		state = "Shutting down"
		stateEmoji = "🔴"
	case "stopping":
		state = "Shutting down"
		stateEmoji = "🔴"
	case "terminated":
		state = "Terminated"
		stateEmoji = "🔴"
	case "stopped":
		state = "Offline"
		stateEmoji = "🔴"
	}
	return state, stateEmoji
}

func TranslateLinodeState(status string) (state string, stateEmoji string) {
	switch status {
	case "running":
		state = "Running"
		stateEmoji = "🟢"
	case "offline":
		state = "Offline"
		stateEmoji = "🔴"
	case "booting":
		state = "Starting"
		stateEmoji = "🟡"
	case "rebooting":
		state = "Rebooting"
		stateEmoji = "🟡"
	case "shutting_down":
		state = "Shutting down"
		stateEmoji = "🔴"
	case "provisioning":
		state = "Starting"
		stateEmoji = "🟡"
	case "deleting":
		state = "Terminated"
		stateEmoji = "🔴"
	case "stopped":
		state = "Offline"
		stateEmoji = "🔴"
	}
	return state, stateEmoji
}

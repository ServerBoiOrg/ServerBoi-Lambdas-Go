package responseutils

import (
	"fmt"
	"time"

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

func TranslateState(service string, status string) (state string, stateEmoji string, err error) {
	switch service {
	case "aws":
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
	case "linode":
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
	}

	return state, stateEmoji, err
}

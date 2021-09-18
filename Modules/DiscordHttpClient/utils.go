package discordhttpclient

import (
	"time"
)

func StatusCodeHandler(headers DiscordHeaders) bool {
	switch headers.StatusCode {
	case 200:
		return true
	case 429:
		time.Sleep(time.Duration(headers.ResetAfter*1000) * time.Millisecond)
		return false
	default:
		// Do more here
		return true
	}
}

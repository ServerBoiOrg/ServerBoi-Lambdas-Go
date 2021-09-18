package main

import dt "github.com/awlsring/discordtypes"

func pong() (pong *dt.InteractionResponse) {
	pong.Type = 1
	return pong
}

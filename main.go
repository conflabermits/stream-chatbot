package main

import (
	"fmt"
	"stream-chatbot/auth"
	"stream-chatbot/common"
	overlay "stream-chatbot/web"
)

/*
1. Check token validity
    * [Twitch dev documentation - How to validate a token](https://dev.twitch.tv/docs/authentication/validate-tokens/#how-to-validate-a-token)
    * `curl -X GET 'https://id.twitch.tv/oauth2/validate' -H 'Authorization: OAuth <token>'`
2. If token is invalid, start oauth process, return token to synchronized channel
    * [Go By Example - Channel Synchonization](https://gobyexample.com/channel-synchronization)
3. Start chatbot
4. Start web overlay
*/

func main() {
	println("Started program!")
	fmt.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))

	go auth.TwitchAuth()
	go overlay.WebOverlay()
	token := <-auth.TokenChan
	fmt.Printf("Received token from auth module: %s\n", token)
}

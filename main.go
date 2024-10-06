package main

import (
	"fmt"
	"log"
	"stream-chatbot/auth"
	"stream-chatbot/chatbot"
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
	log.Println("Started program!")
	fmt.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))

	tokenChan := make(chan string)
	log.Println("Created token channel in main")
	go auth.TwitchAuth(tokenChan)
	log.Println("Kicked off TwitchAuth goroutine")
	twitchToken := <-tokenChan
	log.Println("Passed token from tokenChan to var")
	//<-tokenChan
	//log.Println("Closed tokenChan")
	fmt.Printf("Received token from auth module: %s\n", twitchToken[len(twitchToken)-5:])
	go overlay.WebOverlay()
	log.Println("Kicked off WebOverlay goroutine")
	chatbot.Chatbot(twitchToken)
}

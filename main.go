package main

import (
	"fmt"
	"stream-chatbot/auth"
	"stream-chatbot/common"
	overlay "stream-chatbot/web"
)

func main() {
	println("Started program!")
	fmt.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))

	//tokenChannel := make(chan string)
	//auth.TwitchAuth(tokenChannel)
	//token <- tokenChannel
	go overlay.WebOverlay()
	auth.TwitchAuth()
}

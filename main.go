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

	go auth.TwitchAuth()
	go overlay.WebOverlay()
	token := <-auth.TokenChan
	fmt.Printf("Received token from auth module: %s\n", token)
}

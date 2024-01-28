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
	auth.TwitchAuth()
	overlay.WebOverlay()
}

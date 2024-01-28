package main

import (
	"fmt"
	"stream-chatbot/common"
	overlay "stream-chatbot/web"
)

func main() {
	println("Started program!")
	fmt.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))
	overlay.WebOverlay()
}

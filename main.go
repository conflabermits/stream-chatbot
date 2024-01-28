package main

import (
	"fmt"
	"stream-chatbot/common"
)

func main() {
	println("Hello, World!")
	fmt.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))
}

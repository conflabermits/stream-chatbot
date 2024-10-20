package main

import (
	"flag"
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

type ChatbotCreds struct {
	ClientID       string
	ClientSecret   string
	TwitchUsername string
	TwitchChannel  string
	BroadcasterID  string
	TwitchToken    string
}

type Options struct {
	CredsFile     string
	CredsEnv      bool
	CredsOverride string
}

func parseArgs() (*Options, error) {
	options := &Options{}

	flag.StringVar(&options.CredsFile, "credsFile", "", "Credentials File")
	flag.BoolVar(&options.CredsEnv, "credsEnv", false, "Use environment variables for credentials")
	flag.StringVar(&options.CredsOverride, "credsOverride", "", "Override specific credentials with provided comma-separated key-value pairs")
	flag.Usage = func() {
		fmt.Printf("Usage: stream-chatbot [options]\n\n")
		flag.PrintDefaults()
		fmt.Println("Note: credsOverride takes presedence over credsEnv, which takes precedence over credsFile")
		// File can have defaults, env vars can override those, and credsOverride provides a final absolute override
	}
	flag.Parse()

	return options, nil
}

func main() {
	log.Println("Started program!")
	fmt.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))

	options, err := parseArgs()
	common.CheckErr(err, "parseArgs")
	if options.CredsFile != "" {
		log.Printf("Using credentials file: %s\n", options.CredsFile)
		//Get credentials from file
	}
	if options.CredsEnv {
		log.Println("Using environment variables for credentials")
		//Get credentials from environment variables
	}
	if options.CredsOverride != "" {
		log.Printf("Overriding credentials with provided values\n")
		//For each comma-separated key-value pair, override the value of the key with the provided value
	}

	// Check token validity
	var twitchToken string
	if common.CheckTwitchToken(common.GetEnvVar("twitchToken")) {
		log.Println("Stored twitchToken is valid")
		twitchToken = common.GetEnvVar("twitchToken")
		log.Printf("Received token from file: %s\n", twitchToken[len(twitchToken)-5:])
	} else {
		log.Println("Stored twitchToken is invalid")
		tokenChan := make(chan string)
		log.Println("Created token channel in main")
		go auth.TwitchAuth(tokenChan)
		log.Println("Kicked off TwitchAuth goroutine")
		twitchToken = <-tokenChan
		log.Println("Passed token from tokenChan to var")
		//<-tokenChan
		//log.Println("Closed tokenChan")
		log.Printf("Received token from auth module: %s\n", twitchToken[len(twitchToken)-5:])
	}

	go overlay.WebOverlay()
	log.Println("Kicked off WebOverlay goroutine")
	chatbot.Chatbot(twitchToken)
}

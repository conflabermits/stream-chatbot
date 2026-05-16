package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"encoding/base64"
	"encoding/json"
	"golang.org/x/oauth2"

	"stream-chatbot/auth"
	"stream-chatbot/chatbot"
	"stream-chatbot/common"
	spotify_client "stream-chatbot/spotify"
	overlay "stream-chatbot/web"
	"strings"
)

var ChatbotVars = []string{
	"ClientID",
	"ClientSecret",
	"TwitchUsername",
	"TwitchChannel",
	"BroadcasterID",
	"TwitchToken",
	"SpotifyClientID",
	"SpotifyClientSecret",
	"SpotifyToken",
}

type Options struct {
	CredsFile     string
	CredsEnv      bool
	CredsOverride string
}

func parseArgs() (*Options, error) {
	options := &Options{}

	flag.StringVar(&options.CredsFile, "credsFile", ".creds", "Credentials File")
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

func assignVar(line string) {
	keyval := strings.Split(line, "=")
	key := strings.TrimSpace(keyval[0])
	value := strings.TrimSpace(keyval[1])
	for _, varname := range ChatbotVars {
		if key == varname {
			switch key {
			case "ClientID":
				common.ChatbotCreds["ClientID"] = value
			case "ClientSecret":
				common.ChatbotCreds["ClientSecret"] = value
			case "TwitchUsername":
				common.ChatbotCreds["TwitchUsername"] = value
			case "TwitchChannel":
				common.ChatbotCreds["TwitchChannel"] = value
			case "BroadcasterID":
				common.ChatbotCreds["BroadcasterID"] = value
			case "TwitchToken":
				common.ChatbotCreds["TwitchToken"] = value
			case "SpotifyClientID":
				common.ChatbotCreds["SpotifyClientID"] = value
			case "SpotifyClientSecret":
				common.ChatbotCreds["SpotifyClientSecret"] = value
			case "SpotifyToken":
				common.ChatbotCreds["SpotifyToken"] = value
			}
		}
	}

}

func getChatbotCredsFromFile(filename string) {
	file, err := os.Open(filename)
	common.CheckErr(err, "getChatbotCredsFromFile - Error opening file")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		assignVar(line)
	}
}

func getChatbotCredsFromEnv() {
	for _, key := range ChatbotVars {
		value := common.GetEnvVar(key)
		switch key {
		case "ClientID":
			common.ChatbotCreds["ClientID"] = value
		case "ClientSecret":
			common.ChatbotCreds["ClientSecret"] = value
		case "TwitchUsername":
			common.ChatbotCreds["TwitchUsername"] = value
		case "TwitchChannel":
			common.ChatbotCreds["TwitchChannel"] = value
		case "BroadcasterID":
			common.ChatbotCreds["BroadcasterID"] = value
		case "TwitchToken":
			common.ChatbotCreds["TwitchToken"] = value
		case "SpotifyClientID":
			common.ChatbotCreds["SpotifyClientID"] = value
		case "SpotifyClientSecret":
			common.ChatbotCreds["SpotifyClientSecret"] = value
		case "SpotifyToken":
			common.ChatbotCreds["SpotifyToken"] = value
		}
	}
}

func main() {
	// Ensure the logs directory exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		common.CheckErr(err, "main - Error creating logs directory")
	}
	// Create a log file with the current date and time (GH Copilot)
	logFileName := "logs/" + time.Now().Format("2006-01-02_15-04-05") + ".log"
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	common.CheckErr(err, "main - Error opening log file")
	defer logFile.Close()
	// Set log output to the file
	log.SetOutput(logFile)

	log.Println("Started program!")
	log.Printf("GOPATH is set to: %s\n", common.GetEnvVar("GOPATH"))

	options, err := parseArgs()
	common.CheckErr(err, "parseArgs")

	if options.CredsFile != "" {
		log.Printf("Using credentials file: %s\n", options.CredsFile)
		getChatbotCredsFromFile(options.CredsFile)
	}
	if options.CredsEnv {
		log.Println("Using environment variables for credentials")
		getChatbotCredsFromEnv()
	}
	if options.CredsOverride != "" {
		//For each comma-separated key-value pair, override the value of the key with the provided value
		log.Printf("Overriding credentials with provided values\n")
	}

	twitchToken := common.ChatbotCreds["TwitchToken"]
	if len(twitchToken) > 5 {
		log.Printf("Starting with token: %s\n", twitchToken[len(twitchToken)-5:])
	} else {
		log.Println("Starting without token")
	}

	// Check token validity
	if common.CheckTwitchToken(twitchToken) && len(twitchToken) > 5 {
		log.Println("Stored twitchToken is valid")
	} else {
		log.Println("Stored twitchToken is invalid")
		tokenChan := make(chan string)
		log.Println("Created token channel in main")
		go auth.TwitchAuth(tokenChan, common.ChatbotCreds)
		log.Println("Kicked off TwitchAuth goroutine")
		twitchToken = <-tokenChan
		log.Println("Passed token from tokenChan to var")
		//<-tokenChan
		//log.Println("Closed tokenChan")
		log.Printf("Received token from auth module: %s\n", twitchToken[len(twitchToken)-5:])
		err := common.WriteNewValueToProperties(options.CredsFile, "TwitchToken", twitchToken)
		common.CheckErr(err, "main - Error writing new token to properties file")
	}

	// Spotify Auth and Initialization
	spotifyTokenStr := common.ChatbotCreds["SpotifyToken"]
	var spotifyTok *oauth2.Token

	if len(spotifyTokenStr) > 10 {
		// Attempt to decode base64 and unmarshal
		decodedBytes, err := base64.StdEncoding.DecodeString(spotifyTokenStr)
		if err == nil {
			err = json.Unmarshal(decodedBytes, &spotifyTok)
			if err != nil {
				log.Println("Error unmarshaling Spotify token:", err)
				spotifyTok = nil
			} else {
				log.Println("Loaded Spotify token from creds successfully.")
			}
		} else {
			log.Println("Error decoding Spotify token:", err)
		}
	}

	if spotifyTok == nil || spotifyTok.AccessToken == "" {
		log.Println("Stored Spotify token is invalid or missing.")
		spotifyChan := make(chan *oauth2.Token)
		go auth.SpotifyAuth(spotifyChan, common.ChatbotCreds["SpotifyClientID"], common.ChatbotCreds["SpotifyClientSecret"])
		log.Println("Kicked off SpotifyAuth goroutine")
		spotifyTok = <-spotifyChan
		log.Println("Received token from SpotifyAuth module")
		
		// Marshal and encode
		b, err := json.Marshal(spotifyTok)
		if err == nil {
			encodedStr := base64.StdEncoding.EncodeToString(b)
			err = common.WriteNewValueToProperties(options.CredsFile, "SpotifyToken", encodedStr)
			common.CheckErr(err, "main - Error writing new Spotify token to properties file")
		} else {
			log.Println("Error marshaling Spotify token:", err)
		}
	}

	spotify_client.Initialize(common.ChatbotCreds["SpotifyClientID"], common.ChatbotCreds["SpotifyClientSecret"], spotifyTok)

	go overlay.WebOverlay()
	log.Println("Kicked off WebOverlay goroutine")
	go overlay.DonorboxOverlay()
	log.Println("Kicked off DonorboxOverlay goroutine")

	// Local Console Input
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			// Simulate broadcaster input
			badges := map[string]int{"broadcaster": 1}
			chatbot.ProcessCommand(text, common.ChatbotCreds["TwitchUsername"], badges, nil)
		}
	}()

	chatbot.Chatbot(twitchToken)
}

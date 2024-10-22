package common

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var ChatbotCreds map[string]string = map[string]string{
	"ClientID":       "",
	"ClientSecret":   "",
	"TwitchUsername": "",
	"TwitchChannel":  "",
	"BroadcasterID":  "",
	"TwitchToken":    "",
}

func CheckErr(err error, from string) {
	if err != nil {
		log.Printf("Error: %v\n", err)
		log.Printf("Error from: %s\n", from)
		panic(err)
	}
}

func GetEnvVar(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Error: Environment variable %s does not exist or is not set\n", key)
		panic(fmt.Sprintf("Environment variable not set: %s\n", key))
	}
	return value
}

func CheckTwitchToken(token string) bool {
	// Check token validity
	// Twitch dev documentation - How to validate a token:
	// https://dev.twitch.tv/docs/authentication/validate-tokens/#how-to-validate-a-token

	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	CheckErr(err, "CheckTwitchToken - Error creating request")

	req.Header.Set("Authorization", "OAuth "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	CheckErr(err, "CheckTwitchToken - Error sending request")
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		log.Println("Token is valid")
		return true
	case 401:
		log.Println("Token is invalid")
		return false
	default:
		log.Printf("Unexpected response code: %d\n", resp.StatusCode)
		return false
	}
}

/* func WriteToFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckErr(err, "WriteToFile - Error opening file")
	defer file.Close()

	_, err = file.WriteString("data")
	CheckErr(err, "WriteToFile - Error writing to file")
}

func ReadFromFile(filename string) string {
	file, err := os.Open(filename)
	CheckErr(err, "ReadFromFile - Error opening file")
	defer file.Close()

	data := make([]byte, 1024)
	count, err := file.Read(data)
	CheckErr(err, "ReadFromFile - Error reading from file")
	return string(data[:count])
} */

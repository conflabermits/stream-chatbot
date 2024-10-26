package common

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var ChatbotCreds map[string]string = map[string]string{
	"ClientID":       "",
	"ClientSecret":   "",
	"TwitchUsername": "",
	"TwitchChannel":  "",
	"BroadcasterID":  "",
	"TwitchToken":    "",
}

func WriteNewValueToProperties(filename, key, value string) error {
	// Open the properties file for reading and writing
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Create a temporary buffer to store the modified lines
	var lines []string

	// Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line contains the target key
		if strings.HasPrefix(line, key+"=") {
			// Replace the value with the new one
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		} else {
			// Append the original line to the buffer
			lines = append(lines, line)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return err
	}

	// Truncate the file to overwrite its contents
	if err := file.Truncate(0); err != nil {
		return err
	}

	// Seek to the beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Write the modified lines back to the file
	for _, line := range lines {
		_, err = fmt.Fprintln(file, line)
		if err != nil {
			return err
		}
	}

	return nil
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

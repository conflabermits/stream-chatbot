package common

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func GetEnvVar(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		fmt.Printf("Error: Environment variable %s does not exist or is not set\n", key)
		os.Exit(1)
	}
	return value
}

func CheckTwitchToken(token string) bool {
	// Check token validity
	// Twitch dev documentation - How to validate a token:
	// https://dev.twitch.tv/docs/authentication/validate-tokens/#how-to-validate-a-token

	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return false
	}

	req.Header.Set("Authorization", "OAuth "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return false
	}
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
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if _, err := file.WriteString(data); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}
}

func ReadFromFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	data := make([]byte, 1024)
	count, err := file.Read(data)
	if err != nil {
		fmt.Printf("Error reading from file: %v\n", err)
		os.Exit(1)
	}
	return string(data[:count])
} */

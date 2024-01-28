package common

import (
	"fmt"
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

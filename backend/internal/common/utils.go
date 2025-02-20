package common

import (
	"log"
	"os"
)

func GetEnvironmentVariable(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Unable to load environment variable '%s'", key)
	}

	return val
}

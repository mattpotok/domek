package common

import (
	"fmt"
	"os"
	"path/filepath"
)

const SERVICE_NAME = "domek"

func GetServiceDirPath() string {
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Unable to get home directory: %s\n", err)
		os.Exit(1)
	}

	return filepath.Join(homeDirPath, ".local", "share", SERVICE_NAME)
}

func CreateServiceDirectory() {
	serviceDirPath := GetServiceDirPath()
	err := os.MkdirAll(serviceDirPath, 0o755)
	if err != nil {
		fmt.Printf("Unable to create service directory: %s\n", err)
		os.Exit(1)
	}
}

func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	var zero T
	for _, element := range slice {
		if predicate(element) {
			return element, true
		}
	}

	return zero, false
}

package api

import (
	"fmt"
	"os"
	"strings"
)

const (
	BaseURL string = "https://api.telegram.org"
)

func MakeEndpointStr(resource, key string) string {
	return fmt.Sprintf("%s/bot%s/%s", BaseURL, key, resource)
}

func GetAPIKey(path string) (string, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(key)), nil
}

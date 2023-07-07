package api

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	BaseURL string = "https://api.telegram.org"
)

func MakeEndpointStr(resource, key string) string {
	return fmt.Sprintf("%s/bot%s/%s", BaseURL, key, resource)
}

func GetAPIKey(path string) string {
	key, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	return strings.TrimSpace(string(key))
}

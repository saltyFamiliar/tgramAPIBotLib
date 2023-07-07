package bot

import (
	"encoding/json"
	"fmt"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"io"
	"log"
	"net/http"
	"time"
)

type TgramBot struct {
	Offset int
	key    string
	client *http.Client
}

func NewTgramBot(apiKey string) *TgramBot {
	return &TgramBot{
		Offset: 0,
		key:    apiKey,
		client: &http.Client{},
	}
}

func (bot *TgramBot) APIRequest(resource string) *api.Response {
	reqUrl := api.MakeEndpointStr(resource, bot.key)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		log.Fatalln("Unable to create request")
	}

	response, err := bot.client.Do(req)
	if err != nil {
		log.Fatalln("Unable to get response")
	}

	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatalln("Unable to read response body")
	}

	respBody := &api.Response{}
	if err := json.Unmarshal(body, &respBody); err != nil {
		log.Fatalln("Unable to marshal response body")
	}

	return respBody
}

func (bot *TgramBot) GetMe() *api.User {
	result, err := bot.APIRequest("getMe").Unwrap()
	if err != nil {
		log.Fatalln(err)
	}

	user := &api.User{}
	if err := json.Unmarshal(result, user); err != nil {
		log.Fatalln("Unable to unmarshal response body")
	}

	return user
}

func (bot *TgramBot) SendMsg(msg string) {
	chatID := "405907484"
	req := fmt.Sprintf("sendMessage?chat_id=%s&text=%s", chatID, msg)
	bot.APIRequest(req)
}

func (bot *TgramBot) Spam() {
	for {
		time.Sleep(3 * time.Second)
		bot.SendMsg("Hello world")
	}
}

func (bot *TgramBot) GetUpdates() []api.Update {
	result, err := bot.APIRequest(fmt.Sprintf("getUpdates?offset=%d", bot.Offset)).Unwrap()
	if err != nil {
		log.Fatalln(err)
	}

	var updates []api.Update
	if err := json.Unmarshal(result, &updates); err != nil {
		log.Fatalln("Unable to marshal result into []Update")
	}

	return updates
}

package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"log"
	"net/http"
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

func (bot *TgramBot) APIRequest(ctx context.Context, resource string) (*api.Response, error) {
	reqUrl := api.MakeEndpointStr(resource, bot.key)
	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	response, err := bot.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v,", err)
		}
	}()

	respBody := &api.Response{}
	if err := json.NewDecoder(response.Body).Decode(respBody); err != nil {
		return nil, err
	}

	return respBody, nil
}

func (bot *TgramBot) GetMe(ctx context.Context) (*api.User, error) {
	resp, err := bot.APIRequest(ctx, "getMe")
	if err != nil {
		return nil, err
	}

	result, err := resp.Unwrap()
	if err != nil {
		return nil, err
	}

	user := &api.User{}
	if err := json.Unmarshal(result, user); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body: %v", err)
	}

	return user, nil
}

func (bot *TgramBot) SendMsg(ctx context.Context, msg string, chatID int64) error {
	req := fmt.Sprintf("sendMessage?chat_id=%d&text=%s", chatID, msg)
	_, err := bot.APIRequest(ctx, req)
	return err
}

func (bot *TgramBot) GetUpdates(ctx context.Context) ([]api.Update, error) {
	resp, err := bot.APIRequest(ctx, fmt.Sprintf("getUpdates?offset=%d", bot.Offset))
	if err != nil {
		return nil, err
	}

	result, err := resp.Unwrap()
	if err != nil {
		return nil, err
	}

	var updates []api.Update
	if err := json.Unmarshal(result, &updates); err != nil {
		return nil, fmt.Errorf("unable to marshal result into []Update: %w", err)
	}

	return updates, nil
}

package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"log"
	"net/http"
	"strings"
	"time"
)

type RoutineRegistry map[string]*Routine

type TgramBot struct {
	Offset   int
	key      string
	Registry RoutineRegistry
	client   *http.Client
}

func NewTgramBot(apiKey string) *TgramBot {
	return &TgramBot{
		Offset:   0,
		key:      apiKey,
		Registry: RoutineRegistry{},
		client:   &http.Client{},
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

func (bot *TgramBot) RegisterRoutine(hook string, routine *Routine) error {
	if _, hookTaken := bot.Registry[hook]; !hookTaken {
		bot.Registry[hook] = routine
		return nil
	}
	return fmt.Errorf("couldn't register routine: name taken")
}

func (bot *TgramBot) ParseMessage(msg string) (*Routine, []string, error) {
	words := strings.Split(msg, " ")
	routine, ok := bot.Registry[words[0]]
	if !ok {
		return nil, nil, fmt.Errorf("routine not found")
	}

	var args []string
	if len(words) > 1 {
		args = words[1:]
	}

	return routine, args, nil
}

func (bot *TgramBot) Run() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		updates, err := bot.GetUpdates(ctx)
		cancel()
		if err != nil {
			log.Printf("Error getting updates: %v", err)
		}

		for _, update := range updates {
			bot.Offset = int(update.UpdateId) + 1

			routine, args, err := bot.ParseMessage(update.Message.Text)
			if err != nil {
				fmt.Println(err)
				continue
			}

			respMsg, err := routine.Execute([]string{strings.Join(args, " ")})
			if err != nil {
				fmt.Println(err)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err = bot.SendMsg(ctx, respMsg, update.Message.Chat.Id); err != nil {
				log.Printf("unable to send message: %v ", err)
			}
			cancel()
		}

		time.Sleep(2 * time.Second)
	}
}

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

// RoutineRegistry is a map from hook strings to Routine pointers.
// It is used to store the bot's registered routines.
type RoutineRegistry map[string]*Routine

// TgramBot is the main Telegram bot struct.
// It contains the current update offset, API key,
// registry mapping of hook strings to Routines,
// and an HTTP client for making API requests.
type TgramBot struct {
	Offset   int
	key      string
	Registry RoutineRegistry
	client   *http.Client
}

// NewTgramBot constructs a new TgramBot instance.
// It initializes the offset to 0, API key to the provided key,
// and empty Registry and HTTP client.
func NewTgramBot(apiKey string) *TgramBot {
	return &TgramBot{
		Offset:   0,
		key:      apiKey,
		Registry: RoutineRegistry{},
		client:   &http.Client{},
	}
}

// APIRequest makes a request to the Telegram Bot API.
// It accepts a context.Context and API resource endpoint as arguments.
// It returns a api.Response struct and error.
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

// GetMe retrieves basic information about the bot from the Telegram API.
// It is used mostly for testing purposes to validate the bot's API key.
// It accepts a context.Context as an argument.
// It returns an api.User struct containing info about the bot, and an error.
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

// SendMsg sends a text message to a chat via the Telegram Bot API.
// It accepts a context.Context, message text, and target chat ID.
// It returns any error from the API request.
func (bot *TgramBot) SendMsg(ctx context.Context, msg string, chatID int64) error {
	req := fmt.Sprintf("sendMessage?chat_id=%d&text=%s", chatID, msg)
	_, err := bot.APIRequest(ctx, req)
	return err
}

// SendMsgWithTimeout sends a text message to a chat with a timeout.
// It accepts the message text, target chat ID, and timeout duration.
// It uses a temporary context.Context with the timeout.
// It returns any error from the API request.
func (bot *TgramBot) SendMsgWithTimeout(msg string, chatID int64, timeout time.Duration) error {
	req := fmt.Sprintf("sendMessage?chat_id=%d&text=%s", chatID, msg)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := bot.APIRequest(ctx, req)
	return err
}

// GetUpdates retrieves new update objects from the Telegram Bot API.
// It requests updates with an offset greater than the current offset stored in the TgramBot struct.
// It accepts a context.Context as an argument.
// It returns a slice of api.Update structs representing new updates, and an error.
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

// RegisterRoutine registers a Routine struct to handle a specific hook.
// It accepts the hook name as a string, and pointer to the Routine struct.
// The TgramBot's Registry map is updated to map the hook to the routine.
// It returns an error if the hook name is already taken.
func (bot *TgramBot) RegisterRoutine(hook string, routine *Routine) error {
	if _, hookTaken := bot.Registry[hook]; !hookTaken {
		bot.Registry[hook] = routine
		return nil
	}
	return fmt.Errorf("couldn't register routine: name taken")
}

// ParseMessage parses a chat message to extract the routine hook and arguments.
// It splits the message into words. The first word is assumed to be the routine hook.
// It returns the corresponding Routine struct from the bot's Registry map.
// Any subsequent words are returned as a string slice of arguments.
// It returns an error if no routine is registered for the given hook.
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

// Run starts the main loop for fetching updates and handling requests.
// It creates two channels for updates and jobs.
// A goroutine fetches updates from the API every few seconds
// and sends them to the updates channel.
// Another goroutine listens to the updates channel,
// processes each update into a job,
// updates the bot's offset,
// and sends the job to the jobs channel.
// For each job, a goroutine parses the message,
// executes the matching routine,
// and sends the routine's response back to the user.
// This loop continues indefinitely to continuously
// process updates and handle requests.
func (bot *TgramBot) Run() {
	updatesCh := make(chan []api.Update, 10)
	jobCh := make(chan *api.Message, 10)

	// update producer
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			updates, err := bot.GetUpdates(ctx)
			if err != nil {
				log.Printf("Error getting updates: %v", err)
			}
			cancel()
			updatesCh <- updates
			time.Sleep(4 * time.Second)
		}
	}()

	// consumes updates, produces jobs
	go func() {
		for updates := range updatesCh {
			for _, update := range updates {
				bot.Offset = int(update.UpdateId) + 1
				if update.Message == nil {
					continue
				}

				jobCh <- update.Message
				go func(msg *api.Message) {
					ackMsg := fmt.Sprintf("Received request: %s", msg.Text)
					if err := bot.SendMsgWithTimeout(ackMsg, msg.Chat.Id, 5*time.Second); err != nil {
						fmt.Println(err)
					}
				}(update.Message)
			}
		}
	}()

	// consumes jobs, sends output to user
	for job := range jobCh {
		go func(reqMsg *api.Message) {
			routine, args, err := bot.ParseMessage(reqMsg.Text)
			if err != nil {
				if msgErr := bot.SendMsgWithTimeout(err.Error(), reqMsg.Chat.Id, 5*time.Second); msgErr != nil {
					fmt.Println(msgErr)
				}
				return
			}

			respMsg, err := routine.Execute(args)
			if err != nil {
				if msgErr := bot.SendMsgWithTimeout(err.Error(), reqMsg.Chat.Id, 5*time.Second); msgErr != nil {
					fmt.Println(msgErr)
				}
				return
			}

			if err = bot.SendMsgWithTimeout(respMsg, reqMsg.Chat.Id, 5*time.Second); err != nil {
				fmt.Printf("unable to send message: %v ", err)
			}
		}(job)
	}
}

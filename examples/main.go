package main

import (
	"context"
	"fmt"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"github.com/saltyFamiliar/tgramAPIBotLib/pkg/bot"
	"log"
	"strings"
	"time"
)

func echo(msg string) (string, error) {
	return msg, nil
}

func main() {
	tGramBot := bot.NewTgramBot(api.GetAPIKey("token.txt"))
	echoRoutine := bot.NewRoutine(bot.Action{
		Raw: echo,
		Wrapper: func(i ...interface{}) (string, error) {
			return echo(i[0].(string))
		},
	})

	if err := tGramBot.RegisterRoutine("echo", echoRoutine); err != nil {
		fmt.Println("hook previously registered")
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		updates, err := tGramBot.GetUpdates(ctx)
		cancel()
		if err != nil {
			log.Printf("Error getting updates: %v", err)
		}

		for _, update := range updates {
			reqMsg := strings.Split(update.Message.Text, " ")
			routine := tGramBot.Registry[reqMsg[0]]
			args := []string{}
			if len(reqMsg) > 1 {
				args = reqMsg[1:]
			}

			respMsg, err := routine.Execute([]string{strings.Join(args, " ")})
			if err != nil {
				fmt.Println(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err = tGramBot.SendMsg(ctx, respMsg, update.Message.Chat.Id); err != nil {
				log.Printf("unable to send message: %v ", err)
			}
			cancel()

			tGramBot.Offset = int(update.UpdateId) + 1
		}

		time.Sleep(2 * time.Second)
	}
}

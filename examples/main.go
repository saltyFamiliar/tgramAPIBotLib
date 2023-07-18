package main

import (
	"fmt"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"github.com/saltyFamiliar/tgramAPIBotLib/pkg/bot"
	"log"
)

func echo(msg string) (string, error) {
	return msg, nil
}

func main() {
	apiKey, err := api.GetAPIKey("token.txt")
	if err != nil {
		log.Fatalln("Api key file not found")
	}
	tGramBot := bot.NewTgramBot(apiKey)

	echoRoutine := bot.NewRoutine(bot.Action{
		Raw: echo,
		Wrapper: func(i ...interface{}) (string, error) {
			return echo(i[0].(string))
		},
	})

	if err := tGramBot.RegisterRoutine("echo", echoRoutine); err != nil {
		fmt.Println("hook previously registered")
	}

	tGramBot.Run()
}

package main

import (
	"fmt"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"github.com/saltyFamiliar/tgramAPIBotLib/pkg/bot"
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

	tGramBot.Run()
}

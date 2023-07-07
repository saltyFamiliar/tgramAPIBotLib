package main

import (
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"time"

	"github.com/saltyFamiliar/tgramAPIBotLib/bot"
)

func main() {
	tGramBot := bot.NewTgramBot(api.GetAPIKey("token.txt"))
	for {
		for _, update := range tGramBot.GetUpdates() {
			tGramBot.SendMsg(update.Message.Text)
			tGramBot.Offset = int(update.UpdateId) + 1
		}

		time.Sleep(2 * time.Second)
	}
}

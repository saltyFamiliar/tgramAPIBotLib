package main

import (
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"log"
	"time"

	"github.com/saltyFamiliar/tgramAPIBotLib/bot"
)

func main() {
	tGramBot := bot.NewTgramBot(api.GetAPIKey("token.txt"))
	for {
		updates, err := tGramBot.GetUpdates()
		if err != nil {
			log.Printf("Error getting updates: %v", err)
		}

		for _, update := range updates {
			if err = tGramBot.SendMsg(update.Message.Text, update.Message.Chat.Id); err != nil {
				log.Printf("unable to send message: %v ", err)
			}
			tGramBot.Offset = int(update.UpdateId) + 1
		}

		time.Sleep(2 * time.Second)
	}
}

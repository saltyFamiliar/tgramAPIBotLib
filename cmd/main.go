package main

import (
	"context"
	"github.com/saltyFamiliar/tgramAPIBotLib/api"
	"github.com/saltyFamiliar/tgramAPIBotLib/pkg/bot"
	"log"
	"time"
)

func main() {
	tGramBot := bot.NewTgramBot(api.GetAPIKey("token.txt"))
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		updates, err := tGramBot.GetUpdates(ctx)
		cancel()
		if err != nil {
			log.Printf("Error getting updates: %v", err)
		}

		for _, update := range updates {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err = tGramBot.SendMsg(ctx, update.Message.Text, update.Message.Chat.Id); err != nil {
				log.Printf("unable to send message: %v ", err)
			}
			cancel()
			tGramBot.Offset = int(update.UpdateId) + 1
		}

		time.Sleep(2 * time.Second)
	}
}

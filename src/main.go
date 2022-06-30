package main

import (
	"log"
	"os"
	"time"
	"github.com/baka-lavr/goinchnails/src/database"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


type UpdateData struct {
	User int64
	Chat int64
	Message *tgbotapi.Message
	Callback *tgbotapi.CallbackQuery
}


func main() {
	db := db.InitDB()
	sm := InitMachine(db)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	//bot.Debug = true
	u_config := tgbotapi.NewUpdate(0)
	u_config.Timeout = 10
	updates := bot.GetUpdatesChan(u_config)
	time.Sleep(time.Second)
	updates.Clear()

	for update := range updates {
		//user := update.Message.From.ID
		data := UpdateData{}
		if update.Message != nil {
			data.Message = update.Message
			data.User = update.Message.From.ID
			data.Chat = update.Message.Chat.ID
		}
		if update.CallbackQuery != nil {
			data.Callback = update.CallbackQuery
			data.User = update.CallbackQuery.From.ID
			data.Chat = update.CallbackQuery.Message.Chat.ID
		}

		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,"Error")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
		
		msg := sm.Process(data)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
		if update.CallbackQuery != nil {
			call := tgbotapi.NewCallback(data.Callback.ID, data.Callback.Data)
			if _, err := bot.Request(call); err != nil {
				panic(err)
			}
		}
	}
}
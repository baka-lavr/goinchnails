package main

import (
	//"strconv"
	"log"
	"os"
	"os/signal"
	"context"
	"time"
	"github.com/baka-lavr/goinchnails/src/database"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateData struct {
	User int64
	Chat int64
	Message *tgbotapi.Message
	Callback *tgbotapi.CallbackQuery
	Command bool
}


type Application struct {
	bot *tgbotapi.BotAPI
	sm StateMachine
	db db.DataBase
	updates tgbotapi.UpdatesChannel
	notice chan db.Notifier
}

func InitApplication() Application {
	database := db.InitDB()
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

	notice := make(chan db.Notifier)
	sm := InitMachine(database, notice)

	app := Application{bot,sm,database,updates,notice}

	go app.UpdateHandle()
	go app.NoticeHandler()
	return app
}

func (app *Application) UpdateHandle() {
	for update := range app.updates {
		//user := update.Message.From.ID
		data := UpdateData{}
		if update.Message != nil {
			data.Message = update.Message
			data.User = update.Message.From.ID
			data.Chat = update.Message.Chat.ID
			data.Command = update.Message.IsCommand()
		}
		if update.CallbackQuery != nil {
			data.Callback = update.CallbackQuery
			data.User = update.CallbackQuery.From.ID
			data.Chat = update.CallbackQuery.Message.Chat.ID
		}
		
		msg := app.sm.Process(data)
		var res tgbotapi.Message
		var err error


		if !app.sm.edit {
			res, err = app.bot.Send(msg)
		} else {
			var edit tgbotapi.EditMessageTextConfig
			keys,ok := msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup)
			if ok {
				edit = tgbotapi.NewEditMessageTextAndMarkup(data.User,app.db.GetLastMessage(app.sm.user),msg.Text,*keys)
			} else {
				edit = tgbotapi.NewEditMessageText(data.User,app.db.GetLastMessage(app.sm.user),msg.Text)
			}
			res, err = app.bot.Send(edit)
		}

		log.Println(res.MessageID)
		app.db.SetLastMessage(data.User, res.MessageID)
		if err != nil {
			log.Println(err)
		}
		if update.CallbackQuery != nil {
			call := tgbotapi.NewCallback(data.Callback.ID, data.Callback.Data)
			if _, err := app.bot.Request(call); err != nil {
				panic(err)
			}
		}
	}
}

func (app *Application) NoticeHandler() {
	for message := range app.notice {
		msg := tgbotapi.NewMessage(message.User, message.Text)
		if _, err := app.bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func (app *Application) Close() {
	//close(app.updates)
	close(app.notice)
}

func main() {
	app := InitApplication()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<- sig
	_, cancel := context.WithTimeout(context.Background(),time.Second*10)
	defer cancel()
	app.Close()
	log.Println("Server stopped")
}
package main

import (
	"log"
	"errors"
	"reflect"
	"strings"
	"strconv"
	"github.com/baka-lavr/goinchnails/src/database"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateMachine struct {
	db db.DataBase
	notice chan db.Notifier
	states map[string]State
	state string
	user string
	update UpdateData
	//msg tgbotapi.MessageConfig
	edit bool
}

func InitMachine(db db.DataBase, notice chan db.Notifier) (StateMachine) {
	sm := StateMachine{
		db: db,
		notice: notice,
		states: map[string]State{
			"Start": Start{},
			"Client_Start": Client_Start{},
			"Client_Type": Client_Type{},
			"Client_Master": Client_Master{},
			"Client_Day": Client_Day{},
			"Client_Time": Client_Time{},
			"Client_Phone": Client_Phone{},
			"Client_Confirm": Client_Confirm{},
			"Client_Check": Client_Check{},
			"Client_Delete": Client_Delete{},
			"Master_Start": Master_Start{},
			"Master_Type": Master_Type{},
			"Master_Days": Master_Days{},
			"Master_Time": Master_Time{},
			"Master_Time_End": Master_Time_End{},
			"Master_Name": Master_Name{},
			"Master_Forename": Master_Forename{},
			"Master_Phone": Master_Phone{},
			"Master_Address": Master_Address{},
			"Master_Confirm": Master_Confirm{},
			"Master_Check": Master_Check{},
			"Master_Delete": Master_Delete{},
			"Master_Stop": Master_Stop{},
		},
	}
	return sm
}


func (sm *StateMachine) ChangeState(state string) {
	sm.state = state
	sm.db.SetState(sm.user, state)
	log.Println(sm.state)
}

func (sm *StateMachine) ExecAction() (string, error) {
	var text string
	if res,ok := sm.states[sm.state].(Readable); ok {
		if sm.update.Message == nil {
			text := "Введите сообщение повторно.\n"
			return text, errors.New("Message not found")
		}
		arg := sm.update.Message.Text
		res.Read(sm, arg)
		return "", nil
	}
	if res,ok := sm.states[sm.state].(Contactable); ok {
		res.Contact(sm, sm.update.Message.Contact.PhoneNumber)
		return "", nil
	}
	if sm.update.Message != nil {
		if !sm.update.Message.From.IsBot {
			return "", nil
		}
	}
	if sm.update.Callback == nil {
		text := "Введите сообщение повторно.\n"
		return text, errors.New("Message not found")
	}
	data := sm.update.Callback.Data
	var name string
	var arg string
	if inx := strings.Index(data, ":"); inx != -1 {
		name = data[:inx]
		arg = data[inx+1:]
	} else {
		name = data
	}
	if method := reflect.ValueOf(sm.states[sm.state]).MethodByName(name); method.IsValid() {
		pass := []reflect.Value{reflect.ValueOf(sm)}
		if arg != "" {
			pass = append(pass, reflect.ValueOf(arg))
		}
		res := method.Call(pass)
		if len(res) != 0 {
			text = res[0].Interface().(string)
		} else {
			text = ""
		}
		return text, nil
	} else {
		text := "Введите сообщение повторно.\n"
		return text, errors.New("Message not found")
	}
}

func (sm *StateMachine) GenKeyboard(msg *tgbotapi.MessageConfig) {
	var keyboard tgbotapi.InlineKeyboardMarkup
	log.Println(sm.state)
	resp := sm.states[sm.state].Generate(sm)
	msg.Text = resp.text
	if _,ok := sm.states[sm.state].(Readable); ok {
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		return
	}
	if _,ok := sm.states[sm.state].(Contactable); ok {
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonContact("Телефон"),),)
		return
	}
	var keys [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for _,s := range resp.actions {
		if s.Do == "ROW" {
			if len(row)>0 {
				keys = append(keys, row)
				row = tgbotapi.NewInlineKeyboardRow()
			}
			continue
		}
		log.Println(s)
		row = append(row,tgbotapi.NewInlineKeyboardButtonData(s.Name, s.Do))
	}
	if len(row) > 0 {
		keys = append(keys, row)
	}
	log.Println(keys)
	keyboard = tgbotapi.NewInlineKeyboardMarkup(keys...)
	log.Println(keyboard)
	msg.ReplyMarkup = &keyboard
}

func (sm *StateMachine) Process(update UpdateData) tgbotapi.MessageConfig {
	var err error
	sm.edit = false
	sm.update = update
	sm.user = strconv.FormatInt(update.User,10)
	msg := tgbotapi.NewMessage(update.Chat,"Ошибка")

	sm.state, _ = sm.db.GetState(sm.user)
	pretext, err := sm.ExecAction()
	if err != nil {
		msg.Text = "Ошибка"
	}

	msg.Text = pretext + msg.Text
	sm.GenKeyboard(&msg)
	return msg
}
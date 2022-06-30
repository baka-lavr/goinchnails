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
	user string
	update UpdateData
	state string
	states map[string]State
}

func InitMachine(db db.DataBase) (StateMachine) {
	sm := StateMachine{
		db: db,
		states: map[string]State{
			"Start": Start{},
			"Client_Type": Client_Type{},
			"Client_Master": Client_Master{},
			"Client_Day": Client_Day{},
			"Client_Time": Client_Time{},
			"Client_Confirm": Client_Confirm{},
			"Client_Check": Client_Check{},
			"Client_Delete": Client_Delete{},
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
	//var res []reflect.Value
	if _,ok := sm.states[sm.state].(Readable); ok {
		if sm.update.Message == nil {
			text := "Введите сообщение повторно.\n"
			return text, errors.New("Message not found")
		}
		arg := sm.update.Message.Text
		sm.states[sm.state].(Readable).Read(sm, arg)
		//res = reflect.ValueOf(out)
		return "", nil
	} else {
		if sm.update.Message != nil {
			if !sm.update.Message.From.IsBot {
				//log.Println("MESSAGE")
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
	var keys []tgbotapi.InlineKeyboardButton
	for _,s := range resp.keys {
		log.Println(s)
		keys = append(keys,tgbotapi.NewInlineKeyboardButtonData(s, resp.actions[s]))
	}
	keyboard = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(keys...))
	log.Println(keyboard)
	msg.ReplyMarkup = &keyboard
}

func (sm *StateMachine) Process(update UpdateData) tgbotapi.MessageConfig {
	var err error
	sm.update = update
	sm.user = strconv.FormatInt(update.User,10)
	msg := tgbotapi.NewMessage(update.Chat,"Ошибка")

	sm.state, err = sm.db.GetState(sm.user)
	log.Println(sm.state)
	if err != nil {
		msg.Text = "Ошибка"
	}
	pretext, err := sm.ExecAction()
	if err != nil {
		msg.Text = "Ошибка"
	}

	msg.Text = pretext + msg.Text
	sm.GenKeyboard(&msg)
	return msg
}
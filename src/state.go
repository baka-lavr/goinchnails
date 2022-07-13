package main

import (
	//"log"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/baka-lavr/goinchnails/src/database"
	"strconv"
	"errors"
	"reflect"
	"strings"
)

type Generator interface {
	Action(sm *StateMachine) (string, error)
	KeysResponce(msg *tgbotapi.MessageConfig, resp Respond)
	Generate(sm *StateMachine) Respond

}
type NonSelectable interface {
	Special(sm *StateMachine, text string) string
}

type Selector struct {}
func (state Selector) Action(sm *StateMachine) (string, error) {
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
		var text string
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
func (state Selector) KeysResponce(msg *tgbotapi.MessageConfig, resp Respond) {
	var keyboard tgbotapi.InlineKeyboardMarkup
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
		row = append(row,tgbotapi.NewInlineKeyboardButtonData(s.Name, s.Do))
	}
	if len(row) > 0 {
		keys = append(keys, row)
	}
	keyboard = tgbotapi.NewInlineKeyboardMarkup(keys...)
	msg.ReplyMarkup = &keyboard
}

type Reader struct {
	NonSelectable
}
func (state Reader) Action(sm *StateMachine) (string, error) {
	if sm.update.Message == nil {
		text := "Введите сообщение повторно.\n"
		return text, errors.New("Message not found")
	}
	arg := sm.update.Message.Text
	sm.states[sm.state].(NonSelectable).Special(sm, arg)
	return "", nil
}
func (state Reader) KeysResponce(msg *tgbotapi.MessageConfig, resp Respond) {
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
}

type Contacter struct {
	NonSelectable
}
func (state Contacter) Action(sm *StateMachine) (string, error) {
	if sm.update.Message == nil || sm.update.Message.Contact == nil {
		return "Используй кнопку", errors.New("Message not found")
	}
	sm.states[sm.state].(NonSelectable).Special(sm, sm.update.Message.Contact.PhoneNumber)
	return "", nil
}
func (state Contacter) KeysResponce(msg *tgbotapi.MessageConfig, resp Respond) {
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonContact("Телефон"),),)
}

type Location struct {}

type Action struct {
	Name string
	Do string
}

type Respond struct {
	text string
	actions []Action
}
func NewRespond(text string) Respond {
	res := Respond{text, make([]Action,0)}
	return res
}
func NewRespondList(text string, list []db.List) Respond {
	tx := text
	for i,s := range list {
		tx = tx+"\n"+strconv.Itoa(i+1)+") "+s.Descr
	}
	res := Respond{tx, make([]Action,0)}
	return res
}
func (res *Respond) AddAction(text, action string) {
	res.actions = append(res.actions, Action{text, action})
}
func (res *Respond) AddList(action string,list []db.List) {
	for i,s := range list {
		res.AddAction(s.Name, action+":"+s.ID)
		if (i+1)%3==0 {
			res.AddAction("ROW", "ROW")
		}
	}
	res.AddAction("ROW", "ROW")
}


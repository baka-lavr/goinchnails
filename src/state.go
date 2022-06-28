package main

import (
	//"log"
	"errors"
	"reflect"
	"strings"
	"strconv"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State interface {
	Generate(sm StateMachine) Respond
}
type Readable interface {
	Read(sm StateMachine, arg string)
}

type Respond struct {
	text string
	actions map[string]string
}
func NewRespond(text string) Respond {
	res := Respond{text, make(map[string]string)}
	return res
}
func NewRespondList(text string, list []List) Respond {
	tx := text
	for i,s := range list {
		tx = tx+"\n"+strconv.Itoa(i)+s.Descr
	}
	res := Respond{tx, make(map[string]string)}
	return res
}
func (res *Respond) AddAction(text, action string) {
	res.actions[text] = action
}
func (res *Respond) AddList(list []List) {
	for _,s := range list {
		res.AddAction(s.Name, "Selecting:"+s.ID)
	}
}

type StateMachine struct {
	db DataBase
	user string
	update UpdateData
	state string
	states map[string]State
}

func InitMachine(db DataBase) (StateMachine) {
	sm := StateMachine{
		db: db,
		states: map[string]State{
			"Start": Start{},
			"Client_Type": Client_Type{},
		},
	}
	return sm
}
func (sm *StateMachine) GetUser() string {
	user := sm.update.User
	return strconv.FormatInt(user,10)
}
func (sm *StateMachine) ChangeState(state string) {
	sm.state = state
	sm.db.SetState(sm.GetUser(), state)
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
		sm.states[sm.state].(Readable).Read(*sm, arg)
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
	resp := sm.states[sm.state].Generate(*sm)
	msg.Text = resp.text
	if _,ok := sm.states[sm.state].(Readable); ok {
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		return
	}
	var keys []tgbotapi.InlineKeyboardButton
	for i,s := range resp.actions {
		keys = append(keys,tgbotapi.NewInlineKeyboardButtonData(i, s))
	}
	keyboard = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(keys...))
	msg.ReplyMarkup = &keyboard
}

func (sm *StateMachine) Process(update UpdateData) tgbotapi.MessageConfig {
	var err error
	sm.update = update
	sm.user = strconv.FormatInt(update.User,10)
	msg := tgbotapi.NewMessage(update.Chat,"Ошибка")

	sm.state, err = sm.db.GetState(sm.user)
	if err != nil {
		msg.Text = "Ошибка"
	}
	//_ = sm.Handle()
	pretext, err := sm.ExecAction()
	if err != nil {
		msg.Text = "Ошибка"
	}
	//text := sm.Handle()

	msg.Text = pretext + msg.Text
	//log.Println(sm.actions)
	sm.GenKeyboard(&msg)
	return msg
}




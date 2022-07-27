package main

import (
	"log"
	"errors"
	//"reflect"
	//"strings"
	"strconv"
	"unicode"
	"github.com/baka-lavr/goinchnails/src/database"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

///
func CheckText(text string) bool {
	if text == "" {
		return false
	}
	for _, s := range text {
        if !unicode.IsLetter(s) {
            return false
        }
    }
    return true
}
///

type StateMachine struct {
	db db.DataBase
	notice chan db.Notifier
	states map[string]Generator
	commands map[string]Command
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
		states: map[string]Generator{
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
			"Client_Move": Client_Move{},
			"Client_Move_Day": Client_Move_Day{},
			"Client_Move_Time": Client_Move_Time{},
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
			"Master_Move": Master_Move{},
			"Master_Move_Day": Master_Move_Day{},
			"Master_Move_Time": Master_Move_Time{},
		},
		commands: map[string]Command{
			"/start": StartCommand{},
		},
	}
	return sm
}

func (sm *StateMachine) ChangeState(state string) {
	sm.edit = false
	sm.state = state
	sm.db.SetState(sm.user, state)
	log.Println(sm.state)
}

func (sm *StateMachine) ExecAction() (string, error) {
	//var text string
	res, err := sm.states[sm.state].Action(sm)
	return res, err
}
func (sm *StateMachine) ExecCommand() (string, error) {
	command := sm.commands[sm.update.Message.Text]
	if command == nil {
		return "", errors.New("Wrong Command!")
	}
	res := sm.commands[sm.update.Message.Text].Action(sm)
	return res, nil
}

func (sm *StateMachine) GenKeyboard(msg *tgbotapi.MessageConfig) {
	resp := sm.states[sm.state].Generate(sm)
	msg.Text = resp.text
	sm.states[sm.state].KeysResponce(msg, resp)
	
	//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	
}

func (sm *StateMachine) Process(update UpdateData) tgbotapi.MessageConfig {
	var err error
	var pretext string
	sm.edit = true
	sm.update = update
	sm.user = strconv.FormatInt(update.User,10)
	msg := tgbotapi.NewMessage(update.Chat,"Ошибка")
	var start bool
	if sm.state, start = sm.db.GetState(sm.user); start {
		sm.edit = false
	}
	if sm.update.Command {
		sm.edit = false
		pretext, err = sm.ExecCommand()
	} else {
		pretext, err = sm.ExecAction()
	}
	if err != nil {
		msg.Text = "Ошибка"
	}

	msg.Text = pretext + msg.Text
	sm.GenKeyboard(&msg)
	return msg
}
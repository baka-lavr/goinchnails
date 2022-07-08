package main

import (
	//"log"
	//"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/baka-lavr/goinchnails/src/database"
	"strconv"
)

type State interface {
	Generate(sm *StateMachine) Respond
}
type Readable interface {
	Read(sm *StateMachine, text string) string
}
type Contactable interface {
	Contact(sm *StateMachine, text string)
}
type Locatable interface {
	Locate(sm *StateMachine, text string)
}

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


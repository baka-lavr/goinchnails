package main

import (
	//"log"
	"github.com/baka-lavr/goinchnails/src/database"
	"strconv"
)

type State interface {
	Generate(sm *StateMachine) Respond
}
type Readable interface {
	Read(sm *StateMachine, arg string)
}

type Respond struct {
	text string
	actions map[string]string
	keys []string
}
func NewRespond(text string) Respond {
	res := Respond{text, make(map[string]string), make([]string,0)}
	return res
}
func NewRespondList(text string, list []db.List) Respond {
	tx := text
	for i,s := range list {
		tx = tx+"\n"+strconv.Itoa(i+1)+") "+s.Descr
	}
	res := Respond{tx, make(map[string]string), make([]string,0)}
	return res
}
func (res *Respond) AddAction(text, action string) {
	res.actions[text] = action
	res.keys = append(res.keys, text)
}
func (res *Respond) AddList(action string,list []db.List) {
	for _,s := range list {
		res.AddAction(s.Name, action+":"+s.ID)
	}
}


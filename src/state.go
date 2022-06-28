package main

import (
	//"log"
	"strconv"
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


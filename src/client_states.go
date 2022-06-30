package main

import (
	//"log"
	//"github.com/baka-lavr/goinchnails/src/database"
	"strconv"
)

type Start struct {}
func (Start) Generate(sm *StateMachine) Respond {
	res := NewRespond("Выберите действие")
	res.AddAction("Записаться","BeginningEntry")
	res.AddAction("Мои записи","Checking")
	return res
}
func (Start) BeginningEntry(sm *StateMachine) {
	sm.ChangeState("Client_Type")
}

func (Start) Checking(sm *StateMachine) {
	sm.ChangeState("Client_Check")
}

type Client_Type struct {}
func (Client_Type) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfType()
	res := NewRespondList("Выберите тип услуги:\n", list)
	res.AddList("SelectingType",list)
	res.AddAction("Отменить", "Canceling")
	return res
}
func (Client_Type) SelectingType(sm *StateMachine, tp string) {
	sm.db.EntrySet(sm.user,"type",tp)
	sm.ChangeState("Client_Master")
}
func (Client_Type) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Master struct {}
func (Client_Master) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfMasters(sm.user)
	res := NewRespondList("Выберите мастера:\n", list)
	res.AddList("SelectingMaster",list)
	res.AddAction("Вернуться","Canceling")
	return res
}
func (Client_Master) SelectingMaster(sm *StateMachine, master string) {
	sm.db.EntrySet(sm.user,"master",master)
	sm.ChangeState("Client_Day")
}
func (Client_Master) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Day struct {}
func (Client_Day) Generate(sm *StateMachine) Respond {
	list := sm.db.MasterDays(sm.user)
	res := NewRespondList("Рабочие дни мастера:\n", list)
	res.AddList("SelectingDay",list)
	res.AddAction("Вернуться","Canceling")
	return res
}
func (Client_Day) SelectingDay(sm *StateMachine, day string) {
	sm.db.EntrySet(sm.user,"day",day)
	sm.ChangeState("Client_Time")
}
func (Client_Day) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Time struct {}
func (Client_Time) Generate(sm *StateMachine) Respond {
	list := sm.db.MasterFree(sm.user)
	if len(list) == 0 {
		res := NewRespond("Мастер в этот день занят")
		res.AddAction("Выбрать другой день","Canceling")
		return res
	}
	res := NewRespondList("Свободное время мастера:\n", list)
	res.AddList("SelectingTime",list)
	res.AddAction("Выбрать другой день","Canceling")
	return res
}
func (Client_Time) SelectingTime(sm *StateMachine, time string) {
	sm.db.EntrySet(sm.user,"time",time)
	sm.ChangeState("Client_Confirm")
}
func (Client_Time) Canceling(sm *StateMachine) {
	sm.ChangeState("Client_Day")
}

type Client_Confirm struct{}
func (Client_Confirm) Generate(sm *StateMachine) Respond {
	entry := sm.db.FormEntry(sm.user)
	master := sm.db.GetMaster(entry.Master)
	text := entry.Day+" | "+strconv.Itoa(entry.Time)+"\n"+master.Name
	res := NewRespond(text)
	res.AddAction("Подтвердить","Confirmation")
	res.AddAction("Отменить","Canceling")
	return res
}
func (Client_Confirm) Confirmation(sm *StateMachine) string {
	var text string
	err := sm.db.FinishEntry(sm.user)
	if err != nil {
		text = "Возникла ошибка\n"
	}
	sm.ChangeState("Start")
	return text
}
func (Client_Confirm) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Check struct {}
func (Client_Check) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfEntry(sm.user, "user")
	if len(list) == 0 {
		res := NewRespond("Записей нет")
		res.AddAction("Назад", "Canceling")
		return res
	}
	res := NewRespondList("Ваши записи:\n", list)
	res.AddAction("Удалить запись", "ChooseDeletion")
	res.AddAction("Назад", "Canceling")
	return res
}
func (Client_Check) ChooseDeletion(sm *StateMachine) {
	sm.ChangeState("Client_Delete")
}
func (Client_Check) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Delete struct {}
func (Client_Delete) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfEntry(sm.user, "user")
	if len(list) == 0 {
		res := NewRespond("Записей нет")
		res.AddAction("Назад", "Canceling")
		return res
	}
	res := NewRespondList("Ваши записи:\n", list)
	res.AddList("Deletion", list)
	res.AddAction("Назад", "Canceling")
	return res
}
func (Client_Delete) Deletion(sm *StateMachine, entry string) string {
	var text string
	err := sm.db.DeleteEntry(entry)
	if err != nil {
		text = "Ошибка удаления"
	}
	sm.ChangeState("Client_Check")
	return text
}
func (Client_Delete) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}
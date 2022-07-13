package main

import (
	"log"
	"fmt"
	//"github.com/baka-lavr/goinchnails/src/database"
	//"strconv"
)

type Start struct {
	Selector
}
func (Start) Generate(sm *StateMachine) Respond {
	res := NewRespond("Кто вы?")
	res.AddAction("Клиент","SelectClient")
	res.AddAction("Мастер","SelectMaster")
	return res
}
func (Start) SelectClient(sm *StateMachine) {
	log.Println("CLICK")
	sm.ChangeState("Client_Start")
}
func (Start) SelectMaster(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Client_Start struct {
	Selector
}
func (Client_Start) Generate(sm *StateMachine) Respond {
	res := NewRespond("Выберите действие")
	res.AddAction("Записаться","BeginningEntry")
	res.AddAction("Мои записи","Checking")
	res.AddAction("Сменить тип","Canceling")
	return res
}
func (Client_Start) BeginningEntry(sm *StateMachine) {
	sm.ChangeState("Client_Type")
}

func (Client_Start) Checking(sm *StateMachine) {
	sm.ChangeState("Client_Check")
}
func (Client_Start) Canceling(sm *StateMachine) {
	sm.db.CleanUser(sm.user)
	sm.ChangeState("Start")
}

type Client_Type struct {
	Selector
}
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
	sm.ChangeState("Client_Start")
}

type Client_Master struct {
	Selector
}
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
	sm.ChangeState("Client_Start")
}

type Client_Day struct {
	Selector
}
func (Client_Day) Generate(sm *StateMachine) Respond {
	list := sm.db.MasterDays(sm.user,false)
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
	sm.ChangeState("Client_Start")
}

type Client_Time struct {
	Selector
}
func (Client_Time) Generate(sm *StateMachine) Respond {
	list := sm.db.MasterFree(sm.user,false)
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
	sm.ChangeState("Client_Phone")
}
func (Client_Time) Canceling(sm *StateMachine) {
	sm.ChangeState("Client_Day")
}

type Client_Phone struct {
	Contacter
}
func (Client_Phone) Generate(sm *StateMachine) Respond {
	res := NewRespond("Введите свой номер телефона для связи")
	//res.AddAction("","Read")
	return res
}
func (Client_Phone) Special(sm *StateMachine, text string) string {
	sm.db.EntrySet(sm.user, "phone", text)
	sm.ChangeState("Client_Confirm")
	return ""
}

type Client_Confirm struct{
	Selector
}
func (Client_Confirm) Generate(sm *StateMachine) Respond {
	entry := sm.db.FormEntry(sm.user)
	master := sm.db.GetMaster(entry.Master)
	text := fmt.Sprintf("Запись к %s %s \n%d \n%s \nНа %d:00 в %s",master.Name,master.Forename,master.Phone,master.Address,entry.Time,entry.Day)
	res := NewRespond(text)
	res.AddAction("Подтвердить","Confirmation")
	res.AddAction("Отменить","Canceling")
	return res
}
func (Client_Confirm) Confirmation(sm *StateMachine) string {
	text := "Возникла ошибка\n"
	inx,err := sm.db.FinishEntry(sm.user)
	if err == nil {
		text = ""
		sm.notice<-*inx
	}
	sm.ChangeState("Client_Start")
	return text
}
func (Client_Confirm) Canceling(sm *StateMachine) {
	sm.ChangeState("Client_Start")
}

type Client_Check struct {
	Selector
}
func (Client_Check) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfEntry(sm.user, false)
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
	sm.ChangeState("Client_Start")
}

type Client_Delete struct {
	Selector
}
func (Client_Delete) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfEntry(sm.user, false)
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
	sm.ChangeState("Client_Start")
}
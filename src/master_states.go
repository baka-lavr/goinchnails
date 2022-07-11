package main

import (
	//"strconv"
	//"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"fmt"
	"log"
)

type Master_Start struct {}
func (Master_Start) Generate(sm *StateMachine) Respond {
	res := NewRespond("Выберите операцию")
	master := sm.db.GetMaster(sm.user)
	if master.Name != "" {
		res.AddAction("Редактировать информацию","Changing")
		res.AddAction("Записи", "Checking")
		res.AddAction("ROW","ROW")
		res.AddAction("Прекратить работу", "Deletion")
	} else {
		res.AddAction("Зарегистрироваться","Changing")
	}
	res.AddAction("Сменить тип","Canceling")
	return res
}
func (Master_Start) Changing(sm *StateMachine) {
	sm.ChangeState("Master_Type")
}
func (Master_Start) Checking(sm *StateMachine) {
	sm.ChangeState("Master_Check")
}
func (Master_Start) Deletion(sm *StateMachine) {
	sm.ChangeState("Master_Stop")
}
func (Master_Start) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Master_Type struct {}
func (Master_Type) Generate(sm *StateMachine) Respond {
	list, let := sm.db.SelectedList(sm.user, sm.db.ListOfType())
	res := NewRespondList("",list)
	res.AddList("Switching",list)
	if let {
		res.AddAction("Продолжить", "Continue")
	}
	res.AddAction("Отменить", "Canceling")
	return res
}
func (Master_Type) Switching(sm *StateMachine, sel string) {
	sm.edit = true
	sm.db.SwitchTypes(sm.user, sel)
}
func (Master_Type) Continue(sm *StateMachine) {
	_, let := sm.db.SelectedList(sm.user, sm.db.ListOfType())
	if let {
		sm.ChangeState("Master_Days")
	}
}
func (Master_Type) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Master_Days struct {}
func (Master_Days) Generate(sm *StateMachine) Respond {
	list, let := sm.db.SelectedList(sm.user, sm.db.ListOfDays())
	res := NewRespondList("",list)
	res.AddList("Switching",list)
	if let {
		res.AddAction("Продолжить", "Continue")
	}
	res.AddAction("Отменить", "Canceling")
	return res
}
func (Master_Days) Switching(sm *StateMachine, sel string) {
	sm.edit = true
	sm.db.SwitchTypes(sm.user, sel)
}
func (Master_Days) Continue(sm *StateMachine) {
	_, let := sm.db.SelectedList(sm.user, sm.db.ListOfDays())
	if let {
		sm.ChangeState("Master_Time")
	}
}
func (Master_Days) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Master_Time struct {}
func (Master_Time) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfHours("")
	list = list[:len(list)-1]
	res := NewRespondList("Выберите начало рабочего дня", list)
	res.AddList("SelectingStart", list)
	res.AddAction("Отменить","Canceling")
	return res
}
func (Master_Time) SelectingStart(sm *StateMachine, time string) {
	sm.db.EntrySet(sm.user, "start", time)
	sm.ChangeState("Master_Time_End")
}
func (Master_Time) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Master_Time_End struct {}
func (Master_Time_End) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfHours(sm.user)
	res := NewRespondList("Выберите конец рабочего дня", list)
	res.AddList("SelectingEnd", list)
	res.AddAction("Отменить","Canceling")
	return res
}
func (Master_Time_End) SelectingEnd(sm *StateMachine, time string) {
	sm.db.EntrySet(sm.user, "end", time)
	sm.ChangeState("Master_Name")
}
func (Master_Time_End) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Master_Name struct {}
func (Master_Name) Generate(sm *StateMachine) Respond {
	res := NewRespond("Введите имя")
	res.AddAction("Клиент","Read")
	return res
}
func (Master_Name) Read(sm *StateMachine, text string) string {
	if !CheckText(text) {
		return "Неверный текст"
	}
	sm.db.EntrySet(sm.user, "name", text)
	sm.ChangeState("Master_Forename")
	return ""
}

type Master_Forename struct {}
func (Master_Forename) Generate(sm *StateMachine) Respond {
	res := NewRespond("Введите фамилию")
	res.AddAction("Клиент","Read")
	return res
}
func (Master_Forename) Read(sm *StateMachine, text string) string {
	if !CheckText(text) {
		return "Неверный текст"
	}
	sm.db.EntrySet(sm.user, "forename", text)
	sm.ChangeState("Master_Phone")
	return ""
}

type Master_Phone struct {}
func (Master_Phone) Generate(sm *StateMachine) Respond {
	res := NewRespond("Введите номер телефона \nФормат: +7(**********)")
	//res.AddAction("","Read")
	return res
}
func (Master_Phone) Contact(sm *StateMachine, text string) {
	//_, err := strconv.Atoi(text)
	//if err != nil || len(text) != 10 {
	//	return "Неверный формат"
	//}
	log.Println(text)
	sm.db.EntrySet(sm.user, "phone", text)
	sm.ChangeState("Master_Address")
	//return ""
}

type Master_Address struct {}
func (Master_Address) Generate(sm *StateMachine) Respond {
	res := NewRespond("Введите адрес")
	res.AddAction("","Read")
	return res
}
func (Master_Address) Read(sm *StateMachine, text string) string {
	
	sm.db.EntrySet(sm.user, "address", text)
	sm.ChangeState("Master_Confirm")
	return ""
}

type Master_Confirm struct {}
func (Master_Confirm) Generate(sm *StateMachine) Respond {
	master := sm.db.FormMaster(sm.user)
	text := fmt.Sprintf("%s %s \n%d \n%d:00-%d:00 \nАдрес:%s \nУслуги: \n", master.Name, master.Forename, master.Phone, master.Start, master.End, master.Address)
	for _,s := range master.Services {
		text += s+"\n"
	}
	text += "Рабочие дни: \n"
	for _,s := range master.Days {
		text += s+"\n"
	}
	res := NewRespond(text)
	res.AddAction("Подтвердить", "Confirmation")
	res.AddAction("Отменить", "Cancelation")
	return res
}
func (Master_Confirm) Confirmation(sm *StateMachine) string {
	var text string
	if err := sm.db.CreateMaster(sm.user); err != nil {
		text = "Ошибка"
	}
	sm.ChangeState("Start")
	return text
}
func (Master_Confirm) Cancelation(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Master_Check struct {}
func (Master_Check) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfEntry(sm.user, true)
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
func (Master_Check) ChooseDeletion(sm *StateMachine) {
	sm.ChangeState("Master_Delete")
}
func (Master_Check) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Master_Delete struct {}
func (Master_Delete) Generate(sm *StateMachine) Respond {
	list := sm.db.ListOfEntry(sm.user, true)
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
func (Master_Delete) Deletion(sm *StateMachine, entry string) string {
	var text string
	err := sm.db.DeleteEntry(entry)
	if err != nil {
		text = "Ошибка удаления"
	}
	sm.ChangeState("Master_Check")
	return text
}
func (Master_Delete) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}

type Master_Stop struct {}
func (Master_Stop) Generate(sm *StateMachine) Respond {
	res := NewRespond("Вы точно хотите прекратить работу?")
	res.AddAction("Подтвердить", "Deletion")
	res.AddAction("Назад", "Canceling")
	return res
}
func (Master_Stop) Deletion(sm *StateMachine) {
	sm.db.DeleteMaster(sm.user)
	sm.db.CleanEntries(sm.user)
	sm.ChangeState("Master_Start")
}
func (Master_Stop) Canceling(sm *StateMachine) {
	sm.ChangeState("Master_Start")
}
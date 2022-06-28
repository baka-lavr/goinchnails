package main

type Start struct {}

func (Start) Generate(sm StateMachine) Respond {
	res := NewRespond("Выберите действие")
	res.AddAction("Записаться","BeginningEntry")
	//res.AddAction("Отменить","Canceling")
	return res
}
func (Start) BeginningEntry(sm *StateMachine) {
	sm.ChangeState("Client_Type")
}

func (Start) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Type struct {}

func (Client_Type) Generate(sm StateMachine) Respond {
	list := sm.db.ListOfType()
	res := NewRespondList("Выберите тип услуги:\n", list)
	res.AddList(list)
	res.AddAction("Отменить", "Canceling")
	return res
}
func (Client_Type) Selecting(sm *StateMachine, tp string) {
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
	res.AddList(list)
	res.AddAction("Вернуться","Canceling")
	return res
}
func (Client_Master) Selecting(sm *StateMachine, master string) {
	sm.db.EntrySet(sm.user,"master",master)
	sm.ChangeState("Client_Confirm")
}
func (Client_Master) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}

type Client_Confirm struct{}

func (Client_Confirm) Generate(sm *StateMachine) Respond {
	text, _ := sm.db.EntryShow(sm.user)
	res := NewRespond(text)
	res.AddAction("Вернуться","Canceling")
	return res
}
func (Client_Confirm) Canceling(sm *StateMachine) {
	sm.ChangeState("Start")
}
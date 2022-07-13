package main

type Command interface {
	Action(sm *StateMachine) string
}

type StartCommand struct {}
func (StartCommand) Action(sm *StateMachine) string {
	sm.ChangeState("Start")
	return ""
}
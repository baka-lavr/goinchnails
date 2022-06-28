package main

import (
	"log"
	//"reflect"
)

type Passable interface {
	Exec()
}
type Read interface {
	Foo()
}

type Testing struct {}

func (Testing) Exec() {
	log.Println("Testing")
}

//func (Testing) Foo() {
//	log.Println("Foo")
//}

func main() {
	test := make(map[string]Passable)
	test["one"] = Testing{}
	//test["one"].Exec()
	_, ok := test["one"].(Read)
	log.Println(ok)
	//reflect.ValueOf(test["one"]).MethodByName("Foo").Call([]reflect.Value{})
}


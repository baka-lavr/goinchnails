package main

import (
	//"log"
	//"context"
	//"time"
	//"encoding/base32"
	//"strconv"
)
type Govno interface {}
type Foo struct {}

func main() {
	f,_ := Foo{}.(*Govno)
}


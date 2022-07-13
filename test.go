package main

import (
	//"log"
	//"context"
	//"time"
	//"encoding/base32"
	//"strconv"
)
type Govno interface {
	
}
type Foo struct {}

func main() {
	ctx := context.Background()
	db := InitDB()
	list := db.Client.SMembers(ctx,"masters")
	for _,s := range list {
		db.DeleteMaster(list)
	}
}


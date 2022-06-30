package db

import (
	//"encoding/base32"
	//"reflect"
	//"encoding/json"
	"strconv"
	//"github.com/go-redis/redis/v9"
	"context"
	//"time"
	"log"
	//"errors"
	//"github.com/fatih/structs"
)

type Master struct {
	Name string `redis:"name"`
	Start int `redis:"start"`
	End int `redis:"end"`
}

func (db DataBase) ListOfMasters(user string) []List {
	ctx := context.Background()
	tp := db.Client.HGet(ctx,"user:"+user, "type").Val()
	all := db.Client.SMembers(ctx,"masters:"+tp).Val()
	log.Println("masters:"+tp)
	var list []List
	for _,s := range all {
		var item List
		var master Master
		data := db.Client.HGetAll(ctx,"master:"+s)
		data.Scan(&master)
		item.ID = s
		item.Name = master.Name
		item.Descr = master.Name+"\nНачало работы: "+strconv.Itoa(master.Start)+"\nКонец работы: "+strconv.Itoa(master.End)
		list = append(list,item)
		log.Println(data)
	}
	log.Println(list)
	return list
}

func (db DataBase) GetMaster(master string) Master {
	ctx := context.Background()
	var value Master
	data := db.Client.HGetAll(ctx,"master:"+master)
	data.Scan(&value)
	return value
}

func (db DataBase) MasterDays(user string) []List {
	ctx := context.Background()
	id := db.Client.HGet(ctx, "user:"+user, "master").Val()
	var list []List
	days := db.Client.SMembers(ctx,"master-days:"+id).Val()
	for _,s := range days {
		//s := val_loop.Type().Field(i).Name
		item := List{s,s,s}
		list = append(list, item)
	}
	return list
}

func (db DataBase) MasterFree(user string) []List {
	ctx := context.Background()
	day := db.Client.HGet(ctx, "user:"+user, "day").Val()
	master := db.Client.HGet(ctx, "user:"+user, "master").Val()
	m := db.GetMaster(master)
	time := make(map[int]bool)
	for i := m.Start; i<m.End; i++ {
		time[i] = true
	}
	all := db.Client.SMembers(ctx,"master-entry:"+master).Val()
	for _,s := range all {
		if d := db.Client.HGet(ctx, "entry:"+s, "day").Val(); d != day {
			log.Println(d)
			continue}
		e_time, _ := strconv.Atoi(db.Client.HGet(ctx, "entry:"+s, "time").Val())
		delete(time, e_time)
	}
	if len(time) == 0 {
		return nil
	}
	var list []List
	for i,_ := range time {
		str := strconv.Itoa(i)
		val := List{str, str, str+":00"}
		list = append(list, val)
	}
	return list
}
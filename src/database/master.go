package db

import (
	//"encoding/base32"
	//"reflect"
	//"encoding/json"
	"strconv"
	"github.com/go-redis/redis/v9"
	"context"
	//"time"
	"log"
	//"errors"
	"fmt"
	"github.com/fatih/structs"
)

type Master struct {
	Name string `redis:"name" structs:"name"`
	Forename string `redis:"forename" structs:"forename"`
	Phone int `redis:"phone" structs:"phone"`
	Start int `redis:"start" structs:"start"`
	End int `redis:"end" structs:"end"`
	Address string `redis:"address" structs:"address"`
	Services []string `redis:"-" structs:"-"`
	Days []string `redis:"-" structs:"-"`
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
		item.Descr = fmt.Sprintf("%s\nНачало работы: %d\nКонец работы: %d\nАдрес: %s",master.Name,master.Start,master.End,master.Address)
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

func (db DataBase) CurrentTypes(user string) ([]List, bool) {
	ctx := context.Background()
	one := false
	list := db.ListOfType()
	for i,s := range list {
		set := db.Client.HGet(ctx, "user:"+user, s.Name).Val()
		if set == "true" {
			list[i].Descr += " ВЫБРАНО"
			one = true
		}
	}
	return list, one
}

func (db DataBase) SelectedList(user string, list []List) ([]List, bool) {
	ctx := context.Background()
	one := false
	for i,s := range list {
		set := db.Client.HGet(ctx, "user:"+user, s.Name).Val()
		if set == "true" {
			list[i].Descr += " ВЫБРАНО"
			one = true
		}
	}
	return list, one
}

func (db DataBase) SwitchTypes(user, arg string) {
	ctx := context.Background()
	set := db.Client.HGet(ctx, "user:"+user, arg).Val()
	log.Println(set)
	if set == "true" {
		db.EntrySet(user,arg,"false")
	} else {
		db.EntrySet(user,arg,"true")
	}
}

func (db DataBase) FormMaster(user string) Master {
	ctx := context.Background()
	var master Master
	data := db.Client.HGetAll(ctx,"user:"+user)
	data.Scan(&master)
	serv := db.ListOfType()
	for _,s := range serv {
		check := db.Client.HGet(ctx,"user:"+user,s.Name).Val()
		if check == "true" {
			master.Services = append(master.Services, s.Name)
		}
	}
	days := db.ListOfDays()
	for _,s := range days {
		check := db.Client.HGet(ctx,"user:"+user,s.Name).Val()
		log.Println(s.Name)
		if check == "true" {
			master.Days = append(master.Days, s.Name)
		}
	}
	log.Println(master)
	return master
}

func (db DataBase) DeleteMaster(user string) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "id", user)
	res,_ := db.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		var err error
		id := ctx.Value("id").(string)
		_,err = pipe.Del(ctx, "master:"+id).Result()
		_,err = pipe.SRem(ctx, "masters", id).Result()
		for _,s := range db.ListOfType() {
			_,err = pipe.SRem(ctx, "masters:"+s.Name, id).Result()
		}
		for _,s := range db.ListOfDays() {
			_ = pipe.SRem(ctx, "master-days:"+id, s.Name)
		}
		return err
	})
	log.Println(res)
}

func (db DataBase) CreateMaster(user string) error {
	ctx := context.Background()
	if check := db.Client.HGetAll(ctx, "master:"+user).Val(); len(check) > 0 {
		db.DeleteMaster(user)
	}
	master := db.FormMaster(user)
	ctx = context.WithValue(ctx, "id", user)
	ctx = context.WithValue(ctx, "struct", master)
	res,err := db.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		var err error
		id := ctx.Value("id").(string)
		master := ctx.Value("struct").(Master)
		log.Println(structs.Map(master))
		_,err = pipe.HSet(ctx, "master:"+id, structs.Map(master)).Result()
		_,err = pipe.SAdd(ctx, "masters", id).Result()
		for _,s := range master.Services {
			_,err = pipe.SAdd(ctx, "masters:"+s, id).Result()
		}
		for _,s := range master.Days {
			_ = pipe.SAdd(ctx, "master-days:"+id, s)
		}
		return err
	})
	log.Println(res)
	return err
}
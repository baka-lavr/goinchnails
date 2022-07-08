package db

import (
	//"encoding/base32"
	//"reflect"
	//"encoding/json"
	"strconv"
	"github.com/go-redis/redis/v9"
	"context"
	"time"
	//"log"
	//"errors"
	//"github.com/fatih/structs"
)

type List struct {
	ID string
	Name string
	Descr string
}

type Notifier struct {
	User int64
	Text string
}

type DataBase struct {
	Client *redis.Client
}

func InitDB() DataBase {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0,})
	return DataBase{rdb}
}

func (db DataBase) GetState(user string) (string, error) {
	ctx := context.Background()
	//log.Println(user)
	state := db.Client.HGet(ctx,"user:"+user, "state").Val()
	if state == "" {
		_, err := db.Client.HSet(ctx, "user:"+user, "state", "Start").Result()
		if err != nil {
			return "", err
		}
		return "Start", nil
	}
	return state, nil
}

func (db DataBase) SetState(user string, state string) error {
	ctx := context.Background()
	_, err := db.Client.HSet(ctx, "user:"+user, "state", state).Result()
	if err != nil {
		return err
	}
	_ = db.Client.Expire(ctx, "user:"+user, time.Hour)
	return nil
}

func (db DataBase) ListOfType() []List {
	ctx := context.Background()
	all, _ := db.Client.SMembers(ctx,"types").Result()
	var list []List
	for _,s := range all {
		var item List
		item.ID = s
		item.Name = s
		item.Descr = s
		list = append(list,item)
	}
	return list
}

func (db DataBase) ListOfDays() []List {
	ctx := context.Background()
	all, _ := db.Client.ZRange(ctx,"days",0,-1).Result()
	var list []List
	for _,s := range all {
		var item List
		item.ID = s
		item.Name = s
		item.Descr = s
		list = append(list,item)
	}
	return list
}

func(db DataBase) ListOfHours(user string) []List {
	ctx := context.Background()
	start := 8
	end := 20
	if user != "" {
		start,_ = strconv.Atoi(db.Client.HGet(ctx, "user:"+user, "start").Val())
		start++
	}
	var list []List
	for i := start; i<end; i++ {
		data := strconv.Itoa(i)
		item := List{data,data,data+":00"}
		list = append(list, item)
	}
	return list
}

func (db DataBase) GetLastMessage(user string) int {
	ctx := context.Background()
	msg,_ := db.Client.HGet(ctx, "user:"+user, "message").Int()
	return msg
}

func (db DataBase) SetLastMessage(user int64, msg int) {
	data_u := strconv.FormatInt(user,10)
	data_m := strconv.FormatInt(int64(msg),10)
	db.EntrySet(data_u, "message", data_m)
}

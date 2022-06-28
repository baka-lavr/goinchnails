package main

import (
	"github.com/go-redis/redis/v9"
	"context"
	"time"
	"log"
	//"errors"
)

type List struct {
	ID string
	Name string
	Descr string
}

type Entry struct {
	Type string `redis:"type"`
	Master string `redis:"master"`
	Time int `redis:"time"`
	Day int `redis:"day"`
}

type Master struct {
	Name string `redis:"Name"`
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
	log.Println(user)
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

func (db DataBase) EntrySet(user, key, value string) error {
	ctx := context.Background()
	err := db.Client.HSet(ctx, "user:"+user, key, value).Err()
	if err != nil {
		return err
	}
	return nil
}

func (db DataBase) EntryShow(user string) (string, error) {
	ctx := context.Background()
	value := db.Client.HGetAll(ctx, "user:"+user)
	if value.Err() != nil {
		return "", value.Err()
	}
	var entry Entry
	value.Scan(&entry)
	text := "DEBUG\n"+entry.Type+"\n"+entry.Master
	log.Println(text)
	return text, nil
}

func (db DataBase) ListOfType() []List {
	ctx := context.Background()
	all, _ := db.Client.SMembers(ctx,"types").Result()
	var list []List
	for _,s := range all {
		var item List
		item.ID = s
		item.Name = s
		list = append(list,item)
	}
	return list
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
		list = append(list,item)
		log.Println(data)
	}
	log.Println(list)
	return list
}
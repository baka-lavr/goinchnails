package db

import (
	//"encoding/base32"
	//"reflect"
	//"encoding/json"
	//"strconv"
	"github.com/go-redis/redis/v9"
	"context"
	"time"
	"log"
	//"errors"
	//"github.com/fatih/structs"
)

type List struct {
	ID string
	Name string
	Descr string
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


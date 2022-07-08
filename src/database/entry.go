package db

import (
	"encoding/base32"
	"fmt"
	//"reflect"
	//"encoding/json"
	"strconv"
	"github.com/go-redis/redis/v9"
	"context"
	"time"
	"log"
	"errors"
	"github.com/fatih/structs"
)

type Entry struct {
	Type string `redis:"type" structs:"type"`
	Master string `redis:"master" structs:"master"`
	Time int `redis:"time" structs:"time"`
	Day string `redis:"day" structs:"day"`
	Client string `redis:"client" structs:"client"`
	Phone int `redis:"phone" structs:"phone"`
}

func (db DataBase) EntrySet(user, key, value string) error {
	ctx := context.Background()
	err := db.Client.HSet(ctx, "user:"+user, key, value).Err()
	if err != nil {
		return err
	}
	return nil
}

func (db DataBase) FormEntry(user string) Entry {
	ctx := context.Background()
	var entry Entry
	value := db.Client.HGetAll(ctx, "user:"+user)
	value.Scan(&entry)
	entry.Client = user
	log.Println(entry)
	return entry
}

func (db DataBase) FinishEntry(user string) (*Notifier,error) {
	ctx := context.Background()
	free := db.MasterFree(user)
	time := db.Client.HGet(ctx, "user:"+user, "time").Val()
	for _,s := range free {
		if s.ID == time {
			entry := db.FormEntry(user)
			id,err := db.CreateEntry(entry)
			return id,err
		}
	}
	return nil,errors.New("Ошибка")
}

func (db DataBase) CreateEntry(entry Entry) (*Notifier,error) {
	ctx := context.Background()
	data, _ := time.Now().GobEncode()
	id := base32.StdEncoding.EncodeToString(data)
	//base32.StdEncoding.Encode(id, data)
	check := db.Client.HGet(ctx, "entry:"+id, "client").Val()
	if check != "" {
		return nil,errors.New("Ошибка")
	}
	ctx = context.WithValue(ctx, "id", id)
	ctx = context.WithValue(ctx, "struct", entry)
	res,err := db.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		var err error
		id := ctx.Value("id").(string)
		entry := ctx.Value("struct").(Entry)
		log.Println(id)
		_,err = pipe.HSet(ctx, "entry:"+id, structs.Map(entry)).Result()
		_,err = pipe.SAdd(ctx, "entries", id).Result()
		_,err = pipe.SAdd(ctx, "user-entry:"+entry.Client, id).Result()
		_,err = pipe.SAdd(ctx, "master-entry:"+entry.Master, id).Result()
		return err
	})
	log.Println(res)
	master,_ := strconv.ParseInt(entry.Master,10,0)
	not := Notifier{master, fmt.Sprintf("К вам записались на %d в %s", entry.Time, entry.Day),}
	return &not,err
}

func (db DataBase) DeleteEntry(entry string) error {
	ctx := context.Background()
	str := db.GetEntry(entry)
	ctx = context.WithValue(ctx, "id", entry)
	ctx = context.WithValue(ctx, "struct", str)
	res,err := db.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		var err error
		id := ctx.Value("id").(string)
		entry := ctx.Value("struct").(Entry)
		_,err = pipe.SRem(ctx, "entries", id).Result()
		_,err = pipe.SRem(ctx, "user-entry:"+entry.Client, id).Result()
		_,err = pipe.SRem(ctx, "master-entry:"+entry.Master, id).Result()
		_,err = pipe.Del(ctx, "entry:"+id).Result()
		return err
	})
	log.Println(res)
	return err
}

func (db DataBase) CleanEntries(master string) {
	ctx := context.Background()
	list := db.Client.SMembers(ctx, "master-entry:"+master).Val()
	for _,s := range list {
		_ = db.DeleteEntry(s)
	}
}

func (db DataBase) GetEntry(entry string) Entry {
	ctx := context.Background()
	var value Entry
	data := db.Client.HGetAll(ctx,"entry:"+entry)
	data.Scan(&value)
	return value
}

func (db DataBase) ListOfEntry(user string, who bool) []List {
	ctx := context.Background()
	line := "user"
	if who {
		user = "master"
	}
	all, _ := db.Client.SMembers(ctx,line+"-entry:"+user).Result()
	var list []List
	for _,s := range all {
		entry := db.GetEntry(s)
		master := db.GetMaster(entry.Master)
		var item List
		item.ID = s
		item.Name = strconv.Itoa(entry.Time)+":00|"+entry.Day
		item.Descr = strconv.Itoa(entry.Time)+":00|"+entry.Day+"\n"+master.Name+"\n"
		if who {
			item.Descr += strconv.Itoa(entry.Phone)+"\n"
		} else {
			item.Descr += strconv.Itoa(master.Phone)+"\n"
		}
		list = append(list,item)
	}
	return list
}

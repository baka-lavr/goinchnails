package main

import (
	"log"
	"time"
	"encoding/base32"
)

func main() {
	data, _ := time.Now().GobEncode()
	id := base32.StdEncoding.EncodeToString(data)
	log.Println(id)
}


package main

import (
	"fmt"
	"log"

	"github.com/takayukioda/pemstore"
)

func main() {
	store := pemstore.New("athletics", true)
	key := "pemstore-test"
	exists, err := store.Exists(key)
	if err != nil {
		log.Fatalln(err)
	}

	if exists {
		fmt.Println("Exists; getting value")
		value, err := store.Get(key, true)
		if err != nil {
			log.Fatalln("Failure during getting process", err)
		}
		fmt.Println(value)
	} else {
		fmt.Println("Not exists; storing value")
		if err := store.Store(key, []byte("Some random text"), false); err != nil {
			log.Fatalln("Failure during storing process", err)
		}
	}
}

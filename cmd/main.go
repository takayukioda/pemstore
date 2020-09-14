package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/takayukioda/pemstore"
)

const DEFAULT_PROFILE = "athletics"

func usage() string {
	return "[-profile <profile>]"
}

func main() {
	profile := flag.String("profile", "", "AWS profile to use")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println(usage())
		os.Exit(1)
	}

	if profile == nil {
		*profile = DEFAULT_PROFILE
	}
	store := pemstore.New(profile, true)

	switch args[0] {
	case "get":
		key := args[1]
		exists, err := store.Exists(key)
		if err != nil {
			log.Fatalln(err)
		}
		if !exists {
			log.Println("Couldn't find specified key:", key)
		}
		value, err := store.Get(key, true)
		if err != nil {
			log.Fatalln("Failure during getting process", err)
		}
		fmt.Println(value)
		os.Exit(0)
	case "put":
		key := args[1]
		exists, err := store.Exists(key)
		if err != nil {
			log.Fatalln(err)
		}
		if exists {
			log.Fatalln("Specified key already exists:", key)
		}
		if err := store.Store(key, []byte("Some random text"), false); err != nil {
			log.Fatalln("Failure during storing process", err)
		}
	default:
		keys, err := store.List()
		if err != nil {
			log.Fatalln("Failure during listing process", err)
		}
		for _, key := range keys {
			fmt.Println(key)
		}
		os.Exit(0)
	}

	/*
	key := "pemstore-test"
	fmt.Println("Not exists; storing value")
	if err := store.Store(key, []byte("Some random text"), false); err != nil {
		log.Fatalln("Failure during storing process", err)
	}
	*/
}

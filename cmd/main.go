package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/takayukioda/pemstore"
)

func usage() string {
	return "pemstore [-profile <profile>] <get / list / store>"
}

const (
	EXIT_OK          = 0
	EXIT_ERR_UNKNOWN = 1
	EXIT_ERR_KNOWN   = 2
)

func main() {
	profile := flag.String("profile", "", "AWS profile to use")
	mfa := flag.Bool("mfa", false, "MFA enabled")
	// TODO: Move them into sub command option
	force := flag.Bool("force", false, "Do action forcefully; avaialble for store and delete")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println(usage())
		os.Exit(1)
	}

	store := pemstore.New(profile, *mfa, nil)

	switch args[0] {
	case "get":
		key := args[1]
		exists, err := store.Exists(key)
		if err != nil {
			log.Fatalln(err)
		}
		if !exists {
			log.Println("Couldn't find specified key:", key)
			os.Exit(EXIT_ERR_KNOWN)
		}
		value, err := store.Get(key, true)
		if err != nil {
			log.Fatalln("Failure during getting process", err)
		}
		fmt.Println(value)
		os.Exit(EXIT_OK)
	case "store":
		// FIXME: Fix to store pem key
		key := args[1]
		exists, err := store.Exists(key)
		if err != nil {
			log.Fatalln(err)
		}
		if exists && !(*force) {
			log.Println("Specified key already exists:", key)
			os.Exit(EXIT_ERR_KNOWN)
		}
		if err := store.Store(key, []byte("Some random text"), false); err != nil {
			log.Fatalln("Failure during storing process", err)
		}
	case "list":
		keys, err := store.List()
		if err != nil {
			log.Fatalln("Failure during listing process", err)
		}
		for _, key := range keys {
			fmt.Println(key)
		}
	case "delete":
		key := args[1]
		exists, err := store.Exists(key)
		if err != nil {
			log.Fatalln(err)
		}
		if !exists {
			log.Println("Specfied key not found:", key)
			os.Exit(EXIT_ERR_KNOWN)
		}
		if !(*force) {
			log.Println("Found specified key:", key)
			log.Println("Add `-force` option to delete")
			os.Exit(EXIT_OK)
		}
		if err := store.Remove(key); err != nil {
			log.Fatalln("Failure during deleting process", err)
		}
		fmt.Println("Deleted pem:", key)
		os.Exit(EXIT_OK)
	default:
		usage()
		os.Exit(EXIT_ERR_KNOWN)
	}
}

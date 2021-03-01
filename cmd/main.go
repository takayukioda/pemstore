package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/takayukioda/pemstore"
)

func usage() string {
	return "pemstore [-profile <profile>] <get / list / store>"
}

const (
	EXIT_OK          = 0
	EXIT_ERR_KNOWN   = 1
	EXIT_ERR_UNKNOWN = 2
)

func main() {
	profile := flag.String("profile", "", "AWS profile to use")
	mfa := flag.Bool("mfa", false, "MFA enabled")
	// TODO: Move them into sub command option
	force := flag.Bool("force", false, "Do action forcefully; avaialble for store and delete")
	flag.Parse()
	args := flag.Args()

	if *profile == "" {
		*profile = os.Getenv("AWS_PROFILE")
	}

	storepath := filepath.Join(os.Getenv("HOME"), ".ssh", "pemstore")
	if _, err := os.Stat(storepath); os.IsNotExist(err) {
		if err := os.MkdirAll(storepath, 0755); err != nil {
			log.Println("Failed to create pemstore at", storepath)
			os.Exit(EXIT_ERR_KNOWN)
		}
	} else if err != nil {
		log.Println("Failure during initialize storepath")
		os.Exit(EXIT_ERR_UNKNOWN)
	}

	if len(args) < 1 {
		fmt.Println(usage())
		os.Exit(EXIT_ERR_KNOWN)
	}

	store := pemstore.New(profile, *mfa, nil)

	switch args[0] {
	case "get":
		key := args[1]
		path := filepath.Join(storepath, key)

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			log.Println("File already exists; clean before get:", path)
			os.Exit(EXIT_ERR_KNOWN)
		}
		exists, err := store.Exists(key)
		if err != nil {
			log.Println(err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		if !exists {
			log.Println("Couldn't find specified key:", key)
			os.Exit(EXIT_ERR_KNOWN)
		}
		value, err := store.Download(key, true)
		if err != nil {
			log.Println("Failure during getting process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		if err := ioutil.WriteFile(path, []byte(value), 0600); err != nil {
			log.Println("Failure during writing file process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		fmt.Println("Got pem file to the local")
		fmt.Println("Key:", key)
		fmt.Println("Stored in:", path)
		os.Exit(EXIT_OK)
	case "store":
		// FIXME: Fix to store pem key
		key := args[1]
		path := filepath.Join(storepath, key)
		if len(args) >= 3 {
			var err error
			path, err = filepath.Abs(args[2])
			if err != nil {
				log.Println("Failure during path retrieve process", err)
				os.Exit(EXIT_ERR_UNKNOWN)
			}
		}
		exists, err := store.Exists(key)
		if err != nil {
			log.Println(err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		if exists && !(*force) {
			log.Println("Specified key already exists:", key)
			os.Exit(EXIT_ERR_KNOWN)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Println("No such file: ", path)
			os.Exit(EXIT_ERR_KNOWN)
		}

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println("Failure during reading file process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		if err := store.Store(key, bytes, (*force)); err != nil {
			log.Println("Failure during storing process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		fmt.Println("Stored pem into pemstore")
		fmt.Println("Key:", key)
		fmt.Println("File:", path)
		os.Exit(EXIT_OK)
	case "list":
		keys, err := store.List()
		if err != nil {
			log.Println("Failure during listing process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		for _, key := range keys {
			fmt.Println(key)
		}
		os.Exit(EXIT_OK)
	case "clean":
		key := args[1]
		path := filepath.Join(storepath, key)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Println("No file to clean up", path)
			os.Exit(EXIT_OK)
		}
		if !(*force) {
			log.Println("Found specified file in pemstore:", path)
			log.Println("Add `-force` option to delete")
			os.Exit(EXIT_OK)
		}
		if err := os.Remove(path); err != nil {
			log.Println("Failure during cleaning process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		fmt.Println("Clean up downloaded file:", path)
		os.Exit(EXIT_OK)
	case "delete":
		key := args[1]
		exists, err := store.Exists(key)
		if err != nil {
			log.Println(err)
			os.Exit(EXIT_ERR_UNKNOWN)
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
			log.Println("Failure during deleting process", err)
			os.Exit(EXIT_ERR_UNKNOWN)
		}
		fmt.Println("Deleted pem:", key)
		os.Exit(EXIT_OK)
	default:
		usage()
		os.Exit(EXIT_ERR_KNOWN)
	}
}

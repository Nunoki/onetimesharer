package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Nunoki/onetimesharer/internal/pkg/filestorage"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
	"github.com/Nunoki/onetimesharer/internal/pkg/server"
	"github.com/Nunoki/onetimesharer/internal/pkg/sqlite"
	"github.com/Nunoki/onetimesharer/pkg/aescfb"
)

const defaultPortHTTP uint = 8000
const defaultPortHTTPS uint = 443

var (
	ctx = context.Background()
)

func main() {
	conf, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	encKey := encryptionKey()
	encrypter := aescfb.New(encKey)

	var store server.Storer
	if *conf.JSONFile {
		store, err = filestorage.New(encrypter)
	} else {
		store, err = sqlite.New(ctx, encrypter)
	}
	if err != nil {
		log.Fatal(err)
	}

	server := server.New(conf, store)
	// TODO: middleware to prevent large payload attack

	// Perform graceful shutdown when interrupted from shell
	go func() {
		fmt.Fprintf(os.Stdout, "Listening on port %d\n", *conf.Port)
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	}
	log.Print("Server exited properly")
}

// DOCME
func encryptionKey() (key string) {
	key = os.Getenv("OTS_ENCRYPTION_KEY")
	if len(key) == 0 {
		key = randomizer.String(32)
		fmt.Fprintf(
			os.Stdout,
			"Generated encryption key is: %s (set env variable OTS_ENCRYPTION_KEY to use custom key)\n",
			key,
		)
		return
	}

	if len(key) != 32 {
		fmt.Fprintf(
			os.Stderr,
			"Provided encryption key must be 32 characters long, is %d\n",
			len(key),
		)
		os.Exit(1)
	}

	return
}

// processArgs processes passed arguments and sets up variables appropriately. If a conflict occurs
// with flag configuration, an error is being output to stderr, and the program exits.
func processArgs() (server.Config, error) {
	conf := server.Config{}

	conf.Certfile = flag.String(
		"certfile",
		"",
		"Path to certificate file, required when running on HTTPS",
	)
	conf.HTTPS = flag.Bool(
		"https",
		false,
		"Whether to run on HTTPS (requires --certfile and --keyfile)",
	)
	conf.JSONFile = flag.Bool(
		"json",
		false,
		"Use a JSON file as storage instead of the default SQLite database",
	)
	conf.Keyfile = flag.String(
		"keyfile",
		"",
		"Path to key file, required when running on HTTPS",
	)
	conf.Port = flag.Uint(
		"port",
		0,
		fmt.Sprintf(
			"Port to run on (default %d for HTTP, %d for HTTPS)",
			defaultPortHTTP,
			defaultPortHTTPS,
		),
	)
	flag.Parse()

	if *conf.HTTPS && (*conf.Certfile == "" || *conf.Keyfile == "") {
		return server.Config{}, errors.New("running on HTTPS requires the certification file and key file (see --help)")
	}

	if *conf.Port == 0 {
		if *conf.HTTPS {
			*conf.Port = defaultPortHTTPS
		} else {
			*conf.Port = defaultPortHTTP
		}
	}

	return conf, nil
}

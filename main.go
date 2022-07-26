package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Nunoki/onetimesharer/internal/pkg/aescfb"
	"github.com/Nunoki/onetimesharer/internal/pkg/filestorage"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
	"github.com/Nunoki/onetimesharer/internal/pkg/server"
)

const defaultPortHTTP uint = 8000
const defaultPortHTTPS uint = 443

func main() {
	conf, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	encKey := encryptionKey()
	encrypter := aescfb.New(encKey)
	store := filestorage.New(encrypter)

	server := server.New(conf, store)
	server.Serve()
}

// DOCME
func encryptionKey() (key string) {
	key = os.Getenv("OTS_ENCRYPTION_KEY")
	if len(key) == 0 {
		key = randomizer.RandStr(32)
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

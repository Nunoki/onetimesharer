package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Nunoki/onetimesharer/internal/pkg/aescfb"
	"github.com/Nunoki/onetimesharer/internal/pkg/filestorage"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
	"github.com/Nunoki/onetimesharer/internal/pkg/server"
	"github.com/pborman/getopt/v2"
)

const defaultPort uint = 8000

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

	help := getopt.BoolLong("help", 'h', "Display help")

	conf.Certfile = getopt.String('c', "", "Path to certificate file, required when running on HTTPS")
	conf.HTTPS = getopt.Bool('s', "Secure; Whether to run on HTTPS (requires --certfile and --keyfile)")
	conf.Keyfile = getopt.String('k', "", "Path to key file, required when running on HTTPS")
	conf.Port = getopt.Uint('p', defaultPort, "Port to run on")
	getopt.Parse()

	if *help {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	if *conf.HTTPS && (*conf.Certfile == "" || *conf.Keyfile == "") {
		return server.Config{}, errors.New("running on HTTPS requires the certification file and key file (see --help)")
	}

	return conf, nil
}

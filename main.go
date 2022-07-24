package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Nunoki/onetimesharer/internal/pkg/filestorage"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
)

const defaultPort uint = 8000

var passphrase string

type config struct {
	port     *uint
	https    *bool
	certfile *string
	keyfile  *string
}

type tplData struct {
	ShareURL  string
	SecretKey string
	ErrorMsg  string
}

type server struct {
	config config
	store  store
}

type jsonOutput struct {
	Secret string `json:"secret"`
}

type store interface {
	SaveSecret(secret string) (string, error)
	ValidateSecret(key string) bool
	ReadSecret(key string) (string, error)
}

func main() {
	conf, err := processFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	passphrase = randomizer.RandStr(32)
	fmt.Fprint(os.Stdout, "Passphrase is: "+passphrase+"\n")
	store := filestorage.NewClient(passphrase)

	server := NewServer(conf, store)
	server.Serve(*conf.port)
}

// processFlags processes passed arguments and sets up variables appropriately. If a conflict occurs
// with flag configuration, an error is being output to stderr, and the program exits.
func processFlags() (config, error) {
	conf := config{}

	conf.port = flag.Uint("port", defaultPort, "Port to run on")
	conf.https = flag.Bool("https", false, "Whether to run on HTTPS (requires --certfile and --keyfile)")
	conf.certfile = flag.String("certfile", "", "Path to certificate file, required when running on HTTPS")
	conf.keyfile = flag.String("keyfile", "", "Path to key file, required when running on HTTPS")
	flag.Parse()

	if *conf.https && (*conf.certfile == "" || *conf.keyfile == "") {
		return config{}, errors.New("running on HTTPS requires the certification file and key file (see --help)")
	}

	return conf, nil
}

// DOCME
func NewServer(c config, s store) server {
	server := server{
		config: c,
		store:  s,
	}
	return server
}

// DOCME
func (s server) Serve(port uint) {
	// TODO: test all endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// this is because the "/" pattern of HandleFunc matches everything
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			handleIndex(w, r)
		} else if r.Method == "POST" {
			handlePost(w, r, s.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	// #show_url
	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handleShow(w, r, s.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleFetchSecret(w, r, s.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	portStr := strconv.Itoa(int(*s.config.port))
	log.Print("Listening on port " + portStr)
	log.Fatal(http.ListenAndServe(":"+portStr, nil))
}

// handleIndex serves the default page for creating a new secret
func handleIndex(w http.ResponseWriter, _ *http.Request) {
	outputTpl(w, tplData{})
}

// handlePost stores the posted secret and outputs the generated key for reading it
func handlePost(w http.ResponseWriter, r *http.Request, s store) {
	secret := r.FormValue("secret")
	if secret == "" {
		http.Error(w, "failed to read posted content", http.StatusBadRequest)
		return
	}

	key, err := s.SaveSecret(secret)
	if err != nil {
		log.Print(err)
		http.Error(w, "failed to save secret", http.StatusInternalServerError)
		return
	}

	shareURL := "http://" + r.Host + "/show?key=" + key
	data := tplData{
		// #show_url
		ShareURL: shareURL,
	}
	outputTpl(w, data)
}

// handleShow shows the button that displays the secret
func handleShow(w http.ResponseWriter, r *http.Request, s store) {
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "key not specified", http.StatusBadRequest)
		return
	}

	ok := s.ValidateSecret(key)
	if !ok {
		data := tplData{
			ErrorMsg: "Could not find requested secret",
		}
		outputTpl(w, data)
		return
	}

	data := tplData{
		SecretKey: key,
	}
	outputTpl(w, data)
}

// handleFetchSecret outputs the content of the secret in JSON format
func handleFetchSecret(w http.ResponseWriter, r *http.Request, s store) {
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "key not specified", http.StatusBadRequest)
		return
	}

	secret, err := s.ReadSecret(key)
	if err != nil {
		log.Print(err)
		http.Error(w, "failed to read secret", http.StatusInternalServerError)
		return
	}

	data := jsonOutput{
		Secret: secret,
	}
	output, _ := json.Marshal(data)
	w.Header().Set("Content-type", "application/json")
	w.Write(output)
}

// outputTpl parses the index.html file and outputs it to the w writer, passing the data to it
func outputTpl(w http.ResponseWriter, data tplData) {
	tpl := template.Must(template.New("").Parse(indexHTML()))
	err := tpl.Execute(w, data)

	if err != nil {
		log.Print(err)
	}
}

package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Config struct {
	Certfile *string
	JSONFile *bool
	HTTPS    *bool
	Keyfile  *string
	Port     *uint
}

type tplData struct {
	ErrorMsg  string
	ShareURL  string
	SecretKey string
}

type server struct {
	config Config
	store  Store
}

type jsonOutput struct {
	Secret string `json:"secret"`
}

type Store interface {
	ReadSecret(key string) (string, error)
	SaveSecret(secret string) (string, error)
	ValidateSecret(key string) (bool, error)
	Close() error
}

// DOCME
func New(c Config, s Store) server {
	server := server{
		config: c,
		store:  s,
	}
	return server
}

// DOCME
func (serv server) Shutdown() error {
	return serv.store.Close()
}

// DOCME
func (serv server) Serve() error {
	// TODO: test all endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// this is because the "/" pattern of HandleFunc matches everything
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			serv.handleIndex(w, r)
		} else if r.Method == "POST" {
			serv.handlePost(w, r, serv.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	// #show_url
	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			serv.handleShow(w, r, serv.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			serv.handleFetchSecret(w, r, serv.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	portStr := strconv.Itoa(int(*serv.config.Port))
	if *serv.config.HTTPS {
		return http.ListenAndServeTLS(":"+portStr, *serv.config.Certfile, *serv.config.Keyfile, nil)
	} else {
		return http.ListenAndServe(":"+portStr, nil)
	}
}

// handleIndex serves the default page for creating a new secret
func (serv server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	serv.outputTpl(w, tplData{})
}

// handlePost stores the posted secret and outputs the generated key for reading it
func (serv server) handlePost(w http.ResponseWriter, r *http.Request, s Store) {
	honeypot := r.FormValue("signature")
	if len(honeypot) > 0 {
		// if the honeypot got filled, we will output a successful 200 response, so that the bots
		// don't think they have to try anything further
		fmt.Fprint(w, "ok")
		return
	}

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

	proto := "http"
	if *serv.config.HTTPS {
		proto = "https"
	}
	shareURL := proto + "://" + r.Host + "/show?key=" + key
	data := tplData{
		ShareURL: shareURL, // #show_url
	}
	serv.outputTpl(w, data)
}

// handleShow shows the button that displays the secret
func (serv server) handleShow(w http.ResponseWriter, r *http.Request, s Store) {
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "key not specified", http.StatusBadRequest)
		return
	}

	ok, err := s.ValidateSecret(key)
	if err != nil {
		log.Print(err)
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		data := tplData{
			ErrorMsg: "Could not find requested secret",
		}
		serv.outputTpl(w, data)
		return
	}

	data := tplData{
		SecretKey: key,
	}
	serv.outputTpl(w, data)
}

// handleFetchSecret outputs the content of the secret in JSON format
func (serv server) handleFetchSecret(w http.ResponseWriter, r *http.Request, s Store) {
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
func (serv server) outputTpl(w http.ResponseWriter, data tplData) {
	tpl := template.Must(template.New("").Parse(serv.indexHTML()))
	err := tpl.Execute(w, data)

	if err != nil {
		log.Print(err)
	}
}

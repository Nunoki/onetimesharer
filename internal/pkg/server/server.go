package server

import (
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
	store  Storer
}

type jsonOutput struct {
	Secret string `json:"secret"`
}

type Storer interface {
	ReadSecret(key string) (string, error)
	SaveSecret(secret string) (string, error)
	ValidateSecret(key string) (bool, error)
	Close() error
}

// DOCME
func New(c Config, s Storer) server {
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

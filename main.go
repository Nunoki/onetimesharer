package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
)

const filename = "secrets.json"

type tplData struct {
	shareURL  string
	secretKey string
	errorMsg  string
}

func main() {
	fileExists()
	server()
}

// fileExists makes sure the file with the secrets exists, by creating it if it doesn't already.
// If an error occurs with either reading or creating it, it outputs the error and exits the
// program.
func fileExists() {
	// TODO: test: https://pkg.go.dev/testing/fstest
	_, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		_, err = os.Create(filename)

		if err != nil {
			fmt.Printf("failed to create file: %s\n", err)
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Printf("could not read file: %s\n", err)
		os.Exit(1)
	}
}

// server starts listening on all the endpoints and passes the calls to the handlers
func server() {
	// TODO: test all endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			handleIndex(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}

// handleIndex serves the index.html page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	outputTpl(w, tplData{})
}

// outputTpl reads the index.html template from the file system, outputs it to the w writer, and
// passes the data to it
func outputTpl(w http.ResponseWriter, data tplData) {
	index, err := os.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Missing index page", http.StatusInternalServerError)
		return
	}
	tpl := template.Must(template.New("").
		Parse(string(index)))

	tpl.Execute(w, data)
}

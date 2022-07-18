package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const filename = "secrets.json"

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}

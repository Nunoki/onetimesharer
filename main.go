package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const filename = "secrets.json"
const port = "8000"

type tplData struct {
	ShareURL  string
	SecretKey string
	ErrorMsg  string
}

type collection map[string]string

func main() {
	verifyFile()
	serve()
}

// verifyFile makes sure the file with the secrets exists, by creating it if it doesn't already.
// If an error occurs with either reading or creating it, it outputs the error and exits the
// program.
func verifyFile() {
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

// serve starts listening on all the endpoints and passes the calls to the handlers
func serve() {
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
			handlePost(w, r)
			// TODO
			// } else if r.Method == "DELETE" {
			// 	handleDelete(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	// #show_url
	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handleShow(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleFetchSecret(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	log.Print("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleIndex serves the default page for creating a new secret
func handleIndex(w http.ResponseWriter, _ *http.Request) {
	outputTpl(w, tplData{})
}

// handlePost stores the posted secret and outputs the generated key for reading it
func handlePost(w http.ResponseWriter, r *http.Request) {
	secret := r.FormValue("secret")
	if secret == "" {
		http.Error(w, "failed to read posted content", http.StatusBadRequest)
		return
	}

	key, err := saveSecret(secret)
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
func handleShow(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "key not specified", http.StatusBadRequest)
		return
	}

	ok := validateSecret(key)
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
func handleFetchSecret(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "key not specified", http.StatusBadRequest)
		return
	}

	secret, err := readSecret(key)
	if err != nil {
		log.Print(err)
		http.Error(w, "failed to read secret", http.StatusInternalServerError)
		return
	}

	data := struct {
		Secret string `json:"secret"`
	}{
		Secret: secret,
	}
	output, _ := json.Marshal(data)
	w.Header().Set("Content-type", "application/json")
	w.Write(output)
}

// outputTpl parses the index.html file and outputs it to the w writer, passing the data to it
func outputTpl(w http.ResponseWriter, data tplData) {
	tpl := template.Must(template.ParseFiles("index.html"))
	err := tpl.Execute(w, data)

	if err != nil {
		log.Print(err)
	}
}

// DOCME
func saveSecret(secret string) (string, error) {
	// TODO encrypt
	key := randStr(40)
	secrets, err := readAllSecrets()
	if err != nil {
		return "", err
	}

	secrets[key] = string(secret)

	if err := storeSecrets(secrets); err != nil {
		return "", err
	}

	return key, nil
}

// DOCME
func storeSecrets(secrets collection) error {
	jsonData, err := json.Marshal(secrets)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filename, jsonData, os.FileMode(0111)); err != nil {
		return err
	}
	return nil
}

// DOCME
func deleteSecret(secrets collection, key string) error {
	delete(secrets, key)
	if err := storeSecrets(secrets); err != nil {
		return err
	}
	return nil
}

// DOCME
func readSecret(key string) (string, error) {
	secrets, err := readAllSecrets()
	if err != nil {
		return "", err
	}
	secret, ok := secrets[key]
	if !ok {
		return "", errors.New("not found")
	}
	if err := deleteSecret(secrets, key); err != nil {
		return "", err
	}
	return secret, nil
}

// DOCME
func validateSecret(key string) bool {
	secrets, err := readAllSecrets()
	if err != nil {
		log.Print(err)
		return false
	}
	_, ok := secrets[key]
	return ok
}

// DOCME
func readAllSecrets() (collection, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonData := collection{}
	err = json.Unmarshal(content, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// randStr returns a random string of length n
func randStr(n int) string {
	rand.Seed(time.Now().UnixMilli())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

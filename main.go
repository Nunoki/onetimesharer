package main

import (
	"fmt"
	"os"
)

const filename = "secrets.json"

func main() {
	fileExists()

	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("could not read file: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(f))
}

func fileExists() {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)

		if err != nil {
			fmt.Printf("failed to create file: %s\n", err)
			os.Exit(1)
		}
	}
}

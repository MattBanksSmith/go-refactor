package main

import (
	"go-refactor/internal"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		log.Println("provide a directory")
		return
	}

	// Specify the directory to scan for Go files
	dir := os.Args[1]

	// Recursively scan the directory for Go files
	err := internal.Do(dir)

	if err != nil {
		log.Fatal(err)
	}
}

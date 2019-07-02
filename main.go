package main

import (
	"log"
	"net/http"
	"os"
)

const defaultPort = "8000"

func main() {
	port := os.Getenv("PORT")

	listenPort := defaultPort
	if len(port) != 0 {
		listenPort = port
	}

	log.Printf("start listen on %v", port)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		log.Fatal(err)
	}

}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kmollee/url-short/controller"
)

const defaultPort = "8000"

func main() {
	port := os.Getenv("PORT")

	listenPort := defaultPort
	if len(port) != 0 {
		listenPort = port
	}

	r := controller.New()

	log.Printf("start listen on %v", port)
	if err := http.ListenAndServe(":"+listenPort, r); err != nil {
		log.Fatal(err)
	}

}

package controller

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func New() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", Hello)
	return r
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello")
}

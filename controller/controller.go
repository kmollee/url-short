package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kmollee/url-short/store"
)

type handler struct {
	svc store.Service
}

func New(svc store.Service) http.Handler {
	r := mux.NewRouter()

	h := &handler{svc: svc}

	r.HandleFunc("/", h.encode)
	r.HandleFunc("/redirect/{hash}", h.redirect)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return r
}

func (h *handler) encode(w http.ResponseWriter, r *http.Request) {

	log.Println("ENCODE!")

	switch r.Method {
	case http.MethodGet:
		log.Println("GET")
		fmt.Fprintf(w, `
		<!DOCTYPE html>
<html lang="en">
<head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="X-UA-Compatible" content="ie=edge">
        <title>Url-Short</title>
</head>
<body>
        <form action="/" method="post">
                URL:<input type="text" name="url">
                <input type="submit" value="Send">
        </form>
</body>
</html>	
		`)
		return

	case http.MethodPost:
		log.Println("POST")
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		url := strings.TrimSpace(r.PostFormValue("url"))
		if url == "" {
			return
		}

		if !strings.Contains(url, "http") {
			url = "http://" + url
		}

		hashCode, err := h.svc.Save(url)
		if err != nil {
			log.Printf("ERR: %v", err)
			http.Error(w, "could not save url", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "<html>hash url: <a href=\"/redirect/%s\">%s</a></html>", hashCode, hashCode)
		return
	default:
		http.Error(w, "method is not grant", http.StatusNotAcceptable)
	}
}

func (h *handler) redirect(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	url, err := h.svc.Load(vars["hash"])
	if err != nil {
		log.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusPermanentRedirect)
	return
}

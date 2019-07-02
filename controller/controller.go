package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lytics/base62"
)

func New() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", encode)
	r.HandleFunc("/{hash}", redirect)
	return r
}

func encode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintln(w, `
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

		hashUrl := base62.StdEncoding.EncodeToString([]byte(url))
		fmt.Fprintf(w, "<html>hash url: <a href=\"/%s\">%s</a></html>", hashUrl, hashUrl)
		return

	}
}

func redirect(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	b, err := base62.StdEncoding.DecodeString(vars["hash"])
	if err != nil {
		http.Error(w, "ERROR", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, string(b), http.StatusPermanentRedirect)
	return
}

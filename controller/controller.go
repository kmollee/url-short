package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"html/template"

	"github.com/gorilla/mux"
	"github.com/kmollee/url-short/store"
)

var temps = map[string]*template.Template{}

var funcMap = template.FuncMap{
	"safehtml": func(s string) template.HTML {
		log.Println(s)
		return template.HTML(s)
	},
}

func init() {
	temps["index"] = template.Must(template.New("layout.html").Funcs(funcMap).ParseFiles("template/layout.html", "template/index.html"))
	temps["info"] = template.Must(template.New("layout.html").Funcs(funcMap).ParseFiles("template/layout.html", "template/info.html"))
}

type handler struct {
	svc store.Service
}

func New(svc store.Service) http.Handler {
	r := mux.NewRouter()

	h := &handler{svc: svc}

	r.HandleFunc("/", h.encode)
	r.HandleFunc("/redirect/{hash}", h.redirect)
	r.HandleFunc("/info/{hash}", h.info)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return r
}

func (h *handler) encode(w http.ResponseWriter, r *http.Request) {
	t := temps["index"]
	switch r.Method {
	case http.MethodGet:
		err := t.Execute(w, nil)
		if err != nil {
			log.Println(err)
		}
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

		hashCode, err := h.svc.Save(url)
		if err != nil {
			log.Printf("ERR: %v", err)
			http.Error(w, "could not save url", http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, map[string]string{
			"shortURL":  fmt.Sprintf("http://%s/redirect/%s", r.Host, hashCode),
			"hash":      hashCode,
			"originURL": url,
		})
		if err != nil {
			log.Println(err)
		}
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

	http.Redirect(w, r, url, http.StatusMovedPermanently)
	return
}

func (h *handler) info(w http.ResponseWriter, r *http.Request) {
	t := temps["info"]
	vars := mux.Vars(r)

	item, err := h.svc.Info(vars["hash"])
	if err != nil {
		log.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"shortURL": fmt.Sprintf("http://%s/redirect/%s", r.Host, vars["hash"]),
		"item":     item,
	})
	if err != nil {
		log.Println(err)
	}
	return
}

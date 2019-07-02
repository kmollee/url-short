package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"html/template"

	"github.com/gorilla/mux"
	"github.com/kmollee/url-short/store"
	"github.com/pkg/errors"
)

var temps = map[string]*template.Template{}

type Status struct {
	Code int
	Err  error
	Msg  string
}

var StatusOK = Status{Code: http.StatusOK, Err: nil}

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

	r.HandleFunc("/", responseHandler(h.encode))
	r.HandleFunc("/redirect/{hash}", responseHandler(h.redirect))
	r.HandleFunc("/info/{hash}", responseHandler(h.info))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return r
}

func responseHandler(h func(w http.ResponseWriter, r *http.Request) Status) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		status := h(w, r)

		w.WriteHeader(status.Code)

		if status.Err != nil {
			log.Printf("ERR: %v", status.Err)
		}
		if status.Msg != "" {
			w.Write([]byte(status.Msg))
		}

	}
}

func (h *handler) encode(w http.ResponseWriter, r *http.Request) Status {
	t := temps["index"]
	switch r.Method {
	case http.MethodGet:
		err := t.Execute(w, nil)
		if err != nil {
			return Status{Code: http.StatusInternalServerError, Err: err}
		}
		return StatusOK
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			return Status{Code: http.StatusInternalServerError, Err: err}
		}

		url := strings.TrimSpace(r.PostFormValue("url"))
		if url == "" {
			return Status{Code: http.StatusNotAcceptable, Err: nil, Msg: "URL is empty"}
		}

		if !strings.Contains(url, "http") {
			url = "http://" + url
		}

		hashCode, err := h.svc.Save(url)
		if err != nil {
			return Status{Code: http.StatusInternalServerError, Err: errors.Wrap(err, "could not save url into store")}
		}

		err = t.Execute(w, map[string]string{
			"shortURL":  fmt.Sprintf("http://%s/redirect/%s", r.Host, hashCode),
			"hash":      hashCode,
			"originURL": url,
		})
		if err != nil {
			return Status{Code: http.StatusInternalServerError, Err: errors.Wrap(err, "could not render template")}
		}

		return StatusOK
	default:
		return Status{Code: http.StatusMethodNotAllowed, Err: fmt.Errorf("method is not allow: %s from %v", r.Method, r.RemoteAddr), Msg: "Request method is not allowed"}
	}
}

func (h *handler) redirect(w http.ResponseWriter, r *http.Request) Status {

	vars := mux.Vars(r)
	code := vars["hash"]

	if code == "" {
		return Status{Code: http.StatusNoContent, Err: fmt.Errorf("redirect hash code is empty"), Msg: "hash code is empty"}
	}
	url, err := h.svc.Load(code)

	if err == sql.ErrNoRows {
		return Status{Code: http.StatusNotFound, Err: nil, Msg: "no such short URL"}
	}

	if err != nil {
		return Status{Code: http.StatusInternalServerError, Err: fmt.Errorf("could not load item from store")}
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
	return StatusOK
}

func (h *handler) info(w http.ResponseWriter, r *http.Request) Status {
	t := temps["info"]
	vars := mux.Vars(r)

	item, err := h.svc.Info(vars["hash"])
	if err != nil {
		return Status{Code: http.StatusNotFound, Err: nil, Msg: "not such short URL"}
	}

	err = t.Execute(w, map[string]interface{}{
		"shortURL": fmt.Sprintf("http://%s/redirect/%s", r.Host, vars["hash"]),
		"item":     item,
	})
	if err != nil {
		return Status{Code: http.StatusInternalServerError, Err: errors.Wrap(err, "could not render template")}
	}
	return StatusOK
}

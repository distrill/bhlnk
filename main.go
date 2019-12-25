package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	db, err := NewDbConnection()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	handlers := map[string]http.Handler{
		http.MethodGet: NewRedirectHandler(db),
		http.MethodPut: NewPutRedirectHandler(db),
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", methodWrapper(handlers))
}

func methodWrapper(handlers map[string]http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := handlers[r.Method]; ok {
			handler.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}

func NewRedirectHandler(db *sql.DB) http.Handler {
	redirectHandler := MapHandler(make(map[string]string), fallbackMux())
	redirectHandler = DbHandler(db, redirectHandler)
	return redirectHandler
}

func NewPutRedirectHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		type body struct {
			Url string `json:"url"`
		}
		var b body
		err = json.Unmarshal(bs, &b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		path, err := GetPath(db, b.Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		protocol := "http://"
		if r.TLS != nil {
			protocol = "https://"
		}
		fmt.Fprint(w, protocol+r.Host+"/"+path)
	})
}

func fallbackMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to bhlnk!")
	})
	return mux
}

package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/distrill/blink"
	blinkdb "github.com/distrill/blink/db"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "blink"
	password = "blink"
	dbname   = "blink"
)

func main() {
	db, err := blinkdb.NewDbConnection()
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
	redirectHandler := blink.MapHandler(make(map[string]string), fallbackMux())
	redirectHandler = blink.DbHandler(db, redirectHandler)

	yamlfile := flag.String("yaml", "", "specify a yaml file containing redirects")
	jsonfile := flag.String("json", "", "specify a json file containing redirects")
	flag.Parse()

	if *yamlfile != "" {
		yaml, err := ioutil.ReadFile(*yamlfile)
		if err != nil {
			panic(err)
		}
		redirectHandler, err = blink.YAMLHandler(yaml, redirectHandler)
		if err != nil {
			panic(err)
		}
	}

	if *jsonfile != "" {
		json, err := ioutil.ReadFile(*jsonfile)
		if err != nil {
			panic(err)
		}
		redirectHandler, err = blink.JSONHandler(json, redirectHandler)
		if err != nil {
			panic(err)
		}
	}
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

		path, err := blinkdb.GetPath(db, b.Url)
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

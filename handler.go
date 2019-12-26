package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// NewDbHandler - get url from id derived from path. redirect or 404
func NewDbHandler(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		url, err := GetUrl(db, id)

		// something went wrong
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// found your url
		if url != "" {
			http.Redirect(w, r, url, http.StatusSeeOther)
			return
		}

		// url is not in our database
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}

// NewPutRedirectHandler - get path from url provided in req.body
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

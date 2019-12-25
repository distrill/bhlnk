package main

import (
	"database/sql"
	"net/http"
)

// MapHandler - base handler, looks up paths in map and if exist rewrite to map value (url)
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if url, ok := pathsToUrls[r.URL.Path]; ok {
			http.Redirect(w, r, url, http.StatusSeeOther)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}

// DbHandler - get url from id derived from path. redirect or 404
func DbHandler(db *sql.DB, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		url, err := GetUrl(db, id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else if url != "" {
			http.Redirect(w, r, url, http.StatusSeeOther)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}

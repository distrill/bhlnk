package blink

import (
	"database/sql"
	"encoding/json"
	"net/http"

	blinkdb "github.com/distrill/blink/db"
	"gopkg.in/yaml.v2"
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

type redirect struct {
	Path string `yaml:"path" json:"path"`
	Url  string `yaml:"url" json:"url"`
}

// YAMLHandler - given yaml file contents, parse into [path]url map and send through MapHandler
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	redirects := []redirect{}
	err := yaml.Unmarshal(yml, &redirects)
	if err != nil {
		return nil, err
	}

	pathsToUrls := make(map[string]string)

	for _, record := range redirects {
		pathsToUrls[record.Path] = record.Url
	}

	return MapHandler(pathsToUrls, fallback), nil
}

// JSONHandler - given json file contents, parse into [path]url map and send through MapHandler
func JSONHandler(j []byte, fallback http.Handler) (http.HandlerFunc, error) {
	redirects := []redirect{}
	err := json.Unmarshal(j, &redirects)
	if err != nil {
		return nil, err
	}

	pathsToUrls := make(map[string]string)
	for _, record := range redirects {
		pathsToUrls[record.Path] = record.Url
	}

	return MapHandler(pathsToUrls, fallback), nil
}

// DbHandler - get url from id derived from path. redirect or 404
func DbHandler(db *sql.DB, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		url, err := blinkdb.GetUrl(db, id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else if url != "" {
			http.Redirect(w, r, url, http.StatusSeeOther)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}

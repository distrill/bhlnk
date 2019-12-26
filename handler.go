package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// NewDbHandler - get url from id derived from path. redirect or 404
func NewDbHandler(db *sql.DB) http.HandlerFunc {
	logger := NewLogger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		logger.WithField("path", id).Info("Fetching url for path")
		url, err := GetUrl(db, id)

		// something went wrong
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err":  err,
				"path": id,
			}).Error("Error fetching url")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// found your url
		if url != "" {
			logger.WithFields(logrus.Fields{
				"path": id,
				"url":  url,
			}).Info("Redirecting")
			http.Redirect(w, r, url, http.StatusSeeOther)
			return
		}

		logger.WithFields(logrus.Fields{
			"path": id,
		}).Info("Url not Found")
		// url is not in our database
		http.Error(w, "this is not the hook.", http.StatusNotFound)
	})
}

// NewPutRedirectHandler - get path from url provided in req.body
func NewPutRedirectHandler(db *sql.DB) http.Handler {
	logger := NewLogger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err":  err,
				"body": r.Body,
			}).Error("Error reading body")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		type body struct {
			Url string `json:"url"`
		}
		var b body
		err = json.Unmarshal(bs, &b)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err":  err,
				"body": bs,
			}).Error("Error parsing body")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if b.Url == "" {
			logger.Info("Missing body.url")
			http.Error(w, "body.url is required", http.StatusUnprocessableEntity)
		}

		path, err := GetPath(db, b.Url)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err": err,
				"url": b.Url,
			}).Error("Error getting path from url")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		protocol := "http://"
		if r.TLS != nil {
			protocol = "https://"
		}

		logger.WithFields(logrus.Fields{
			"url":  b.Url,
			"path": path,
		}).Info("Gonna send it")
		fmt.Fprint(w, protocol+r.Host+"/"+path)
	})
}

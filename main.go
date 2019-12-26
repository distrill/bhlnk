package main

import (
	"github.com/rs/cors"
	"net/http"
)

func main() {
	logger := NewLogger()
	db, err := NewDbConnection()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	defer FlushLogs()

	handlers := map[string]http.Handler{
		http.MethodGet: NewDbHandler(db),
		http.MethodPut: NewPutRedirectHandler(db),
	}

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"PUT", "POST", "GET"},
	}).Handler(methodWrapper(handlers))

	logger.Info("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
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

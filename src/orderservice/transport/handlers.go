package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func Router() http.Handler {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.HandleFunc("/hello-world", helloWorld).Methods(http.MethodGet)
	s.HandleFunc("/kitty", getKitty).Methods(http.MethodGet)
	return logMiddleware(r)
}

func helloWorld(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "Hello World!")
	if err != nil {
		return
	}
}

type Kitty struct {
	Name string `json:"name"`
}

func getKitty(w http.ResponseWriter, r *http.Request) {
	cat := Kitty{"Кот"}
	b, _ := json.Marshal(cat)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, string(b))
	if err != nil {
		return
	}
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"method":     r.Method,
			"url":        r.URL,
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.UserAgent(),
		}).Info("got a new request")
		h.ServeHTTP(w, r)
	})
}

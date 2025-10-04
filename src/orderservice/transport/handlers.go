package transport

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.HandleFunc("/hello-world", helloWorld).Methods(http.MethodGet)

	return r
}

func helloWorld(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

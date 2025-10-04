package main

import (
	"net/http"
	"orderservice/transport"
)

func main() {
	router := transport.Router()
	err := http.ListenAndServe(":8000", router)
	if err != nil {
		return
	}
}

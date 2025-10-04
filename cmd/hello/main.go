package main

import (
	"context"
	"net/http"
	"orderservice/pkg/hello/transport"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	file, err := os.OpenFile("orderservice.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				return
			}
		}(file)
	}

	serverUrl := ":8000"
	log.WithFields(log.Fields{"url": serverUrl}).Info("Starting server")

	killSignalChan := getKillSignalChan()
	srv := startServer(serverUrl)

	waitForKillSignal(killSignalChan)
	err = srv.Shutdown(context.Background())
	if err != nil {
		return
	}
}

func startServer(serverUrl string) *http.Server {
	router := transport.Router()
	srv := &http.Server{Addr: serverUrl, Handler: router}
	go func() {
		log.Fatal(http.ListenAndServe(serverUrl, router))
	}()

	return srv
}

func getKillSignalChan() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	return signalChan
}

func waitForKillSignal(signalChan chan os.Signal) {
	killSignalChan := <-signalChan
	switch killSignalChan {
	case os.Interrupt:
		log.Info("Got SIGINT...")
	case syscall.SIGTERM:
		log.Info("Got SIGTERM...")
	}
}

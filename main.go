package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/branow/peer-chat/config"
	"github.com/branow/peer-chat/handlers"
)

func main() {
	slog.SetLogLoggerLevel(slog.Level(config.GetConfing().LogLevel()))

	if err := start(); err != nil {
		panic(err)
	}
}

func start() error {
	server := NewServer(config.GetConfing().Port())
	slog.Info("Server started", "addr", server.Addr)
	return server.ListenAndServe()
}

func NewServer(port uint) *http.Server {
	mux := &http.ServeMux{}
	handlers.HandleServeMux(mux)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(int(port)),
		Handler: mux,
	}
	return server
}

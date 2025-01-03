package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/branow/peer-chat/config"
	"github.com/branow/peer-chat/handlers"
)

func main() {
	logLevel := slog.Level(config.GetConfig().LogLevel())
	slog.SetLogLoggerLevel(logLevel)

	if err := start(); err != nil {
		slog.Error("Server startup failed:", "error", err)
		os.Exit(1)
	}
}

func start() error {
	server := NewServer(config.GetConfig().Port())
	slog.Info("Server started:", "addr", server.Addr)
	return server.ListenAndServe()
}

func NewServer(port int) *http.Server {
	mux := &http.ServeMux{}
	handlers.HandleServeMux(mux)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(int(port)),
		Handler: mux,
	}
	return server
}

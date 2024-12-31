package main

import (
	"flag"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/branow/peer-chat/handlers"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	port := flag.Uint("p", 8080, "Server port")
	flag.Parse()

	if err := start(*port); err != nil {
		panic(err)
	}
}

func start(port uint) error {
	server := NewServer(port)
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

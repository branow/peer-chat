package main

import (
	"log/slog"
	"net/http"

	"github.com/branow/peer-chat/handlers"
	"github.com/branow/peer-chat/model"
	"golang.org/x/net/websocket"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	if err := start(); err != nil {
		panic(err)
	}
}

func start() error {
	server := NewServer()
	return server.ListenAndServe()
}

func NewServer() *http.Server {
	mux := &http.ServeMux{}
	handlers.HandleServeMux(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return server
}

func HandleSignalingServer(mux *http.ServeMux) {
	peerConn := model.NewPeerConnection()
	mux.Handle("/ws/", websocket.Handler(func(c *websocket.Conn) {
		client := model.NewClient(c)
		peerConn.AddClient(client)
		client.Wait()
	}))
}

package main

import (
	"log/slog"
	"net/http"

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
	HandleSignalingServer(mux)
	HandleResourceServer(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return server
}

func HandleSignalingServer(mux *http.ServeMux) {
	peerConn := NewPeerConnection()
	mux.Handle("/ws", websocket.Handler(func(c *websocket.Conn) {
		client := NewClient(c)
		peerConn.AddClient(client)
		client.Wait()
	}))
}

func HandleResourceServer(mux *http.ServeMux) {
	mux.Handle("/", http.FileServer(http.Dir("./static")))
}

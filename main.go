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
		peerConn.AddConnection(c)
		// err := peerConn.AddConnection(c)
		// if err != nil {
		// 	_, err := c.Write([]byte(err.Error()))
		// 	if err != nil {
		// 		slog.Error("Write Close Error Messsage:", "error", err.Error())
		// 	}
		// 	c.Close()
		// }
		peerConn.Wait()
	}))
}

func HandleResourceServer(mux *http.ServeMux) {
	mux.Handle("/", http.FileServer(http.Dir("./static")))
}

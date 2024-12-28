package main

import (
	"errors"
	"log/slog"
	"sync"

	"golang.org/x/net/websocket"
)

var ErrClientIsClosed = errors.New("client is closed")

type Client struct {
	connection *websocket.Conn
	out        chan []byte
	in         chan []byte
	isClosed   bool
	onClose    func()
	wg         sync.WaitGroup
}

func NewClient(conn *websocket.Conn) *Client {
	client := &Client{
		connection: conn,
		out:        make(chan []byte),
		in:         make(chan []byte),
		onClose:    func() {},
	}
	client.start()
	return client
}

func (c *Client) Wait() {
	c.wg.Wait()
}

func (c *Client) Send(data []byte) error {
	if c.isClosed {
		return ErrClientIsClosed
	}
	c.out <- data
	return nil
}

func (c *Client) Receive() ([]byte, error) {
	if c.isClosed {
		return nil, ErrClientIsClosed
	}
	data := <-c.in
	return data, nil
}

func (c *Client) SetOnClose(onClose func()) {
	if onClose != nil {
		c.onClose = onClose
	}
}

func (c *Client) readMessages() {
	defer func() {
		close(c.in)
		c.close()
	}()
	bufSize := 4088
	buf := make([]byte, bufSize)
outer:
	for {
		data := []byte{}
		hasNext := true
		for hasNext {
			n, err := c.connection.Read(buf)
			if err != nil {
				slog.Error("Connection Read Error: ", "error", err)
				break outer
			}
			data = append(data, buf[:n]...)
			hasNext = n == bufSize
		}
		c.in <- data
	}
}

func (c *Client) writeMessages() {
	defer func() {
		close(c.out)
		c.close()
	}()
	for data := range c.out {
		_, err := c.connection.Write(data)
		if err != nil {
			slog.Error("Connection Write Error: ", "error", err)
			break
		}
	}
}

func (c *Client) start() {
	c.wg.Add(2)
	go func() {
		defer c.wg.Done()
		c.readMessages()
	}()
	go func() {
		defer c.wg.Done()
		c.writeMessages()
	}()
	slog.Debug("Client starts")
}

func (c *Client) close() {
	slog.Debug("Client: close")
	c.connection.Close()
	c.isClosed = true
	c.onClose()
}

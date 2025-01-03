package model

import (
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var ErrClientIsClosed = errors.New("client is closed")

// Client maintains a websocket connection and provides basic operations
// for reading and writing data.
type Client struct {
	id         int // Is used to identify client during debugging
	connection *websocket.Conn
	out        chan []byte
	in         chan []byte
	isClosed   int32 // Use atomic for thread-safety
	onClose    func()
	wg         sync.WaitGroup
}

func NewClient(conn *websocket.Conn) *Client {
	client := &Client{
		id:         rand.Intn(1e5),
		connection: conn,
		out:        make(chan []byte, 100),
		in:         make(chan []byte, 100),
		onClose:    func() {},
	}
	client.start()
	return client
}

func (c *Client) Id() int {
	return c.id
}

// Wait blocks until all read and write goroutines for the client
// have finished, indicating that the conenction is closed.
func (c *Client) Wait() {
	c.wg.Wait()
}

// Send pushes the given slice of bytes into the output channel.
// Blocks if the output channel buffer is full.
func (c *Client) Send(data []byte) error {
	if atomic.LoadInt32(&c.isClosed) == 1 {
		return ErrClientIsClosed
	}
	c.out <- data
	return nil
}

// Receive retrives the slive of bytes from the input channel.
// Blocks if there is no data available.
func (c *Client) Receive() ([]byte, error) {
	if atomic.LoadInt32(&c.isClosed) == 1 {
		return nil, ErrClientIsClosed
	}
	return <-c.in, nil
}

// SetOnClose assigns a callback function to be executed after
// the connection is closed.
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

	for {
		_, r, err := c.connection.NextReader()
		if err != nil {
			slog.Error("Client read:", "client-id", c.id, "error", err)
			break
		}

		data, err := io.ReadAll(r)
		if err != nil {
			slog.Error("Client read:", "client-id", c.id, "error", err)
			break
		}

		c.in <- data
	}
}

// writeMessages reads data from the output channel and writes it
// to the WebSocket connection. Terminates on write errors or channel closure.
func (c *Client) writeMessages() {
	defer func() {
		close(c.out)
		c.close()
	}()

	for data := range c.out {
		if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Error("Client write:", "client-id", c.id, "error", err)
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
	slog.Debug("Client started:", "client-id", c.id)
}

// close safely closes WebSocket connection, invokes the onClose callback,
// and marks the client as closed using an atomic operation.
func (c *Client) close() {
	if atomic.CompareAndSwapInt32(&c.isClosed, 0, 1) {
		_ = c.connection.Close()
		c.onClose()
		slog.Debug("Client closed:", "client-id", c.id)
	}
}

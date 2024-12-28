package main

import (
	"encoding/json"
	"errors"
	"log/slog"

	"golang.org/x/net/websocket"
)

var ErrTooManyClients = errors.New("too many clients")

var requestOfferMessage = struct {
	MessageType string `json:"type"`
}{
	MessageType: "request-offer",
}

type PeerConnection struct {
	sender   *Client
	receiver *Client
}

func NewPeerConnection() *PeerConnection {
	return &PeerConnection{}
}

func (c *PeerConnection) AddConnection(conn *websocket.Conn) error {
	client := NewClient(conn)
	return c.addClient(client)
}

func (c *PeerConnection) Wait() {
	if c.sender != nil {
		c.sender.Wait()
	}
	if c.receiver != nil {
		c.receiver.Wait()
	}
}

func (c *PeerConnection) addClient(client *Client) error {
	slog.Debug("PeerConnection: add client", "client", client)
	if c.sender == nil {
		c.sender = client
		client.SetOnClose(func() {
			c.sender = nil
			c.signal()
		})
	} else if c.receiver == nil {
		c.receiver = client
		client.SetOnClose(func() {
			c.receiver = nil
			c.signal()
		})
	} else {
		return ErrTooManyClients
	}
	c.signal()
	return nil
}

func (c *PeerConnection) signal() error {
	if c.sender == nil || c.receiver == nil {
		return nil
	}
	slog.Debug("PeerConnection: start signalling")

	message, err := json.Marshal(requestOfferMessage)
	if err != nil {
		return err
	}

	c.sender.Send(message)
	slog.Debug("PeerConnection: send offer-request")
	offer, err := c.sender.Receive()
	slog.Debug("PeerConnection: receive offer")
	if err != nil {
		return err
	}
	slog.Debug("PeerConnection: send offer")
	c.receiver.Send(offer)

	answer, err := c.receiver.Receive()
	slog.Debug("PeerConnection: receive answer")
	if err != nil {
		return err
	}
	slog.Debug("PeerConnection: send answer")
	c.sender.Send(answer)

	slog.Debug("PeerConnection: siganlling ended successfuly")
	return nil
}

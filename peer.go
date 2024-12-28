package main

import (
	"encoding/json"
	"errors"
	"log/slog"
)

var ErrTooManyClients = errors.New("too many clients")

type message struct {
	MessageType string `json:"type"`
	Data        string `json:"data"`
}

var (
	RequestOfferMessage = message{MessageType: "request-offer"}
	WaitForPeerMessage  = message{
		MessageType: "wait", Data: "Wait for peer",
	}
	WaitForRoomMessage = message{
		MessageType: "wait", Data: "Wait for room",
	}
)

type PeerConnection struct {
	clients  *ClientList
	sender   *Client
	receiver *Client
}

func NewPeerConnection() *PeerConnection {
	return &PeerConnection{
		clients: NewClientList(),
	}
}

func (c *PeerConnection) AddClient(client *Client) {
	slog.Debug("PeerConnection: add client", "client", client)

	c.clients.AddClient(client)

	client.SetOnClose(func() {
		c.clients.RemoveClient(client)
		if c.sender == client || c.receiver == client {
			if c.sender == client {
				c.sender = nil
			}
			if c.receiver == client {
				c.receiver = nil
			}
			if err := c.signal(); err != nil {
				slog.Error("Signaling On Client Close:", "error", err)
			}
		}
	})

	if c.sender == nil || c.receiver == nil {
		if err := c.signal(); err != nil {
			slog.Error("Signaling Error:", "error", err)

			message, err := json.Marshal(message{MessageType: "error", Data: err.Error()})
			raisePanic(err) // Remove in production

			_ = client.Send(message)
		}
	} else {
		message, err := json.Marshal(WaitForRoomMessage)
		raisePanic(err) // Remove in production

		if err := client.Send(message); err != nil {
			slog.Error("Adding Client Error:", "error", err)
		}
	}
}

func (c *PeerConnection) signal() error {
	firstClients := c.clients.FindFirst(2)
	c.sender = firstClients[0]
	c.receiver = firstClients[1]
	slog.Debug("", "clients", len(c.clients.clients))
	slog.Debug("", "sender", c.sender, "receiver", c.receiver)

	if c.sender == nil || c.receiver == nil {
		message, err := json.Marshal(WaitForPeerMessage)
		raisePanic(err) // Remove in production
		if c.sender != nil {
			c.sender.Send(message)
		}
		if c.receiver != nil {
			c.receiver.Send(message)
		}
		return nil
	}

	slog.Debug("PeerConnection: start signalling")

	message, err := json.Marshal(RequestOfferMessage)
	raisePanic(err) // Remove in production

	if err := c.sender.Send(message); err != nil {
		return err
	}
	slog.Debug("PeerConnection: send offer-request")

	offer, err := c.sender.Receive()
	if err != nil {
		return err
	}
	slog.Debug("PeerConnection: receive offer")

	if err := c.receiver.Send(offer); err != nil {
		return err
	}
	slog.Debug("PeerConnection: send offer")

	answer, err := c.receiver.Receive()
	if err != nil {
		return err
	}
	slog.Debug("PeerConnection: receive answer")

	if err := c.sender.Send(answer); err != nil {
		return err
	}
	slog.Debug("PeerConnection: send answer")

	slog.Debug("PeerConnection: siganlling ended successfuly")
	return nil
}

// Use this function only during development
func raisePanic(err error) {
	if err != nil {
		panic(err)
	}
}

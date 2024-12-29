package model

import (
	"encoding/json"
	"log/slog"
	"math/rand"
)

type message struct {
	MessageType string `json:"type"`
	Data        string `json:"data"`
}

// Messages are used to inform clients about the state of Peer Connection.
var (
	RequestOfferMessage = message{MessageType: "request-offer"}
	WaitForPeerMessage  = message{
		MessageType: "wait", Data: "Wait for peer",
	}
	WaitForRoomMessage = message{
		MessageType: "wait", Data: "Wait for room",
	}
)

// PeerConnection establishes and manages peer-to-peer connection between two clients.
type PeerConnection struct {
	id                int // Is used to identify PeerConnection during debugging.
	clients           *ClientList
	sender            *Client
	receiver          *Client
	onEmptyConnection func()
}

func NewPeerConnection() *PeerConnection {
	return &PeerConnection{
		id:                rand.Intn(1e6),
		clients:           NewClientList(),
		onEmptyConnection: func() {},
	}
}

func (c *PeerConnection) Id() int {
	return c.id
}

func (c *PeerConnection) SetOnEmptyConnection(onEmptyConnection func()) {
	if onEmptyConnection != nil {
		c.onEmptyConnection = onEmptyConnection
	}
}

// GetClients returns the number of clients connected to this
// peer connection including all which are waiting.
func (c *PeerConnection) GetClients() int {
	return c.clients.Size()
}

// AddClient adds a new client to the peer connection.
// If two clients are available, it starts the signaling process.
func (c *PeerConnection) AddClient(client *Client) {
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
				slog.Error("Signaling on client close", "peer-connection", c.Id(),
					"client", client.Id(), "error", err)
				return
			}
			if c.sender == nil && c.receiver == nil {
				c.onEmptyConnection()
			}
		}
	})
	c.clients.AddClient(client)
	slog.Debug("PeerConnection added client", "peer-coonnection", c.Id(),
		"client", client.Id())

	if c.sender == nil || c.receiver == nil {
		if err := c.signal(); err != nil {
			slog.Error("Signaling on client add", "peer-connection", c.Id(),
				"client", client.Id(), "error", err)

			message, _ := json.Marshal(message{MessageType: "error", Data: err.Error()})
			_ = client.Send(message)
		}
	} else {
		message, _ := json.Marshal(WaitForRoomMessage)
		if err := client.Send(message); err != nil {
			slog.Error("Sending client message", "peer-connection", c.Id(),
				"client", client.Id(), "error", err)
		}
	}
}

// signal manages the signaling process to establish a peer connection.
func (c *PeerConnection) signal() error {
	clients := c.clients.FindFirst(2)
	if len(clients) < 2 {
		message, _ := json.Marshal(WaitForPeerMessage)
		for _, client := range clients {
			_ = client.Send(message)
		}
		return nil
	}

	c.sender, c.receiver = clients[0], clients[1]
	slog.Debug("Starting signaling", "peer-connection", c.Id(),
		"sender", c.sender.Id(), "receiver", c.receiver.Id())

	message, _ := json.Marshal(RequestOfferMessage)
	if err := c.sender.Send(message); err != nil {
		return err
	}

	offer, err := c.sender.Receive()
	if err != nil {
		return err
	}

	if err := c.receiver.Send(offer); err != nil {
		return err
	}

	answer, err := c.receiver.Receive()
	if err != nil {
		return err
	}

	if err := c.sender.Send(answer); err != nil {
		return err
	}

	slog.Debug("Finished signaling", "peer-connection", c.Id(),
		"sender", c.sender.Id(), "receiver", c.receiver.Id())
	return nil
}

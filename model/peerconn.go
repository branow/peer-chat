package model

import (
	"log/slog"
	"math/rand"
)

// Messages are used to inform clients about the state of Peer Connection.
var (
	RequestOfferMessage = Message{MessageType: RequestOffer}
	WaitForPeerMessage  = Message{
		MessageType: Wait, Data: "Wait for peer",
	}
	WaitForRoomMessage = Message{
		MessageType: Wait, Data: "Wait for room",
	}
	ErrorMessage = Message{
		MessageType: Error,
	}
)

// PeerConnection establishes and manages peer-to-peer connection between two clients.
type PeerConnection struct {
	id                int // Is used to identify PeerConnection during debugging.
	clients           *ClientList
	sender            *Peer
	receiver          *Peer
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
	client.SetOnClose(func() { c.removeClient(client) })
	c.clients.AddClient(client)
	slog.Debug("PeerConnection added client:", "peer-coonnection", c.Id(),
		"client", client.Id())

	if c.sender != nil && c.receiver != nil {
		if err := NewPeer(client).SendMessage(WaitForRoomMessage); err != nil {
			slog.Error("Sending client message:", "peer-connection", c.Id(),
				"client", client.Id(), "error", err)
		}
		return
	}

	if err := c.signal(); err != nil {
		slog.Error("Signaling on client add:", "peer-connection", c.Id(),
			"client", client.Id(), "error", err)

		errorMessage := Message{MessageType: Error, Data: err.Error()}
		_ = NewPeer(client).SendMessage(errorMessage)
	}
}

// signal manages the signaling process to establish a peer connection.
func (c *PeerConnection) signal() error {
	clients := c.clients.FindFirst(2)
	if len(clients) < 2 {
		for _, client := range clients {
			_ = NewPeer(client).SendMessage(WaitForPeerMessage)
		}
		return nil
	}

	c.sender, c.receiver = NewPeer(clients[0]), NewPeer(clients[1])
	slog.Debug("Starting signaling:", "peer-connection", c.Id(),
		"sender", c.sender.Id(), "receiver", c.receiver.Id())

	if err := c.sender.SendMessage(RequestOfferMessage); err != nil {
		return err
	}

	offer, err := c.sender.ReceiveExpected(Offer)
	if err != nil {
		return err
	}

	if err := c.receiver.SendMessage(offer); err != nil {
		return err
	}

	answer, err := c.receiver.ReceiveExpected(Answer)
	if err != nil {
		return err
	}

	if err := c.sender.SendMessage(answer); err != nil {
		return err
	}

	slog.Debug("Finished signaling:", "peer-connection", c.Id(),
		"sender", c.sender.Id(), "receiver", c.receiver.Id())
	return nil
}

func (c *PeerConnection) removeClient(client *Client) {
	c.clients.RemoveClient(client)
	if c.isSender(client) || c.isReceiver(client) {
		if c.isSender(client) {
			c.sender = nil
		}
		if c.isReceiver(client) {
			c.receiver = nil
		}
		if err := c.signal(); err != nil {
			slog.Error("Signaling on client close", "peer-connection", c.Id(),
				"client", client.Id(), "error", err)
			c.onEmptyConnection()
			return
		}
	}
	if c.sender == nil && c.receiver == nil {
		c.onEmptyConnection()
	}
}

func (c *PeerConnection) isSender(client *Client) bool {
	return c.sender != nil && c.sender.Client == client
}

func (c *PeerConnection) isReceiver(client *Client) bool {
	return c.receiver != nil && c.receiver.Client == client
}

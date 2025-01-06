package model

import (
	"encoding/json"
)

type MessageType string

const (
	RequestOffer = "request-offer"
	Offer        = "offer"
	Answer       = "answer"
	Wait         = "wait"
	Error        = "error"
)

type Message struct {
	MessageType string `json:"type"`
	Data        string `json:"data"`
	SDP         string `json:"sdp"`
}

type Peer struct {
	*Client
}

func NewPeer(c *Client) *Peer {
	return &Peer{Client: c}
}

func (p *Peer) ReceiveExpected(messageType MessageType) (Message, error) {
	message, err := p.ReceiveMessage()
	if err != nil {
		return message, err
	}
	if message.MessageType != string(messageType) {
		return Message{}, UnexpectedMessageTypeError{
			ExpectedType:    messageType,
			ReceivedMessage: message,
		}
	}
	return message, nil
}

func (p *Peer) ReceiveMessage() (Message, error) {
	var message Message
	bytes, err := p.Receive()
	if err != nil {
		return message, err
	}
	if err := json.Unmarshal(bytes, &message); err != nil {
		return message, IllegalMessageError{ReceivedMessage: bytes, Cause: err}
	}
	return message, nil
}

func (p *Peer) SendMessage(message Message) error {
	bytes, _ := json.Marshal(message)
	return p.Send(bytes)
}

type UnexpectedMessageTypeError struct {
	ExpectedType    MessageType
	ReceivedMessage Message
}

func (e UnexpectedMessageTypeError) Error() string {
	return "receive unexpected message type"
}

type IllegalMessageError struct {
	ReceivedMessage []byte
	Cause           error
}

func (e IllegalMessageError) Error() string {
	return "receive illegal message"
}

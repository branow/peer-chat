package main

import (
	"sync"
)

type ClientList struct {
	clients map[*Client]uint8
	mutex   sync.RWMutex
	counter uint8
}

func NewClientList() *ClientList {
	return &ClientList{
		clients: map[*Client]uint8{},
	}
}

func (cl *ClientList) AddClient(client *Client) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	cl.clients[client] = cl.counter
	cl.counter++
}

func (cl *ClientList) RemoveClient(client *Client) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	if _, ok := cl.clients[client]; ok {
		_ = client.connection.Close() //Ignore error
		delete(cl.clients, client)
	}
}

func (cl *ClientList) FindFirst(n uint8) []*Client {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	ents := entries(make([]*entry, n))
	for k, v := range cl.clients {
		ent := entry{key: k, value: v}
		ents.setIfMin(&ent)
	}
	return ents.get()
}

type entry struct {
	key   *Client
	value uint8
}

type entries []*entry

func (ents entries) setIfMin(ent *entry) {
	for i, e := range ents {
		if e == nil {
			ents[i] = ent
			return
		}
	}

	minI := 0
	minVal := uint8(2<<7 - 1)
	for i, e := range ents {
		if e.value < minVal {
			minVal = e.value
			minI = i
		}
	}

	if ent.value < minVal {
		ents[minI] = ent
	}
}

func (ents entries) get() []*Client {
	clients := make([]*Client, len(ents))
	for i, e := range ents {
		if e != nil {
			clients[i] = e.key
		}
	}
	return clients
}

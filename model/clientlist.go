package model

import (
	"sort"
	"sync"
)

// ClientList maintains a thread-safe list of clients with unique counters
// for tracking and ordering.
type ClientList struct {
	clients map[*Client]uint8
	mutex   sync.RWMutex
	// Counter uses uint8 to limit the maximum number of client to 255,
	// assuming this is sufficient.
	counter uint8
}

func NewClientList() *ClientList {
	return &ClientList{
		clients: map[*Client]uint8{},
	}
}

func (cl *ClientList) Size() int {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	return len(cl.clients)
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
		// Ensure the connection is closed before removing the client,
		// even if it is likely already closed.
		_ = client.connection.Close()
		delete(cl.clients, client)
	}
}

// Retrives the first `n` clients based on their counter value (in ascending order).
// If the number of available clients is less than `n`, it returns all of them.
func (cl *ClientList) FindFirst(n int) []*Client {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	var sortedClients []*Client
	for client := range cl.clients {
		sortedClients = append(sortedClients, client)
	}

	sort.Slice(sortedClients, func(i, j int) bool {
		return cl.clients[sortedClients[i]] < cl.clients[sortedClients[j]]
	})

	if len(sortedClients) > n {
		sortedClients = sortedClients[:n]
	}
	return sortedClients
}

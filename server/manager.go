package server

import (
	"net"

	"github.com/cornelk/hashmap"
)

type Client struct {
	Address     *net.UDPAddr
	LastMessage int64
}

// Identiy: Client struct
var clients = hashmap.New[string, Client]()

// AddClient adds a client to the hashmap
func AddClient(id string, client Client) {
	clients.Insert(id, client)
}

// RemoveClient removes a client from the hashmap
func RemoveClient(id string) {
	clients.Del(id)
}

// ExistsClient checks if a client exists in the hashmap
func ExistsClient(id string) bool {
	_, exists := clients.Get(id)
	return exists
}

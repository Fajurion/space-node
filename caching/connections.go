package caching

import "github.com/dgraph-io/ristretto"

var connectionsCache *ristretto.Cache

func setupConnectionsCache() {

	var err error
	connectionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5, // expecting to store 10k connections
		MaxCost:     1e4, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,  // Some random number, check docs
	})

	if err != nil {
		panic(err)
	}

}

type ConnectedClient struct {

	//* Connection information
	Address string   // IP and port of the client
	Key     [32]byte // Encryption key for the connection to the client

	//* User information
	UserID   string // User ID
	Username string // Username
	Tag      string // Tag
	Session  string // Connected session

}

// StoreConnection stores a connection in the cache
func StoreConnection(client ConnectedClient) {
	connectionsCache.Set(client.Address, client, 1)
}

// GetConnection returns a connection from the cache
func GetConnection(address string) (ConnectedClient, bool) {
	client, found := connectionsCache.Get(address)
	if found {
		return client.(ConnectedClient), true
	}
	return ConnectedClient{}, false
}

// DeleteConnection deletes a connection from the cache
func DeleteConnection(address string) {
	connectionsCache.Del(address)
}

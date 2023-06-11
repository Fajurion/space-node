package caching

import "github.com/dgraph-io/ristretto"

// ! For setting please ALWAYS use cost 1
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

// StoreConnection stores a connection in the cache
func StoreConnection(client ConnectedClient) {
	connectionsCache.Set(client.Address, client.UserID, 1)
}

// ExistsConnection checks if a connection exists in the cache
func ExistsConnection(address string) bool {
	_, found := connectionsCache.Get(address)
	return found
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

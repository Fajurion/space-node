package caching

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

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

// StoreConnection stores a connection in the cache for the user ttl
func StoreConnection(client ConnectedClient) {
	connectionsCache.SetWithTTL(client.Address, client.UserID, 1, UserTTL)
}

// RefreshConnection refreshes a connection in the cache for the user ttl
func RefreshConnection(address string) bool {

	client, valid := connectionsCache.Get(address)
	if !valid {
		return false
	}

	connectionsCache.SetWithTTL(address, client, 1, UserTTL)
	return true
}

func LastConnectionRefresh(address string) time.Duration {
	time, found := connectionsCache.GetTTL(address)

	if found {
		return time
	}

	return -1
}

// ExistsConnection checks if a connection exists in the cache
func ExistsConnection(address string) bool {
	_, found := connectionsCache.Get(address)
	return found
}

// GetConnection returns a connection from the cache
func GetConnection(address string) (string, bool) {
	client, found := connectionsCache.Get(address)
	if found {
		return client.(string), true
	}
	return "", false
}

// DeleteConnection deletes a connection from the cache
func DeleteConnection(address string) {
	connectionsCache.Del(address)
}

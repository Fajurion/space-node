package caching

import (
	"crypto/cipher"
	"strings"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/util"
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

type Connection struct {
	ID          string       // ID of the account
	Address     string       // Address of the client
	Key         cipher.Block // Encryption key
	LastMessage int64        // Last message received from the client
}

// StoreConnection stores a connection in the cache for the user ttl
func StoreConnection(client ConnectedClient, clientAddress string) {

	ip := strings.Split(client.Address, ":")[0]
	connectionsCache.SetWithTTL(ip, Connection{
		ID:      client.ID,
		Address: clientAddress,
		Key:     client.Key,
	}, 1, UserTTL)
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
func GetConnection(address string) (Connection, bool) {
	client, found := connectionsCache.Get(address)
	if found {
		return client.(Connection), true
	}
	return Connection{}, false
}

// DeleteConnection deletes a connection from the cache
func DeleteConnection(address string) {

	if integration.Testing {
		util.Log.Println("[udp]", address, "was deleted")
	}

	connectionsCache.Del(address)
}

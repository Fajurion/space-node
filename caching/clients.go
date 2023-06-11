package caching

import "github.com/dgraph-io/ristretto"

// ! For setting please ALWAYS use cost 1
var usersCache *ristretto.Cache

func setupUsersCache() {

	var err error
	usersCache, err = ristretto.NewCache(&ristretto.Config{
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
	Address     string // IP and port of the client
	Key         []byte // Encryption key for the connection to the client
	LastMessage int64  // Last message received from the client

	//* User information
	UserID   string // User ID
	Username string // Username
	Tag      string // Tag
	Session  string // Connected session

}

// StoreUser stores a user in the cache
func StoreUser(client ConnectedClient) {
	usersCache.Set(client.UserID, client, 1)
}

// ExistsUser checks if a user exists in the cache
func ExistsUser(userID string) bool {
	_, found := usersCache.Get(userID)
	return found
}

// GetUser returns a user from the cache
func GetUser(userID string) (ConnectedClient, bool) {
	client, found := usersCache.Get(userID)
	if found {
		return client.(ConnectedClient), true
	}
	return ConnectedClient{}, false
}

// DeleteUser deletes a user from the cache
func DeleteUser(userID string) {
	usersCache.Del(userID)
}

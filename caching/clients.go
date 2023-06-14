package caching

import (
	"crypto/cipher"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

// ! For setting please ALWAYS use cost 1
var usersCache *ristretto.Cache

func setupUsersCache() {

	var err error
	usersCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5, // expecting to store 10k connections
		MaxCost:     1e4, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,  // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			client := item.Value.(ConnectedClient)

			util.Log.Println("[udp]", client.ID, "("+client.Username+"#"+client.Tag+")", "disconnected")
		},
	})

	if err != nil {
		panic(err)
	}

}

type ConnectedClient struct {

	//* Connection information
	Address    string       // IP and port of the client
	Key        cipher.Block // Encryption key for the connection to the client
	TargetType int          // Type of the target
	TargetID   string       // Target ID

	//* User information
	ID       string // User ID
	Username string // Username
	Tag      string // Tag

}

// StoreUser stores a user in the cache
func StoreUser(client ConnectedClient) {
	usersCache.SetWithTTL(client.ID, client, 1, UserTTL)
}

// RefreshUser refreshes a user in the cache
func RefreshUser(userID string) bool {

	client, valid := usersCache.Get(userID)
	if !valid {
		return false
	}

	// Refresh room if entered one
	con := client.(ConnectedClient)
	if con.TargetType == 1 {
		RefreshRoom(con.TargetID)
	}

	usersCache.SetWithTTL(userID, client, 1, UserTTL)
	return true
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

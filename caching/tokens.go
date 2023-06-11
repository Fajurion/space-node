package caching

import (
	"crypto/sha256"

	"github.com/dgraph-io/ristretto"
)

// ! For setting please ALWAYS use cost 1
var tokenCache *ristretto.Cache

func setupTokenCache() {

	var err error
	tokenCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5, // expecting to store 10k items
		MaxCost:     1e4, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,  // Some random number, check docs
	})

	if err != nil {
		panic(err)
	}

}

type Client struct {
	Token  string // Auth token
	Secret string // Auth secret

	UserID   string // User ID
	Username string // Username
	Tag      string // Tag
	Session  string // Connected session
}

// StoreToken stores a token in the cache
func StoreToken(client Client) {
	tokenCache.Set(client.UserID, client, 1)
}

// GetToken returns a token from the cache
func GetToken(account string) (Client, bool) {
	client, found := tokenCache.Get(account)
	if found {
		return client.(Client), true
	}
	return Client{}, false
}

func (client Client) ToConnected(Address string) ConnectedClient {

	// Generate encryption key using hash of token
	key := sha256.Sum256([]byte(client.Token))

	return ConnectedClient{
		Address:  Address,
		Key:      key[:],
		UserID:   client.UserID,
		Username: client.Username,
		Tag:      client.Tag,
		Session:  client.Session,
	}
}

func (client Client) GetKey() []byte {
	arr := sha256.Sum256([]byte(client.Token))
	return arr[:]
}

// DeleteToken deletes a token from the cache
func DeleteToken(account string) {
	tokenCache.Del(account)
}

// ExistsToken checks if a token exists in the cache
func ExistsToken(account string) bool {
	_, found := tokenCache.Get(account)
	return found
}

package caching

import (
	"crypto/aes"
	"crypto/sha256"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/util"
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
		OnEvict: func(item *ristretto.Item) {
			client := item.Value.(Client)

			util.Log.Println("[udp]", client.ID, "was deleted")
		},
	})

	if err != nil {
		panic(err)
	}

}

const TargetRoom = 0
const TargetWorld = 1 // When world protocol is implemented

type Client struct {
	Token  string // Auth token
	Secret string // Auth secret

	TargetType int    // Target type
	Target     string // Target id

	ID       string // User ID
	Username string // Username
	Tag      string // Tag
	Session  string // Connected session
}

// Time for which a token is valid
const TokenTTL = time.Hour * 1

// GenerateRoomToken generates a token for a room given a client
func GenerateRoomToken(client Client, Room string) (string, string) {

	client.Token = util.GenerateToken(200)
	client.Secret = util.GenerateToken(32)
	client.TargetType = TargetRoom
	client.Target = Room

	storeToken(client)
	return client.Token, client.Secret
}

// GenerateRoomTestToken generates a token for a room given a client
func GenerateRoomTestToken(client Client, Room string) (string, string) {

	if !integration.Testing {
		panic("Cannot generate test token when not testing")
	}

	client.Token = client.ID
	client.Secret = client.ID
	client.TargetType = TargetRoom
	client.Target = Room

	storeToken(client)
	return client.Token, client.Secret
}

// storeToken stores a token in the cache
func storeToken(client Client) {
	tokenCache.SetWithTTL(client.ID, client, 1, TokenTTL)
}

// GetToken returns a token from the cache
func GetToken(account string) (Client, bool) {
	client, found := tokenCache.Get(account)
	if found {
		return client.(Client), true
	}
	return Client{}, false
}

func (client Client) ToConnected(Address string) (ConnectedClient, bool) {

	// Generate encryption key using hash of token
	key := sha256.Sum256([]byte(client.Token))
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		util.Log.Println("[udp]", "Error creating cipher for", client.ID, err)
		return ConnectedClient{}, false
	}

	return ConnectedClient{
		Address:  Address,
		Key:      cipher,
		ID:       client.ID,
		Username: client.Username,
		Tag:      client.Tag,
	}, true
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

// RandomTestClient returns a random client for testing
func RandomTestClient() Client {
	id := util.GenerateToken(8)
	return Client{
		ID:       id,
		Username: "tester " + id,
		Tag:      "tag",
	}
}

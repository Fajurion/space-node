package caching

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"net"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

type Connection struct {
	ID             string
	Room           string
	ClientID       string
	CurrentSession string
	UDP            *net.UDPAddr
	Key            []byte
	Cipher         cipher.Block
}

func (c *Connection) KeyBase64() string {
	return base64.StdEncoding.EncodeToString(c.Key)
}

// ! Always use cost 1
var connectionsCache *ristretto.Cache // ConnectionID -> Connection
var clientIDCache *ristretto.Cache    // ClientID -> ConnectionID

func setupConnectionsCache() {

	var err error
	connectionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 10_000_000, // 1 million expected connections
		MaxCost:     1 << 30,    // 1 GB
		BufferItems: 64,
	})

	if err != nil {
		panic(err)
	}

	clientIDCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 10_000_000, // 1 million expected connections
		MaxCost:     1 << 30,    // 1 GB
		BufferItems: 64,

		OnEvict: func(item *ristretto.Item) {
			util.Log.Println("[cache] cached client id of connection", item.Value, "was deleted")
		},
	})

	if err != nil {
		panic(err)
	}
}

// packetHash = encrypted hash included in the packet by the client
// hash = computed hash of the packet
func VerifyUDP(clientId string, udp net.Addr, hash []byte, packetHash []byte) (Connection, bool) {

	// Get connection
	connectionId, valid := clientIDCache.Get(clientId)
	if !valid {
		return Connection{}, false
	}

	obj, valid := connectionsCache.Get(connectionId.(string))
	if !valid {
		return Connection{}, false
	}
	conn := obj.(Connection)

	// Verify hash
	cipher, err := aes.NewCipher(conn.Key)
	if err != nil {
		util.Log.Println("Error: Couldn't create cipher:", err)
		return Connection{}, false
	}

	bytes, err := base64.StdEncoding.DecodeString(string(packetHash))
	if err != nil {
		util.Log.Println("Error: Couldn't decode hash:", err)
		return Connection{}, false
	}

	decrypted, err := util.DecryptAES(cipher, bytes)
	if err != nil {
		util.Log.Println("Error: Couldn't decrypt hash:", err)
		return Connection{}, false
	}

	if !util.CompareHash(decrypted, hash) {
		util.Log.Println("Error: Hashes don't match")
		util.Log.Println("Expected:", decrypted)
		util.Log.Println("Got:", hash)
		return Connection{}, false
	}

	// Set UDP
	if conn.UDP == nil {
		udp, err := net.ResolveUDPAddr("udp", udp.String())
		if err != nil {
			util.Log.Println("Error: Couldn't resolve udp address:", err)
			return Connection{}, false
		}

		conn.UDP = udp
		valid := EnterUDP(conn.Room, conn.ID, clientId, udp, &conn.Key)
		if !valid {
			util.Log.Println("Error: Couldn't enter udp")
			return Connection{}, false
		}
		connectionsCache.Set(connectionId, conn, 1)
		connectionsCache.Wait()
		util.Log.Println("Success: UDP set")
	}
	return conn, true
}

func EmptyConnection(connId string, room string) Connection {

	// Generate encryption key
	key, err := util.GenerateKey()
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Store in cache
	clientId := util.GenerateToken(10)
	conn := Connection{
		ID:       connId,
		Room:     room,
		ClientID: clientId,
		UDP:      nil,
		Key:      key,
		Cipher:   block,
	}
	connectionsCache.Set(connId, conn, 1)
	clientIDCache.Set(clientId, connId, 1)

	return conn
}

func GetConnection(connId string) (Connection, bool) {
	conn, valid := connectionsCache.Get(connId)
	if !valid {
		return Connection{}, false
	}
	return conn.(Connection), valid
}

// TODO: Create a test for all deletion functions
func DeleteConnection(connId string) {
	obj, valid := connectionsCache.Get(connId)
	if !valid {
		return
	}
	connection := obj.(Connection)
	connectionsCache.Del(connId)
	clientIDCache.Del(connection.ClientID)
}

func JoinSession(connId string, sessionId string) bool {
	obj, valid := connectionsCache.Get(connId)
	if !valid {
		return false
	}
	connection := obj.(Connection)
	connection.CurrentSession = sessionId
	connectionsCache.Set(connId, connection, 1)
	connectionsCache.Wait()
	return true
}

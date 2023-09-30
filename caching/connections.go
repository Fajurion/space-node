package caching

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"net"
	"time"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

type Connection struct {
	ID       string
	Room     string
	ClientID string
	UDP      *net.UDPAddr
	Key      []byte
	Cipher   cipher.Block
}

func (c *Connection) KeyBase64() string {
	return base64.StdEncoding.EncodeToString(c.Key)
}

// ! Always use cost 1
var connectionsCache *ristretto.Cache // ClientID -> Connection

const connectionTTL = 5 * time.Minute
const connectionPacketTTL = 1 * time.Hour

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
}

// packetHash = encrypted hash included in the packet by the client
// hash = computed hash of the packet
func VerifyUDP(clientId string, udp net.Addr, hash []byte, packetHash []byte) (Connection, bool) {

	// Get connection
	conn, valid := GetConnection(clientId)
	if !valid {
		return Connection{}, false
	}

	// Verify hash
	decrypted, err := util.DecryptAES(conn.Cipher, packetHash)
	if err != nil {
		return Connection{}, false
	}
	if !util.CompareHash(decrypted, hash) {
		return Connection{}, false
	}

	// Set UDP
	if conn.UDP != nil {
		udp, err := net.ResolveUDPAddr("udp", udp.String())
		if err != nil {
			return Connection{}, false
		}

		conn.UDP = udp
		valid := EnterUDP(conn.Room, conn.ID, udp)
		if !valid {
			return Connection{}, false
		}
		connectionsCache.SetWithTTL(clientId, conn, 1, connectionPacketTTL)
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
	clientId := util.GenerateToken(20)
	conn := Connection{
		ID:       connId,
		Room:     room,
		ClientID: clientId,
		UDP:      nil,
		Key:      key,
		Cipher:   block,
	}
	connectionsCache.SetWithTTL(clientId, conn, 1, connectionTTL)

	return conn
}

func GetConnection(ip string) (Connection, bool) {
	conn, valid := connectionsCache.Get(ip)
	return conn.(Connection), valid
}

func DeleteConnection(ip string) {
	connectionsCache.Del(ip)
}

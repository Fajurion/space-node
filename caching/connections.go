package caching

import (
	"crypto/cipher"
	"encoding/base64"
	"net"
	"time"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

type Connection struct {
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
var connectionsCache *ristretto.Cache // IP -> Connection

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

func SetupUDP(connId string, addr string, udp *net.UDPAddr) bool {

	// Get connection
	ip := addr + ":" + connId
	conn, valid := GetConnection(ip)
	if !valid {
		return false
	}

	// Set UDP
	conn.UDP = udp
	connectionsCache.SetWithTTL(ip, conn, 1, connectionPacketTTL)

	return true
}

func EmptyConnection(connId string, room string, addr string) Connection {

	// Generate encryption key
	key, err := util.GenerateKey()
	if err != nil {
		panic(err)
	}

	// Store in cache
	conn := Connection{
		Room:     room,
		ClientID: util.GenerateToken(4),
		UDP:      nil,
		Key:      key,
		Cipher:   nil,
	}
	ip := addr + ":" + conn.ClientID
	connectionsCache.SetWithTTL(ip, conn, 1, connectionTTL)

	return conn
}

func GetConnection(ip string) (Connection, bool) {
	conn, valid := connectionsCache.Get(ip)
	return conn.(Connection), valid
}

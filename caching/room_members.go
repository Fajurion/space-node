package caching

import (
	"net"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

type RoomConnection struct {
	ID         string
	Connection *net.UDPAddr
	Adapter    string
}

// TODO: Store Room ID -> Connections
type RoomConnections struct {
	Connections []RoomConnection
}

// ! For setting please ALWAYS use cost 1
var roomConnectionsCache *ristretto.Cache

func setupRoomConnectionsCache() {

	var err error
	roomConnectionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // expecting to store 1k room connections
		MaxCost:     1 << 30, // maximum cost of cache is 1GB
		BufferItems: 64,      // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			util.Log.Println("[udp]", item.Key, "'s connections were deleted")
		},
	})

	if err != nil {
		panic(err)
	}

}

// Get all connections from a room
func GetAllConnections(room string) []*net.UDPAddr {

	connections, found := roomConnectionsCache.Get(room)

	if !found {
		return nil
	}

	return connections.([]*net.UDPAddr)
}

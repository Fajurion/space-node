package caching

import (
	"net"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

type RoomConnection struct {
	Connected  bool
	Connection *net.UDPAddr
	Adapter    string
	Data       string
}

// TODO: Store Room ID -> Connections
type RoomConnections map[string]RoomConnection

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

// JoinRoom adds a member to a room in the cache
func EnterUDP(roomID string, connectionId string, addr *net.UDPAddr) bool {

	room, valid := GetRoom(roomID)
	if !valid {
		return false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomID)
	if !valid {
		room.Mutex.Unlock()
		return false
	}

	obj, valid := roomConnectionsCache.Get(roomID)
	if !valid {
		room.Mutex.Unlock()
		return false
	}
	connections := obj.(RoomConnections)
	conn := connections[connectionId]
	if conn.Connected {
		room.Mutex.Unlock()
		return false
	}
	connections[connectionId] = RoomConnection{
		Connected:  true,
		Connection: addr,
		Data:       conn.Data,
		Adapter:    connectionId,
	}

	// Refresh room
	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)
	roomsCache.Wait()
	room.Mutex.Unlock()

	return true
}

// Sets the member data
func SetMemberData(roomID string, connectionId string, data string) bool {

	room, valid := GetRoom(roomID)
	if !valid {
		return false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomID)
	if !valid {
		room.Mutex.Unlock()
		return false
	}

	obj, valid := roomConnectionsCache.Get(roomID)
	if !valid {
		room.Mutex.Unlock()
		return false
	}
	connections := obj.(RoomConnections)
	if connections[connectionId].Connected {
		room.Mutex.Unlock()
		return false
	}
	connections[connectionId] = RoomConnection{
		Connected:  false,
		Connection: nil,
		Adapter:    connectionId,
		Data:       data,
	}

	// Refresh room
	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)
	roomsCache.Wait()
	room.Mutex.Unlock()

	return true
}

// Get all connections from a room
func GetAllConnections(room string) (RoomConnections, bool) {

	connections, found := roomConnectionsCache.Get(room)

	if !found {
		return nil, false
	}

	return connections.(RoomConnections), true
}

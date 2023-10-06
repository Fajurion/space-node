package caching

import (
	"sync"
	"time"

	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

// ! For setting please ALWAYS use cost 1
var roomsCache *ristretto.Cache

func setupRoomsCache() {

	var err error
	roomsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // expecting to store 1k rooms
		MaxCost:     1 << 30, // maximum cost of cache is 1GB
		BufferItems: 64,      // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			room := item.Value.(Room)

			util.Log.Println("[cache] room", room.ID, "was deleted")
		},
	})

	if err != nil {
		panic(err)
	}

}

type Room struct {
	Mutex *sync.Mutex
	ID    string // Account ID of the owner
	Data  string // Encrypted room data
	Start int64  // Timestamp of when the room was created
}

const RoomTTL = time.Minute * 5

// CreateRoom creates a room in the cache
func CreateRoom(roomId string, data string) {
	roomsCache.SetWithTTL(roomId, Room{&sync.Mutex{}, roomId, data, time.Now().UnixMilli()}, 1, RoomTTL)
	roomConnectionsCache.SetWithTTL(roomId, RoomConnections{}, 1, RoomTTL)
	roomsCache.Wait()
}

// JoinRoom adds a member to a room in the cache
func JoinRoom(roomID string, connectionId string) bool {

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
	connections[connectionId] = RoomConnection{
		Connected:  false,
		Connection: nil,
		Adapter:    connectionId,
		Data:       "",
	}

	// Refresh room
	roomConnectionsCache.SetWithTTL(roomID, connections, 1, RoomTTL)
	roomConnectionsCache.Wait()
	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)
	roomsCache.Wait()
	room.Mutex.Unlock()

	return true
}

// RefreshRoom refreshes a room in the cache
func RefreshRoom(roomID string) bool {

	room, valid := GetRoom(roomID)
	if !valid {
		return false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomID)
	if !valid {
		return false
	}

	connections, valid := GetAllConnections(roomID)
	if !valid {
		return false
	}

	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)
	roomConnectionsCache.SetWithTTL(roomID, connections, 1, RoomTTL)
	roomsCache.Wait()
	room.Mutex.Unlock()

	return true
}

// DeleteRoom deletes a room from the cache
func DeleteRoom(roomID string) {
	roomsCache.Del(roomID)
}

// GetRoom gets a room from the cache
func GetRoom(roomID string) (Room, bool) {
	object, valid := roomsCache.Get(roomID)
	if !valid {
		return Room{}, false
	}

	return object.(Room), true
}

func SetRoomData(roomID string, data string) bool {

	room, valid := GetRoom(roomID)
	if !valid {
		return false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomID)
	if !valid {
		return false
	}

	room.Data = data
	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)

	roomsCache.Wait()
	room.Mutex.Unlock()

	return true
}

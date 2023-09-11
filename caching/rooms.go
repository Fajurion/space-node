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

			util.Log.Println("[udp]", room.ID, "was deleted")
		},
	})

	if err != nil {
		panic(err)
	}

}

type Room struct {
	Mutex   *sync.Mutex
	ID      string   // ID acts as a sort of token for the room (anyone with the ID can join)
	Data    string   // Encrypted room data
	Start   int64    // Timestamp of when the room was created
	Members []string // Encrypted member IDs
}

const RoomTTL = time.Minute * 5

// CreateRoom creates a room in the cache
func CreateRoom(data string) string {
	roomId := util.GenerateToken(16)
	roomsCache.SetWithTTL(roomId, Room{&sync.Mutex{}, roomId, data, time.Now().UnixMilli(), []string{}}, 1, RoomTTL)
	roomsCache.Wait()
	return roomId
}

// JoinRoom adds a member to a room in the cache
func JoinRoom(roomID string, member string) bool {

	room, valid := GetRoom(roomID)
	if !valid {
		return false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomID)
	if !valid {
		return false
	}

	room.Members = append(room.Members, member)
	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)

	roomsCache.Wait()
	room.Mutex.Unlock()

	return true
}

// LeaveRoom removes a member from a room in the cache
func LeaveRoom(roomID string, member string) bool {

	room, valid := GetRoom(roomID)
	if !valid {
		return false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomID)
	if !valid {
		return false
	}

	for i, m := range room.Members {
		if m == member {
			room.Members = append(room.Members[:i], room.Members[i+1:]...)
			break
		}
	}

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

	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)

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

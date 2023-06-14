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
		NumCounters: 1e4, // expecting to store 1k rooms
		MaxCost:     1e3, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,  // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			room := item.Value.(Room)

			util.Log.Println("[udp]", room.ID, "was deleted")
		},
	})

	if err != nil {
		panic(err)
	}

}

var roomLocker map[string]*sync.Mutex = make(map[string]*sync.Mutex)

type Room struct {
	Mutex   *sync.Mutex
	ID      string
	Members []string
}

const RoomTTL = time.Minute * 5

// CreateRoom creates a room in the cache
func CreateRoom(roomID string) {
	roomsCache.SetWithTTL(roomID, Room{ID: roomID, Members: []string{}, Mutex: &sync.Mutex{}}, 1, RoomTTL)
	roomsCache.Wait()
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

	room, valid := roomsCache.Get(roomID)
	if !valid {
		return false
	}

	roomsCache.SetWithTTL(roomID, room, 1, RoomTTL)
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

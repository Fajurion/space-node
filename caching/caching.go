package caching

import (
	"time"

	"fajurion.com/voice-node/util"
)

func SetupMemory() {
	setupRoomsCache()
	setupRoomConnectionsCache()
}

func CloseCaches() {
	util.Log.Println("Closing caches...")
	roomsCache.Close()
	roomConnectionsCache.Close()
}

const UserTTL = 30 * time.Second

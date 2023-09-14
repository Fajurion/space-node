package caching

import (
	"time"

	"fajurion.com/voice-node/util"
)

func SetupMemory() {
	setupRoomsCache()
	setupRoomConnectionsCache()
	setupConnectionsCache()
}

func CloseCaches() {
	util.Log.Println("Closing caches...")
	roomsCache.Close()
	roomConnectionsCache.Close()
	connectionsCache.Close()
}

const UserTTL = 30 * time.Second

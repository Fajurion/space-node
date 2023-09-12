package caching

import (
	"time"

	"fajurion.com/voice-node/util"
)

func SetupMemory() {
	setupRoomsCache()
	setupRoomConnectionsCache()
	setupConnectionsCache()
	setupDeletionsCache()
}

func CloseCaches() {
	util.Log.Println("Closing caches...")
	roomsCache.Close()
	roomConnectionsCache.Close()
	connectionsCache.Close()
	deletionsCache.Close()
}

const UserTTL = 30 * time.Second

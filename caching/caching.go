package caching

import (
	"fajurion.com/voice-node/util"
)

func SetupMemory() {
	setupRoomsCache()
	setupRoomConnectionsCache()
	setupConnectionsCache()
	setupSessionsCache()
	setupTablesCache()
}

func CloseCaches() {
	util.Log.Println("Closing caches...")
	roomsCache.Close()
	roomConnectionsCache.Close()
	connectionsCache.Close()
	sessionsCache.Close()
	tablesCache.Close()
}

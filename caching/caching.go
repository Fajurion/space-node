package caching

import "fajurion.com/voice-node/util"

func SetupMemory() {
	setupTokenCache()
	setupConnectionsCache()
	setupUsersCache()
}

func CloseCaches() {
	util.Log.Println("Closing caches...")

	tokenCache.Close()
	connectionsCache.Close()
	usersCache.Close()
}

package caching

func SetupMemory() {
	setupTokenCache()
	setupConnectionsCache()
}

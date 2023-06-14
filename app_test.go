package main

import (
	"testing"

	"fajurion.com/voice-node/caching"
)

func TestConcurrency(t *testing.T) {

	// Setup memory
	caching.SetupMemory()
	caching.TestRooms(t)
}

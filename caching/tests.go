package caching

import (
	"testing"
	"time"

	"fajurion.com/voice-node/util"
)

// TestRooms tests the caching of rooms
func TestRooms(t *testing.T) {

	// Test caching
	CreateRoom("id", "test")
	for i := 0; i < 10; i++ {
		go func() {
			valid := JoinRoom("test", util.GenerateToken(5), "")
			if !valid {
				t.Error("Room not found")
			}
		}()
	}

	for i := 0; i < 5; i++ {
		go func() {
			valid := RefreshRoom("test")
			if !valid {
				t.Error("Room not found")
			}
		}()
	}

	time.Sleep(time.Millisecond * 500)
	_, valid := GetRoom("test")
	if !valid {
		t.Error("Room not found")
	}

	// TODO: Fix this

	/*
		if len(room.Members) != 10 {
			t.Errorf("Room has wrong number of members (expected 10, got %d)", len(room.Members))
		}
	*/
}

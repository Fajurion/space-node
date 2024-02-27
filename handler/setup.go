package handler

import (
	"context"
	"os"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/server"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipesfiber/wshandler"
	"github.com/livekit/protocol/livekit"
)

// Action: setup
func setup(message wshandler.Message) {

	if message.ValidateForm("data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}
	data := message.Data["data"].(string)

	// Generate new connection
	connection := caching.EmptyConnection(message.Client.ID, message.Client.Session)

	// Insert data
	if !caching.SetMemberData(message.Client.Session, message.Client.ID, connection.ClientID, data) {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	if !SendRoomData(message.Client.Session) {
		wshandler.ErrorResponse(message, integration.ErrorServer)
		return
	}

	// Check if livekit room already exists
	rooms, err := server.RoomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{
		Names: []string{message.Client.Session},
	})
	if err != nil {
		wshandler.ErrorResponse(message, integration.ErrorServer)
		return
	}

	if len(rooms.Rooms) > 0 {
		wshandler.NormalResponse(message, map[string]interface{}{
			"success": true,
			"id":      connection.ClientID,
			"key":     connection.KeyBase64(),
			"port":    util.UDPPort,
			"url":     os.Getenv("LK_URL"),
		})
		return
	}

	_, err = server.RoomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name:            message.Client.Session,
		EmptyTimeout:    120,
		MaxParticipants: 100,
	})
	if err != nil {
		wshandler.ErrorResponse(message, integration.ErrorServer)
		return
	}

	wshandler.NormalResponse(message, map[string]interface{}{
		"success": true,
		"id":      connection.ClientID,
		"key":     connection.KeyBase64(),
		"port":    util.UDPPort,
		"url":     os.Getenv("LK_URL"),
	})
}

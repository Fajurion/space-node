package routes

import (
	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"github.com/gofiber/fiber/v2"
)

type createRoomRequest struct {
	NodeToken string `json:"node_token"`
	Data      string `json:"data"`
}

// Route: /create
func createRoom(c *fiber.Ctx) error {

	// Parse request
	var req createRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return integration.InvalidRequest(c)
	}

	if req.NodeToken != integration.NODE_TOKEN {
		return integration.InvalidRequest(c)
	}

	// Create room
	roomID := caching.CreateRoom(req.Data)

	return c.JSON(fiber.Map{
		"success": true,
		"id":      roomID,
	})
}

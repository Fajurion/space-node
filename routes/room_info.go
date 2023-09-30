package routes

import (
	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"github.com/gofiber/fiber/v2"
)

type roomInfoRequest struct {
	Room string `json:"room"`
}

// Route: /info
func roomInfo(c *fiber.Ctx) error {

	// Parse request
	var req roomInfoRequest
	if c.BodyParser(&req) != nil {
		return integration.InvalidRequest(c)
	}

	room, validRoom := caching.GetRoom(req.Room)
	members, valid := caching.GetAllConnections(req.Room)
	if !valid || !validRoom {
		return integration.InvalidRequest(c)
	}

	returnableMembers := make([]string, len(members))
	i := 0
	for _, member := range members {
		returnableMembers[i] = member.Data
		i++
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    room.Data,
		"start":   room.Start,
		"members": returnableMembers,
	})
}

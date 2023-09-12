package routes

import (
	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipesfiber"
	"github.com/gofiber/fiber/v2"
)

type LeaveRoomRequest struct {
	Connection string `json:"connection"`
}

func leaveRoom(c *fiber.Ctx) error {

	var req LeaveRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return integration.InvalidRequest(c)
	}

	connections := pipesfiber.GetSessions(req.Connection)
	if len(connections) == 0 {
		return integration.InvalidRequest(c)
	}

	_, valid := caching.GetIP(connections[0])
	if !valid {
		return integration.InvalidRequest(c)
	}

	return nil
}

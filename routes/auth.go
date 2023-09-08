package routes

import (
	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipesfiber"
	"github.com/gofiber/fiber/v2"
)

type intializeRequest struct {
	Account   string `json:"account"` // Account ID
	Session   string `json:"session"` // Room ID in this app
	NodeToken string `json:"node_token"`
	End       int64  `json:"end"`
}

func initializeConnection(c *fiber.Ctx) error {

	// Parse the request
	var req intializeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if integration.NODE_TOKEN != req.NodeToken {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create token
	token, secret := caching.GenerateRoomToken(caching.Client{
		ID: req.Account,
	}, req.Session)

	tk := token + "." + secret
	pipesfiber.AddToken(tk, pipesfiber.ConnectionToken{
		UserID:  req.Account,
		Session: req.Session, // Again, this would be the room ID
		Data:    nil,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"load":    0, // TODO: Add real load in the future
		"token":   tk,
	})
}

package routes

import (
	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
	"github.com/gofiber/fiber/v2"
)

func initalize(c *fiber.Ctx) error {
	return util.FailedRequest(c, "not.implemented", nil)
}

type createTokenRequest struct {

	// Node data
	Token      string `json:"token"`
	TargetType int    `json:"targetType"`
	Target     string `json:"target"`

	// Account data
	Account  string `json:"account"`
	Username string `json:"username"`
	Tag      string `json:"tag"`
}

func createToken(c *fiber.Ctx) error {

	// Parse request
	var req createTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return util.InvalidRequest(c)
	}

	if req.Token != integration.NODE_TOKEN {
		return util.InvalidRequest(c)
	}

	if req.TargetType != caching.TargetRoom {
		return util.InvalidRequest(c)
	}

	// Create token
	token := caching.GenerateRoomToken(caching.Client{
		ID:       req.Account,
		Username: req.Username,
		Tag:      req.Tag,
	}, req.Target)

	return c.JSON(fiber.Map{
		"success": true,
		"token":   token,
	})
}

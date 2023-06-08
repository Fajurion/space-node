package routes

import (
	"log"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/util"
	"github.com/gofiber/fiber/v2"
)

type intializeRequest struct {
	NodeToken  string   `json:"node_token"`
	Session    string   `json:"session"`
	UserID     string   `json:"user_id"`
	Username   string   `json:"username"`
	Tag        string   `json:"tag"`
	SessionIds []string `json:"session_ids"`
}

func initalize(c *fiber.Ctx) error {

	// Parse the request
	var req intializeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if integration.NODE_TOKEN != req.NodeToken {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	log.Println(req.UserID, "|", req.SessionIds)

	tk := util.GenerateToken(200)

	/*
		// Check if there are too many users
		if bridge.GetConnections(req.UserID) >= 3 {
			return util.FailedRequest(c, "too.many.connections", nil)
		}

		bridge.AddToken(tk, req.UserID, req.Session, req.Username, req.Tag) */

	return c.JSON(fiber.Map{
		"success": true,
		"load":    0,
		"token":   tk,
	})
}

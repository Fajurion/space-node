package routes

import (
	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
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

	if caching.ExistsUser(req.UserID) || caching.ExistsToken(req.UserID) {
		return util.FailedRequest(c, "already.connected", nil)
	}

	util.Log.Println("Request by backend: Generating auth token for", req.UserID, "("+req.Username+")")
	tk := util.GenerateToken(200)    // Token is used to encrypt the connection (AES 256 hashed)
	secret := util.GenerateToken(32) // Secret is used to authenticate the user

	// Generate new token
	var client = caching.Client{
		Token:  tk,
		Secret: secret,

		UserID:   req.UserID,
		Username: req.Username,
		Tag:      req.Tag,
		Session:  req.Session,
	}
	caching.StoreToken(client)

	return c.JSON(fiber.Map{
		"success": true,
		"load":    0,
		"token":   tk + "." + secret,
	})
}

package routes

import (
	integration "fajurion.com/node-integration"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/receive"
	"github.com/gofiber/fiber/v2"
)

type socketlessRq struct {
	This    string        `json:"this"`
	Token   string        `json:"token"`
	Message pipes.Message `json:"message"`
}

func socketlessEvent(c *fiber.Ctx) error {

	// Parse request
	var rq socketlessRq
	if err := c.BodyParser(&rq); err != nil {
		return integration.InvalidRequest(c)
	}

	// Check token
	if rq.Token != pipes.CurrentNode.Token {
		return integration.InvalidRequest(c)
	}

	receive.HandleMessage("ws", rq.Message)
	return nil
}

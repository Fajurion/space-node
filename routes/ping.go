package routes

import (
	integration "fajurion.com/node-integration"
	"github.com/gofiber/fiber/v2"
)

func ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"gateway": integration.NODE_ID,
		"app":     "fj.proximity-voice",
	})
}

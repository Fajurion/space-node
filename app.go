package main

import (
	"github.com/Fajurion/pipes"
	"github.com/gofiber/fiber/v2"

	integration "fajurion.com/node-integration"
)

func main() {

	app := fiber.New()

	if integration.Setup() {
		pipes.SetupCurrent(integration.NODE_ID, integration.NODE_TOKEN)
		return
	}

	app.Listen(":3000")

}

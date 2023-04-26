package main

import (
	"github.com/Fajurion/pipes"
	"github.com/gofiber/fiber/v2"

	integration "fajurion.com/node-integration"
)

func main() {

	app := fiber.New()

	pipes.SetupCurrent("app", "app")

	if integration.Setup() {
		return
	}

	app.Listen(":3000")

}

package main

import (
	"github.com/Fajurion/pipes"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	pipes.SetupCurrent("app", "app")

	app.Listen(":3000")

}

package routes

import "github.com/gofiber/fiber/v2"

func SetupRoutes(router fiber.Router) {
	router.Post("/socketless", socketlessEvent)
	router.Post("/ping", ping)
}

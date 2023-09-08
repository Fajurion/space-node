package routes

import (
	"log"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber"
	pipesfroutes "github.com/Fajurion/pipesfiber/routes"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {
	router.Post("/socketless", socketlessEvent)
	router.Post("/ping", ping)

	// Auth
	router.Post("/auth/initalize", initializeConnection)

	setupPipesFiber(router)
}

func setupPipesFiber(router fiber.Router) {
	pipesfiber.Setup(pipesfiber.Config{
		ExpectedConnections: 10_0_0_0,       // 10 thousand, but funny
		SessionDuration:     time.Hour * 24, // This is kinda important

		// Report nodes as offline
		NodeDisconnectHandler: func(node pipes.Node) {
			integration.ReportOffline(node)
		},

		// Handle client disconnect
		ClientDisconnectHandler: func(client *pipesfiber.Client) {
			if integration.Testing {
				log.Println("Client disconnected:", client.ID)
			}

			util.PostRequest("/node/disconnect", map[string]interface{}{
				"node":    integration.NODE_ID,
				"token":   integration.NODE_TOKEN,
				"session": client.Session,
			})
		},

		// Handle enter network
		ClientConnectHandler: func(client *pipesfiber.Client) bool {
			if integration.Testing {
				log.Println("Client connected:", client.ID)
			}
			return false
		},

		// Handle client entering network
		ClientEnterNetworkHandler: func(client *pipesfiber.Client) bool {
			return false
		},
	})
	router.Route("/", pipesfroutes.SetupRoutes)
}

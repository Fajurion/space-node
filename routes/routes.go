package routes

import (
	"log"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
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
		},

		// Handle enter network
		ClientConnectHandler: func(client *pipesfiber.Client) bool {

			if integration.Testing {
				log.Println("Client connected:", client.ID)
			}

			// Generate new connection
			connection := caching.EmptyConnection(client.Session, client.Conn.RemoteAddr().String())

			client.SendEvent(pipes.Event{
				Name:   "udp",
				Sender: send.SenderSystem,
				Data: map[string]interface{}{
					"id":  connection.ClientID,
					"key": connection.KeyBase64(),
				},
			})

			return false
		},

		// Handle client entering network
		ClientEnterNetworkHandler: func(client *pipesfiber.Client) bool {
			return false
		},
	})
	router.Route("/", pipesfroutes.SetupRoutes)
}

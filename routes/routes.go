package routes

import (
	"log"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/handler"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/adapter"
	"github.com/Fajurion/pipes/send"
	"github.com/Fajurion/pipesfiber"
	pipesfroutes "github.com/Fajurion/pipesfiber/routes"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {
	router.Post("/socketless", socketlessEvent)
	router.Post("/ping", ping)

	// Auth
	router.Post("/auth/initialize", initializeConnection)
	router.Post("/leave", leaveRoom)
	router.Post("/info", roomInfo)

	setupPipesFiber(router)
}

func setupPipesFiber(router fiber.Router) {
	adapter.SetupCaching()
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

			// Remove from room
			caching.RemoveMember(client.Session, client.ID)
			caching.DeleteConnection(client.ID)

			// Send leave event
			handler.SendRoomData(client.Session)
		},

		// Handle enter network
		ClientConnectHandler: func(client *pipesfiber.Client) bool {

			if integration.Testing {
				log.Println("Client connected:", client.ID)
			}

			// Send room info
			room, validRoom := caching.GetRoom(client.Session)
			members, valid := caching.GetAllConnections(client.Session)
			if !valid || !validRoom {
				return false
			}

			returnableMembers := make([]string, len(members))
			i := 0
			for _, member := range members {
				returnableMembers[i] = member.Data
				i++
			}

			client.SendEvent(pipes.Event{
				Name:   "room_info",
				Sender: send.SenderSystem,
				Data: map[string]interface{}{
					"start":   room.Start,
					"room":    room.Data,
					"members": returnableMembers,
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

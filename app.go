package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/connection"
	"github.com/gofiber/fiber/v2"

	integration "fajurion.com/node-integration" // Propietary package (might be replaced with an open-source alternative in the future)
	"fajurion.com/voice-node/routes"
	"fajurion.com/voice-node/server"
)

var APP_ID uint = 0

func main() {

	app := fiber.New()
	app.Route("/", routes.SetupRoutes)

	if !integration.Setup() {
		return
	}

	pipes.SetupCurrent(integration.NODE_ID, integration.NODE_TOKEN)
	log.Println("[voice-node] Starting..")

	// Query current node
	_, _, currentApp, domain := integration.GetCurrent()
	APP_ID = currentApp

	log.Printf("[voice-node] Node %s on app %d\n", pipes.CurrentNode.ID, APP_ID)

	// Report online status
	res := integration.SetOnline()
	parseNodes(res)

	pipes.SetupSocketless(domain + "/socketless")

	// Check if test mode or production
	args := strings.Split(domain, ":")
	port, err := strconv.Atoi(args[1])
	if err != nil {
		log.Println("Error: Couldn't parse port of current node")
		return
	}

	// Wait for servers to start before adoption
	go func() {
		time.Sleep(time.Second * 2)
		pipes.IterateNodes(func(_ string, node pipes.Node) bool {

			log.Println("[voice-node] Connecting to node " + node.ID)

			if err := connection.ConnectUDP(node); err != nil {
				log.Println("[voice-node] Error: Couldn't connect to node " + node.ID)
				return false
			}

			return true
		})
	}()

	pipes.SetupUDP(fmt.Sprintf("%s:%d", args[0], port+1))
	pipes.DebugLogs = true

	if integration.Testing {

		// Start on localhost
		go server.Listen("localhost", port+1)
		app.Listen(fmt.Sprintf("localhost:%d", port))
	} else {

		// Start on all interfaces
		go server.Listen("0.0.0.0", port+1)
		app.Listen(fmt.Sprintf("0.0.0.0:%d", port))
	}
}

// Shared function between all nodes
func parseNodes(res map[string]interface{}) bool {

	if res["nodes"] == nil {
		return true
	}

	nodeList := res["nodes"].([]interface{})

	for _, node := range nodeList {
		n := node.(map[string]interface{})

		// Extract port and domain
		args := strings.Split(n["domain"].(string), ":")
		domain := args[0]
		port, err := strconv.Atoi(args[1])
		if err != nil {
			log.Println("Error: Couldn't parse port of node " + n["id"].(string))
			return true
		}

		pipes.AddNode(pipes.Node{
			ID:    fmt.Sprintf("%d", int(n["id"].(float64))),
			Token: n["token"].(string),
			SL:    fmt.Sprintf("http://%s:%d/sl", domain, port),
			UDP:   fmt.Sprintf("%s:%d", domain, port+1),
		})
	}

	return false
}

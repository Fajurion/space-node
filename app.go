package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Fajurion/pipes"
	"github.com/gofiber/fiber/v2"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/server"
)

var APP_ID uint = 0
var nodeID uint = 0

func main() {

	app := fiber.New()

	if !integration.Setup() {
		return
	}

	pipes.SetupCurrent(integration.NODE_ID, integration.NODE_TOKEN)
	log.Println("[voice-node] Starting..")

	nID, _ := strconv.Atoi(integration.NODE_ID)
	nodeID = uint(nID)

	// Query current node
	_, _, currentApp, domain := integration.GetCurrent()
	APP_ID = currentApp

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

	pipes.SetupUDP(fmt.Sprintf("%s:%d", args[0], port+1))

	if integration.Testing {

		// Start on localhost
		app.Listen(fmt.Sprintf("localhost:%d", port))
		server.Listen("localhost", port+1)
	} else {

		// Start on all interfaces
		app.Listen(fmt.Sprintf("0.0.0.0:%d", port))
		server.Listen("0.0.0.0", port+1)
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
			ID:    fmt.Sprintf("%f", n["id"].(float64)),
			Token: n["token"].(string),
			SL:    fmt.Sprintf("%s:%d", domain, port),
			UDP:   fmt.Sprintf("%s:%d", domain, port+1),
		})
	}

	return false
}

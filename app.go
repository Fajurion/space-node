package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/connection"
	"github.com/gofiber/fiber/v2"

	integration "fajurion.com/node-integration" // Propietary package (might be replaced with an open-source alternative in the future)
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/routes"
	"fajurion.com/voice-node/server"
	"fajurion.com/voice-node/util"
)

var APP_ID uint = 0

func main() {

	// Setup memory
	caching.SetupMemory()

	app := fiber.New()
	app.Route("/", routes.SetupRoutes)

	if !integration.Setup() {
		return
	}

	pipes.SetupCurrent(fmt.Sprintf("%d", integration.NODE_ID), integration.NODE_TOKEN)
	util.Log.Println("Starting..")

	// Query current node
	_, _, currentApp, domain := integration.GetCurrent()
	APP_ID = currentApp

	util.Log.Printf("Node %s on app %d\n", pipes.CurrentNode.ID, APP_ID)

	// Report online status
	integration.SetOnline()

	pipes.SetupSocketless("http://" + domain + "/socketless")

	// Connect to other nodes
	pipes.IterateNodes(func(_ string, node pipes.Node) bool {

		log.Println("Connecting to node " + node.WS)

		if err := connection.ConnectWS(node); err != nil {
			log.Println(err.Error())
		}
		return true
	})

	// Check if test mode or production
	args := strings.Split(domain, ":")
	port, err := strconv.Atoi(args[1])
	if err != nil {
		util.Log.Println("Error: Couldn't parse port of current node")
		return
	}

	// Test encryption
	first := testEncryption()
	second := testEncryption()

	if reflect.DeepEqual(first, second) {
		util.Log.Println("Error: Encryption is not working properly")
		return
	}

	util.Log.Println("Encryption is working properly!")

	pipes.DebugLogs = true
	server.SetupChannels()

	// Create testing room
	if integration.Testing {
		caching.CreateRoom("test")

		amount, _ := strconv.Atoi(os.Getenv("TESTING_AMOUNT"))
		for i := 0; i < amount; i++ {
			client := caching.RandomTestClient()
			tk, secret := caching.GenerateRoomTestToken(client, "test")
			util.Log.Println("TEST CLIENT", client.ID, "---------------------")
			util.Log.Println("Token:", tk)
			util.Log.Println("Secret:", secret)
		}
	}

	// Close caches on exit
	defer caching.CloseCaches()

	// Start on localhost
	go server.Listen(os.Getenv("LISTEN"), port+1)
	app.Listen(fmt.Sprintf("%s:%d", os.Getenv("LISTEN"), port))
}

// This function is used to test if the encryption is working properly and always different
func testEncryption() []byte {

	encrypted, err := connection.Encrypt(pipes.CurrentNode.ID, []byte("H"))
	if err != nil {
		util.Log.Println("Error: Couldn't encrypt message")
		return nil
	}

	util.Log.Println("Encrypted message: " + base64.StdEncoding.EncodeToString(encrypted))

	decrypted, err := connection.Decrypt(pipes.CurrentNode.ID, encrypted)
	if err != nil {
		util.Log.Println("Error: Couldn't decrypt message")
		return nil
	}

	util.Log.Println("Decrypted message: " + string(decrypted))

	return encrypted
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

		// Add node to pipes
		pipes.AddNode(pipes.Node{
			ID:    fmt.Sprintf("%d", int64(n["id"].(float64))),
			Token: n["token"].(string),
			WS:    "ws://" + fmt.Sprintf("%s:%d", domain, port) + "/adoption",
		})
	}

	return false
}

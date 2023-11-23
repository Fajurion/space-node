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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	integration "fajurion.com/node-integration" // Propietary package (might be replaced with an open-source alternative in the future)
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/caching/games/launcher"
	"fajurion.com/voice-node/handler"
	"fajurion.com/voice-node/routes"
	"fajurion.com/voice-node/server"
	"fajurion.com/voice-node/util"
)

var APP_ID uint = 0

func main() {

	// Setup memory
	caching.SetupMemory()

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	app.Route("/", routes.SetupRoutes)

	if !integration.Setup() {
		return
	}

	launcher.InitGames()
	pipes.SetupCurrent(fmt.Sprintf("%d", integration.NODE_ID), integration.NODE_TOKEN)
	util.Log.Println("Starting..")

	// Query current node
	_, _, currentApp, domain := integration.GetCurrent()
	APP_ID = currentApp

	util.Log.Printf("Node %s on app %d\n", pipes.CurrentNode.ID, APP_ID)

	pipes.SetupSocketless("http://" + domain + "/socketless")
	pipes.SetupWS("ws://" + domain + "/adoption")
	handler.Initialize()

	// Report online status
	res := integration.SetOnline()
	parseNodes(res)

	// Check if test mode or production
	args := strings.Split(domain, ":")
	var err error
	util.Port, err = strconv.Atoi(args[1])
	if err != nil {
		util.Log.Println("Error: Couldn't parse port of current node")
		return
	}
	util.UDPPort = util.Port + 1

	// Test encryption
	first := testEncryption()
	second := testEncryption()

	if reflect.DeepEqual(first, second) {
		util.Log.Println("Error: Encryption is not working properly")
		return
	}

	util.Log.Println("Encryption is working properly!")

	pipes.DebugLogs = true

	// Create testing room
	if integration.Testing {
		caching.CreateRoom("id", "test")

		amount, err := strconv.Atoi(os.Getenv("TESTING_AMOUNT"))
		if err != nil {
			util.Log.Println("Error: Couldn't parse testing amount")
			return
		}

		for i := 0; i < amount; i++ {
			clientId := util.GenerateToken(5)
			connection := caching.EmptyConnection(clientId, "id")
			valid := caching.JoinRoom("id", connection.ClientID)
			if !valid {
				util.Log.Println("Error: Couldn't join room")
				return
			}
			util.Log.Println("--- TESTING CLIENT ---")
			util.Log.Println(connection.ClientID + ":" + connection.KeyBase64())
			util.Log.Println("----------------------")
		}
	}

	// Close caches on exit
	defer caching.CloseCaches()

	// Connect to other nodes
	pipes.IterateNodes(func(_ string, node pipes.Node) bool {

		log.Println("Connecting to node " + node.WS)

		if err := connection.ConnectWS(node); err != nil {
			log.Println(err.Error())
		}
		return true
	})

	// Start on localhost
	go server.Listen(os.Getenv("LISTEN"), util.UDPPort)
	app.Listen(fmt.Sprintf("%s:%d", os.Getenv("LISTEN"), util.Port))
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

func generateBase64Key() string {
	bytes, _ := util.GenerateKey()
	return base64.StdEncoding.EncodeToString(bytes)
}

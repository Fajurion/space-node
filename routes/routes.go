package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/handler"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/adapter"
	"github.com/Fajurion/pipesfiber"
	pipesfroutes "github.com/Fajurion/pipesfiber/routes"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {
	router.Post("/socketless", socketlessEvent)
	router.Post("/ping", ping)

	router.Post("/pub", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"pub": integration.PackageRSAPublicKey(integration.NodePublicKey),
		})
	})

	// These are publicly accessible yk (cause this can be public information cause encryption and stuff)
	router.Post("/leave", leaveRoom)
	router.Post("/info", roomInfo)

	setupPipesFiber(router)
}

func setupPipesFiber(router fiber.Router) {
	adapter.SetupCaching()
	log.Println("JWT Secret:", integration.JwtSecret)
	pipesfiber.Setup(pipesfiber.Config{
		Secret:              []byte(integration.JwtSecret),
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

		// Validate token and create room
		TokenValidateHandler: func(claims *pipesfiber.ConnectionTokenClaims, attachments string) bool {

			// Create room (if needed)
			_, valid := caching.GetRoom(claims.Session)
			if !valid {
				util.Log.Println("Creating new room for", claims.Account, "("+claims.Session+")")
				caching.CreateRoom(claims.Session, "")
			} else {
				util.Log.Println("Room already exists for", claims.Account, "("+claims.Session+")")
			}

			return false
		},

		// Handle enter network
		ClientConnectHandler: func(client *pipesfiber.Client, key string) bool {

			// Get the AES key from attachments
			aesKeyEncrypted, err := base64.StdEncoding.DecodeString(key)
			if err != nil {
				return true
			}

			// Decrypt AES key
			aesKey, err := integration.DecryptRSA(integration.NodePrivateKey, aesKeyEncrypted)
			if err != nil {
				return true
			}

			// Just for debug purposes
			log.Println(base64.StdEncoding.EncodeToString(aesKey))

			// Set AES key in client data
			client.Data = ExtraClientData{aesKey}
			pipesfiber.UpdateClient(client)

			if integration.Testing {
				log.Println("Client connected:", client.ID)
			}
			return false
		},

		// Handle client entering network
		ClientEnterNetworkHandler: func(client *pipesfiber.Client, key string) bool {
			return false
		},

		// Set default encoding middleware
		DecodingMiddleware:       EncryptionDecodingMiddleware,
		ClientEncodingMiddleware: EncryptionClientEncodingMiddleware,
	})
	router.Route("/", pipesfroutes.SetupRoutes)
}

// Extra client data attached to the pipes-fiber client
type ExtraClientData struct {
	Key []byte // AES encryption key
}

// Middleware for pipes-fiber to add encryption support
func EncryptionDecodingMiddleware(client *pipesfiber.Client, bytes []byte) (pipesfiber.Message, error) {

	log.Println("DECRYPTING")

	// Decrypt the message using AES
	key := client.Data.(ExtraClientData).Key
	log.Println(len(bytes))
	messageEncoded, err := integration.DecryptAES(key, bytes)
	if err != nil {
		return pipesfiber.Message{}, err
	}

	// Unmarshal the message using sonic
	var message pipesfiber.Message
	err = sonic.Unmarshal(messageEncoded, &message)
	if err != nil {
		return pipesfiber.Message{}, err
	}

	log.Println("DECRYPTED")

	return message, nil
}

// Middleware for pipes-fiber to add encryption support (in encoding)
func EncryptionClientEncodingMiddleware(client *pipesfiber.Client, message []byte) ([]byte, error) {

	// Handle potential errors (with casting in particular)
	defer func() {
		if err := recover(); err != nil {
			pipesfiber.ReportClientError(client, "encryption failure (probably casting)", errors.ErrUnsupported)
		}
	}()

	// Check if the encryption key is set
	if client.Data == nil {
		return nil, errors.New("no encryption key set")
	}

	// Encrypt the message using the client encryption key
	key := client.Data.(ExtraClientData).Key
	log.Println("ENCODING KEY: "+base64.StdEncoding.EncodeToString(key), client.ID, string(message))
	result, err := integration.EncryptAES(key, message)
	hash := sha256.Sum256(result)
	log.Println("hash: " + base64.StdEncoding.EncodeToString(hash[:]))
	return result, err
}

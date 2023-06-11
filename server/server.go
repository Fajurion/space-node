package server

import (
	"crypto/aes"
	"encoding/base64"
	"net"
	"strings"
	"time"

	integration "fajurion.com/node-integration"
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes/connection"
	"github.com/Fajurion/pipes/receive"
	"github.com/bytedance/sonic"

	pipesUtil "github.com/Fajurion/pipes/util"
)

const PrefixNode = 'n'
const PrefixClient = 'c'
const PrefixTesting = 't' // Only enabled in testing mode

func Listen(domain string, port int) {

	connection.GeneralPrefix = []byte{PrefixNode, ':'}

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(domain),
	}

	util.Log.Println("Starting UDP server..")

	// Start udp server
	udpServ, err := net.ListenUDP("udp", &addr)
	if err != nil {
		util.Log.Println("[udp] Error: ", err)
		panic(err)
	}
	defer udpServ.Close()

	buffer := make([]byte, 8*1024) // 8 kb buffer

	util.Log.Println("UDP server started")

	for {
		offset, clientAddr, err := udpServ.ReadFrom(buffer) // Use client addr to rate limit in the future
		if err != nil {
			util.Log.Println("[udp] Error: ", err)
			continue
		}

		// Extract message
		msg := buffer[:offset]
		util.Log.Println("Client:", clientAddr.String(), "| Message:", string(msg))

		// Check if client wants to send to node
		exists := caching.ExistsConnection(clientAddr.String())
		node := msg[0] == PrefixNode

		if exists && node {
			caching.DeleteConnection(clientAddr.String())
			continue
		}

		if node {
			err := receive.ReceiveUDP(msg[2:])
			if err != nil {
				util.Log.Println("[udp] Error receiving node message: ", err)
			}

			continue
		}

		if msg[0] != PrefixClient {
			caching.DeleteConnection(clientAddr.String())
			continue
		}

		// Register client
		if exists {

			// Update last message
			client, valid := caching.GetConnection(clientAddr.String())
			if !valid {
				util.Log.Println("[udp] Error getting client, even though exists: ", err)
				continue
			}

			// Check if sent too quickly
			if time.Now().UnixMilli()-client.LastMessage < 50 {
				// TODO: Block
				continue
			}

			util.Log.Println("[udp]", string(msg[2:]), offset)

			// Echo (for now)
			_, err = udpServ.WriteTo(msg[2:], clientAddr)
			if err != nil {
				util.Log.Println("[udp] Error echoing message: ", err)
				continue
			}

			client.LastMessage = time.Now().UnixMilli()
			continue
		}

		// Register client (AUTH)
		stringMsg := string(msg[2:])
		args := strings.Split(stringMsg, ":")

		//! Check for testing stuff
		if msg[0] == PrefixTesting && integration.Testing {

			if len(args) < 3 {
				util.Log.Println("[udp] Error: Invalid testing packet")
				continue
			}

			// Read auth packet ((c:)account:(base64 encoded + AES encrypted secret):(json for account details))
			var accountDetails map[string]interface{}
			err := sonic.UnmarshalString(args[2], &accountDetails)
			if err != nil {
				util.Log.Println("[udp] Error: Invalid testing packet")
				continue
			}

			if args[0] != accountDetails["id"].(string) {
				util.Log.Println("[udp] Error: Invalid account id")
				continue
			}

			// Read account data packet
			var testClient = caching.Client{
				Token:  integration.TestingToken,
				Secret: integration.TestingToken,

				UserID:   accountDetails["id"].(string),
				Username: accountDetails["username"].(string),
				Tag:      accountDetails["tag"].(string),
				Session:  accountDetails["session"].(string),
			}

			// Read auth packet ((c:)account:(base64 encoded + AES encrypted secret):(base64 json packet))
			if !auth(args[1], testClient, clientAddr.String()) {
				continue
			}

			util.Log.Println("[udp] ^Testing client registered")
			continue
		}

		//* Register client (AUTH)

		msg = msg[2:]
		if len(msg) < 1 {
			continue
		}

		// Read auth packet ((c:)account:(base64 encoded + AES encrypted secret))
		if len(args) < 2 {
			util.Log.Println("[udp] Error: Invalid auth packet")
			continue
		}

		// Get auth token
		accountId := args[0]
		secret := args[1]

		client, valid := caching.GetToken(accountId)
		if !valid {
			util.Log.Println("[udp] Error: Invalid auth token")
			continue
		}

		if !auth(secret, client, clientAddr.String()) {
			continue
		}
	}
}

func auth(secret string, client caching.Client, address string) bool {

	cipher, err := aes.NewCipher(client.GetKey())
	if err != nil {
		util.Log.Println("[udp] Error creating cipher: ", err)
		return false
	}

	var decrypted, decoded []byte
	decoded, err = base64.StdEncoding.DecodeString(secret)
	if err != nil {
		util.Log.Println("[udp] Error decoding message: ", err)
		return false
	}

	decrypted, err = pipesUtil.DecryptAES(cipher, decoded)
	if err != nil {
		util.Log.Println("[udp] Error decrypting message: ", err)
		return false
	}

	decryptedMsg := string(decrypted)
	if decryptedMsg != client.Secret {
		util.Log.Println("[udp] Error: Invalid secret")
		return false
	}

	// Add client
	caching.DeleteToken(client.UserID)
	connectedClient := client.ToConnected(address)

	caching.StoreConnection(connectedClient)
	caching.StoreUser(connectedClient)
	util.Log.Println("[udp]", connectedClient.UserID+"("+connectedClient.Username+"#"+connectedClient.Tag+") connected")

	return true
}

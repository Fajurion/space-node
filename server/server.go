package server

import (
	"crypto/aes"
	"encoding/base64"
	"net"
	"strings"
	"time"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes/connection"
	"github.com/Fajurion/pipes/receive"

	pipesUtil "github.com/Fajurion/pipes/util"
)

const PrefixNode = 'n'
const PrefixClient = 'c'

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

		//* Register client (AUTH)
		msg = msg[2:]
		if len(msg) < 1 {
			continue
		}

		// Read auth packet ((c:)account:(base64 encoded + AES encrypted secret))
		stringMsg := string(msg)
		args := strings.Split(stringMsg, ":")

		if len(args) < 2 {
			util.Log.Println("[udp] Error: Invalid auth packet")
			continue
		}

		accountId := args[0]
		secret := args[1]

		// Get auth token
		client, valid := caching.GetToken(accountId)
		if !valid {
			util.Log.Println("[udp] Error: Invalid auth token")
			continue
		}

		cipher, err := aes.NewCipher(client.GetKey())
		if err != nil {
			util.Log.Println("[udp] Error creating cipher: ", err)
			continue
		}

		var decrypted, decoded []byte
		decoded, err = base64.StdEncoding.DecodeString(secret)
		if err != nil {
			util.Log.Println("[udp] Error decoding message: ", err)
			continue
		}

		decrypted, err = pipesUtil.DecryptAES(cipher, decoded)
		if err != nil {
			util.Log.Println("[udp] Error decrypting message: ", err)
			continue
		}

		decryptedMsg := string(decrypted)
		if decryptedMsg != client.Secret {
			util.Log.Println("[udp] Error: Invalid secret")
			continue
		}

		// Add client
		caching.DeleteToken(accountId)
		connectedClient := client.ToConnected(clientAddr.String())

		caching.StoreConnection(connectedClient)
		caching.StoreUser(connectedClient)
		util.Log.Println("[udp]", connectedClient.UserID+"("+connectedClient.Username+"#"+connectedClient.Tag+") connected")

	}

}

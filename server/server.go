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
	"github.com/Fajurion/pipes/adapter"

	pipesUtil "github.com/Fajurion/pipes/util"
)

const PrefixClient = 'c'    // For client auth
const PrefixEncrypted = 'e' // For encrypted traffic with the client

var udpServ *net.UDPConn

func Listen(domain string, port int) {

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(domain),
	}

	util.Log.Println("Starting UDP server..")

	// Start udp server
	var err error
	udpServ, err = net.ListenUDP("udp", &addr)
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

		//* protocol standard: PREFIX+CLIENT_ID:CHANNEL:MESSAGE
		// Prefix: 1 byte
		// Client ID: 4 bytes
		// Channel: 1 byte
		// Message: rest of the packet
		msg := buffer[:offset]

		if len(msg) < 9 {
			util.Log.Println("[udp] Error: Invalid message")
			continue
		}

		// Check if client wants to send to node
		endIndex := 6
		clientID := string(msg[1 : endIndex-1])
		ip := strings.Split(clientAddr.String(), ":")[0] + clientID // Get client ip + client id
		connection, exists := caching.GetConnection(ip)

		if integration.Testing {
			util.Log.Println("[udp] Message from:", ip, "with prefix: ", string(msg[0]), "found: ", exists)
		}

		// Register client
		if exists {

			if msg[0] != PrefixEncrypted {
				util.Log.Println("[udp] Error: No encryption prefix")
				caching.DeleteConnection(ip)
				continue
			}

			decrypted, err := pipesUtil.DecryptAES(connection.Key, msg[endIndex:])
			if err != nil {
				util.Log.Println("[udp] Error: Invalid message")
				continue
			}

			// Check if sent too quickly
			if time.Now().UnixMilli()-connection.LastMessage < 50 {
				// TODO: Block (deletion for now)
				caching.DeleteConnection(ip)
				continue
			}

			// Handle channels
			err = ExecuteChannel(connection.ID, clientID, decrypted, clientAddr)
			if err != nil {
				util.Log.Println("[udp]", connection.ID+": Error:", err)
			}

			connection.LastMessage = time.Now().UnixMilli()
			continue
		}

		if msg[0] != PrefixClient {

			if integration.Testing {
				util.Log.Println("[udp] Error: Invalid prefix")
			}

			caching.DeleteConnection(ip)
			continue
		}

		//* Register client (AUTH)
		stringMsg := string(msg[endIndex:])
		util.Log.Println("[udp] Auth packet: ", stringMsg)
		args := strings.Split(stringMsg, ":")

		msg = msg[2:]
		if len(msg) < 1 {
			continue
		}

		// Read auth packet ((c:)account:room:(base64 encoded + AES encrypted secret))
		if len(args) < 3 {
			util.Log.Println("[udp] Error: Invalid auth packet")
			continue
		}

		// Get auth token
		accountId := args[0]
		room := args[1]
		secret := args[2]

		client, valid := caching.GetToken(accountId)
		if !valid {
			util.Log.Println("[udp] Error: Invalid auth token (not found)")
			continue
		}

		if client.Account != accountId {
			util.Log.Println("[udp] Error: Invalid auth token (invalid account id)")
			caching.DeleteToken(accountId)
			continue
		}

		// Join room if correct
		connectedClient, authenticated := auth(secret, client, ip, clientAddr.String(), clientID)
		if !authenticated {
			caching.DeleteConnection(ip)
			caching.DeleteUser(accountId)
			adapter.RemoveUDP(accountId)
			continue
		}

		SendConfirmation(clientAddr.String(), client.Account, &connectedClient.Key)
		caching.JoinRoom(room, accountId)
	}
}

func auth(secret string, client caching.Client, address string, clientAddress string, clientID string) (caching.ConnectedClient, bool) {

	cipher, err := aes.NewCipher(client.GetKey())
	if err != nil {
		util.Log.Println("[udp] Error creating cipher: ", err)
		return caching.ConnectedClient{}, false
	}

	var decrypted, decoded []byte
	decoded, err = base64.StdEncoding.DecodeString(secret)
	if err != nil {
		util.Log.Println("[udp] Error decoding message: ", err)
		return caching.ConnectedClient{}, false
	}

	decrypted, err = pipesUtil.DecryptAES(cipher, decoded)
	if err != nil {
		util.Log.Println("[udp] Error decrypting message: ", err)
		return caching.ConnectedClient{}, false
	}

	decryptedMsg := string(decrypted)
	if decryptedMsg != client.Secret {
		util.Log.Println("[udp] Error: Invalid secret")
		return caching.ConnectedClient{}, false
	}

	// Add client
	caching.DeleteToken(client.Account)
	connectedClient, valid := client.ToConnected(clientAddress, clientID)
	if !valid {
		return caching.ConnectedClient{}, false
	}

	caching.StoreConnection(connectedClient, clientAddress)
	caching.StoreUser(connectedClient)
	util.Log.Println("[udp]", connectedClient.Account, "connected")

	err = AddAdapter(connectedClient)
	if err != nil {
		util.Log.Println("[udp] Error adding adapter: ", err)
		return caching.ConnectedClient{}, false
	}

	return connectedClient, true
}

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
	"github.com/Fajurion/pipes/connection"
	"github.com/Fajurion/pipes/receive"

	pipesUtil "github.com/Fajurion/pipes/util"
)

const PrefixNode = 'n'      // For node auth & communication
const PrefixClient = 'c'    // For client auth
const PrefixEncrypted = 'e' // For encrypted traffic with the client

var udpServ *net.UDPConn

func Listen(domain string, port int) {

	connection.GeneralPrefix = []byte{PrefixNode, ':'}

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

		// Extract message
		msg := buffer[:offset]

		// Check if client wants to send to node
		ip := strings.Split(clientAddr.String(), ":")[0]
		connection, exists := caching.GetConnection(ip)
		node := msg[0] == PrefixNode

		if integration.Testing {
			util.Log.Println("[udp] Message from: ", ip, "with prefix: ", string(msg[0]), "found: ", exists)
		}

		if exists && node {

			if integration.Testing {
				util.Log.Println("[udp] Deleted client ", ip, " because it tried to send a node message")
			}

			caching.DeleteConnection(ip)
			continue
		}

		if node {
			err := receive.ReceiveUDP(msg[2:])
			if err != nil {
				util.Log.Println("[udp] Error receiving node message: ", err)
			}

			continue
		}

		// Register client
		if exists {

			if msg[0] != PrefixEncrypted {
				util.Log.Println("[udp] Error: No encryption prefix")
				caching.DeleteConnection(ip)
				continue
			}

			decrypted, err := pipesUtil.DecryptAES(connection.Key, msg[2:])
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
			err = ExecuteChannel(connection.ID, decrypted, clientAddr)
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
		stringMsg := string(msg[2:])
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

		if client.ID != accountId {
			util.Log.Println("[udp] Error: Invalid auth token (invalid account id)")
			caching.DeleteToken(accountId)
			continue
		}

		// Join room if correct
		connectedClient, authenticated := auth(secret, client, ip, clientAddr.String())
		if !authenticated {
			caching.DeleteConnection(ip)
			caching.DeleteUser(accountId)
			adapter.RemoveUDP(accountId)
			continue
		}

		SendConfirmation(clientAddr.String(), client.ID, &connectedClient.Key)
		caching.JoinRoom(room, accountId)
	}
}

func auth(secret string, client caching.Client, address string, clientAddress string) (caching.ConnectedClient, bool) {

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
	caching.DeleteToken(client.ID)
	connectedClient, valid := client.ToConnected(clientAddress)
	if !valid {
		return caching.ConnectedClient{}, false
	}

	caching.StoreConnection(connectedClient, clientAddress)
	caching.StoreUser(connectedClient)
	util.Log.Println("[udp]", connectedClient.ID+"("+connectedClient.Username+"#"+connectedClient.Tag+") connected")

	err = AddAdapter(connectedClient)
	if err != nil {
		util.Log.Println("[udp] Error adding adapter: ", err)
		return caching.ConnectedClient{}, false
	}

	return connectedClient, true
}

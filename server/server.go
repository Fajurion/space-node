package server

import (
	"log"
	"net"
	"time"

	"github.com/Fajurion/pipes/connection"
	"github.com/Fajurion/pipes/receive"
)

const PrefixNode = 'n'
const PrefixClient = 'c'

func Listen(domain string, port int) {

	connection.GeneralPrefix = []byte{PrefixNode, ':'}

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(domain),
	}

	log.Println("Starting UDP server..")

	// Start udp server
	udpServ, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Println("[udp] Error: ", err)
		panic(err)
	}
	defer udpServ.Close()

	buffer := make([]byte, 8*1024) // 8 kb buffer

	log.Println("UDP server started")

	for {
		offset, clientAddr, err := udpServ.ReadFrom(buffer) // Use client addr to rate limit in the future
		if err != nil {
			log.Println("[udp] Error: ", err)
			continue
		}

		// Extract message
		msg := buffer[:offset]

		// Check if client wants to send to node
		exists := ExistsClient(clientAddr.String())
		node := msg[0] == PrefixNode

		if exists && node {
			RemoveClient(clientAddr.String())
			continue
		}

		if node {
			err := receive.ReceiveUDP(msg[2:])
			if err != nil {
				log.Println("[udp] Error receiving node message: ", err)
			}

			continue
		}

		if msg[0] != PrefixClient {
			RemoveClient(clientAddr.String())
			continue
		}

		// Register client
		if exists {

			// Update last message
			client, err := GetClient(clientAddr.String())
			if err != nil {
				log.Println("[udp] Error getting client: ", err)
				continue
			}

			// Check if sent too quickly
			if time.Now().UnixMilli()-client.LastMessage < 50 {
				// TODO: Block
				continue
			}

			log.Println("[udp]", string(msg[2:]), offset)

			// Echo (for now)
			_, err = udpServ.WriteTo(msg[2:], clientAddr)
			if err != nil {
				log.Println("[udp] Error echoing message: ", err)
				continue
			}

			client.LastMessage = time.Now().UnixMilli()
			continue
		}

		// Data will be retrieved from node_backend later
		AddClient(clientAddr.String(), Client{
			Username:    "just_for_testing",
			Address:     clientAddr,
			LastMessage: time.Now().UnixMilli(),
		})

		// TODO: Add handler
	}

}

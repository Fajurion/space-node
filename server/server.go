package server

import (
	"log"
	"net"
	"time"
)

func Listen(domain string, port int) {

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

		// Register client
		if ExistsClient(clientAddr.String()) {

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

			log.Println("[udp]", string(msg), offset)

			// Echo (for now)
			_, err = udpServ.WriteTo(msg, clientAddr)
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

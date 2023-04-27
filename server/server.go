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

	// Start udp server
	go func() {

		// Start udp server
		udpServ, err := net.ListenUDP("udp", &addr)
		if err != nil {
			panic(err)
		}
		defer udpServ.Close()

		buffer := make([]byte, 8*1024) // 8 kb buffer

		for {
			offset, clientAddr, err := udpServ.ReadFromUDP(buffer) // Use client addr to rate limit in the future
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

				// Echo (for now)
				_, err = udpServ.WriteToUDP(msg, &addr)
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
	}()

}

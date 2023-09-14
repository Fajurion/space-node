package server

import (
	"net"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
)

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

		//* protocol standard: CLIENT_ID:VOICE_DATA
		// Client ID: 4 bytes
		// Voice data: rest of the packet
		go func(msg []byte) {
			if len(msg) < 9 {
				util.Log.Println("[udp] Error: Invalid message")
				return
			}

			// Check if client wants to send to node
			clientID := string(msg[0:4])
			ip := clientAddr.String() + ":" + clientID
			connection, exists := caching.GetConnection(ip)
			if !exists {
				util.Log.Println("[udp] Error: Connection not found")
				return
			}

			if connection.UDP == nil {

			}

			// Send voice data to room
			SendToRoom(connection.Room, msg[4:offset])

		}(buffer[:offset])
	}
}

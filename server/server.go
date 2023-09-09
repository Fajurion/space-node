package server

import (
	"net"

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
		offset, _, err := udpServ.ReadFrom(buffer) // Use client addr to rate limit in the future
		if err != nil {
			util.Log.Println("[udp] Error: ", err)
			continue
		}

		//* protocol standard: PREFIX+CLIENT_ID:VOICE_DATA
		// Prefix: 1 byte
		// Client ID: 4 bytes
		// Voice data: rest of the packet
		msg := buffer[:offset]

		if len(msg) < 9 {
			util.Log.Println("[udp] Error: Invalid message")
			continue
		}

		// Check if client wants to send to node
		//endIndex := 6
		//clientID := string(msg[1 : endIndex-1])
		//connection, exists := caching.GetConnection(ip)

		// TODO: New logic
	}
}

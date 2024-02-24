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

		//* protocol standard: CLIENT_ID:HASH:VOICE_DATA
		// Client ID: 10 bytes
		// Verifier: 32 bytes
		// Voice data: rest of the packet
		go func(msg []byte) {
			if len(msg) < 300 {
				util.Log.Println("[udp] Error: Invalid message length")
				return
			}

			// Verify connection
			clientID := string(msg[0:10])
			hash := msg[10:42]
			voiceData := msg[42:]

			conn, valid := caching.VerifyUDP(clientID, clientAddr, hash, voiceData)
			if !valid {
				util.Log.Println("[udp] Error: Could not verify connection or packet dropped")
				return
			}

			if len(voiceData) <= 5 {
				util.Log.Println("[udp] Success: Init packet received")
				return
			}

			// Send voice data to room
			SendToRoom(conn.Room, clientID, voiceData)

		}(buffer[:offset])
	}
}

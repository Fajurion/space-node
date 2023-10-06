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
		// Verifier: variable length (till next ':')
		// Voice data: rest of the packet
		go func(msg []byte) {
			if len(msg) < 52 {
				util.Log.Println("[udp] Error: Invalid message length")
				return
			}

			// Verify connection
			clientID := string(msg[0:10])
			beginning := 11
			end := beginning + 32 // Must be longer than 32 cause hash is 32 and encrypted = longer
			found := false
			for ; end < min(beginning+100, len(msg)); /* to prevent overflow */ end++ {
				if msg[end] == ':' {
					found = true
					break
				}
			}
			if !found {
				util.Log.Println("[udp] Error: Invalid message format")
				return
			}
			verifier := msg[beginning-1 : end]
			voiceData := msg[end+1 : offset]
			hashedData := util.Hash(voiceData)

			conn, valid := caching.VerifyUDP(clientID, clientAddr, hashedData, verifier)
			if !valid {
				util.Log.Println("[udp] Error: Could not verify connection")
				return
			}

			// Send voice data to room
			SendToRoom(conn.Room, []byte(clientID), voiceData)

		}(buffer[:offset])
	}
}

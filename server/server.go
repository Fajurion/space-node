package server

import (
	"log"
	"net"
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
			_ /* offset */, _, err := udpServ.ReadFromUDP(buffer) // Use client addr to rate limit in the future
			if err != nil {
				log.Println("[udp] Error: ", err)
				return
			}

			// msg := buffer[:offset]
			// TODO: Add handler
		}
	}()

}

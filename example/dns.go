package main

import (
	"fmt"
	"github.com/grantmd/go-airplay"
	"net"
)

func main() {
	fmt.Println("Listening for multicast DNS...")
	// Listen on the multicast address and port
	socket, err := net.ListenMulticastUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		panic(err)
	}
	// Don't forget to close it!
	defer socket.Close()

	var msg airplay.DNSMessage
	// Loop forever waiting for messages
	for {
		// Buffer for the message
		buffer := make([]byte, 4096)
		// Block and wait for a message on the socket
		read, _, err := socket.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		// Parse the buffer (up to "read" bytes) into a message object
		err = msg.Parse(buffer[:read])
		if err != nil {
			panic(err)
		}

		fmt.Println(msg.String())
	}
}

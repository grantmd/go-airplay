//
// http://www.ietf.org/rfc/rfc6762.txt - Multicast DNS
// http://www.ietf.org/rfc/rfc6763.txt - DNS-Based Service Discovery
//

package main

import (
	"log"
	"net"
)

func main() {
	log.Println("Starting up...")
	// Listen on the multicast address and port
	socket, err := net.ListenMulticastUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		log.Fatal(err)
	}
	// Don't forget to close it!
	defer socket.Close()

	log.Println("Waiting for messages...")
	// Loop forever waiting for messages
	for {
		// Buffer for the message
		buff := make([]byte, 4096)
		// Block and wait for a message on the socket
		read, addr, err := socket.ReadFromUDP(buff)
		if err != nil {
			log.Fatal(err)
		}
		// Print out the source address and the buffer up to "read" bytes
		log.Printf("%s: %s", addr, buff[:read])
	}
}

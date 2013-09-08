//
// Handles discovery of airplay devices on the local network. That means
// that this is basically a multicast DNS and service discovery client,
// except that we only implement the bare minimum to do airplay devices.
//
// Relevant RFCs:
// http://www.ietf.org/rfc/rfc6762.txt - Multicast DNS
// http://www.ietf.org/rfc/rfc6763.txt - DNS-Based Service Discovery
//

package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Starting up...")
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

	// Put the listener in its own goroutine
	fmt.Println("Waiting for messages...")
	go listen(socket)

	// Wait for a keypress to exit
	fmt.Println("Ctrl+C to exit")
	var input string
	fmt.Scanln(&input)
	fmt.Println("done")
}

func listen(socket *net.UDPConn) {
	// Loop forever waiting for messages
	for {
		// Buffer for the message
		buff := make([]byte, 4096)
		// Block and wait for a message on the socket
		read, addr, err := socket.ReadFromUDP(buff)
		if err != nil {
			panic(err)
		}
		// Print out the source address and the buffer up to "read" bytes
		fmt.Printf("%s: 0x%x\n", addr, buff[:read])
	}
}

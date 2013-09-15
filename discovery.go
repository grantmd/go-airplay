//
// Handles discovery of airplay devices on the local network. That means
// that this is basically a multicast DNS and service discovery client,
// except that we only implement the bare minimum to do airplay devices.
//
// Relevant RFCs:
// http://www.ietf.org/rfc/rfc6762.txt - Multicast DNS
// http://www.ietf.org/rfc/rfc6763.txt - DNS-Based Service Discovery
//
// http://nto.github.io/AirPlay.html - Unofficial AirPlay Protocol Specification
//

package main

import (
	"fmt"
	"net"
)

//
// Main functions for starting up and listening for records start here
//

func main() {
	fmt.Println("Starting up...")
	// Listen on the multicast address and port
	socketIn, err := net.ListenMulticastUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		panic(err)
	}
	// Don't forget to close it!
	defer socketIn.Close()

	// Put the listener in its own goroutine
	fmt.Println("Waiting for messages...")
	go listen(socketIn)

	// Bootstrap us by sending a query for any airplay-related entries
	var msg DNSMessage

	q := Question{
		Name:  "_raop._tcp.local.",
		Type:  12, // PTR
		Class: 1,
	}
	msg.AddQuestion(q)

	q = Question{
		Name:  "_airplay._tcp.local.",
		Type:  12, // PTR
		Class: 1,
	}
	msg.AddQuestion(q)

	buffer, err := msg.Pack()
	if err != nil {
		panic(err)
	}

	socketOut, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		panic(err)
	}
	// Don't forget to close it!
	defer socketOut.Close()

	// Write the payload
	_, err = socketIn.Write(buffer)
	if err != nil {
		panic(err)
	}

	// Wait for a keypress to exit
	fmt.Println("Ctrl+C to exit")
	var input string
	for {
		fmt.Scanln(&input)
	}
}

// Listen on a socket for multicast records and parse them
func listen(socket *net.UDPConn) {
	var msg DNSMessage
	// Loop forever waiting for messages
	for {
		// Buffer for the message
		buffer := make([]byte, 4096)
		// Block and wait for a message on the socket
		read, addr, err := socket.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		// Parse the buffer (up to "read" bytes) into a message object
		err = msg.Parse(buffer[:read])
		if err != nil {
			panic(err)
		}

		// Print out the source address and the message
		fmt.Printf("\nBroadcast from %s:\n%s\n", addr, msg.String())
	}
}

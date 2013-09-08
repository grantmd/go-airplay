//
// Handles discovery of airplay devices on the local network. That means
// that this is basically a multicast DNS and service discovery client,
// except that we only implement the bare minimum to do airplay devices.
//
// Relevant RFCs:
// http://www.ietf.org/rfc/rfc1035.txt - DNS
// http://www.ietf.org/rfc/rfc6762.txt - Multicast DNS
// http://www.ietf.org/rfc/rfc6763.txt - DNS-Based Service Discovery
//

package main

import (
	"fmt"
	"net"
)

// A representation of a full message, including header
type Message struct {
	// Header values
	Id                   uint16 // Can be used to match a response to a question
	IsResponse           bool   // True if this is a response, false for if this is a question
	Opcode               int    // What kind of query is this?
	IsAuthoritative      bool   // If this is a response, true if the responding server is authoritative for the question
	IsTruncated          bool   // True if the message was truncated
	IsRecursionDesired   bool   // Copied from the question, true if we want the server to process the query recursively
	IsRecursionAvailable bool   // True if the server supports recursive queries
	IsZero               bool   // Reserved, must be false
	Rcode                int    // Response code of the response, 0 for no errors
	//AuthenticatedData  bool
	//CheckingDisabled   bool

	// Message values
}

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

		// Parse the buffer (up to "read" bytes) into a message object
		msg, err := parseMessage(buff[:read])
		if err != nil {
			panic(err)
		}

		// Print out the source address and the message
		fmt.Printf("\n%s:\n%s\n", addr, msg.String())
	}
}

func parseMessage(buffer []byte) (Message, error) {
	offset := 0 // Point in the buffer that we are reading

	// Header first
	var msg Message

	msg.Id = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
	offset += 2

	msg.IsResponse = (buffer[offset] != 0)
	offset += 1

	msg.Opcode = int(buffer[offset])
	offset += 1

	msg.IsAuthoritative = (buffer[offset] != 0)
	offset += 1

	msg.IsTruncated = (buffer[offset] != 0)
	offset += 1

	msg.IsRecursionDesired = (buffer[offset] != 0)
	offset += 1

	msg.IsRecursionAvailable = (buffer[offset] != 0)
	offset += 1

	msg.IsZero = (buffer[offset] != 0)
	offset += 1

	msg.Rcode = int(buffer[offset])
	offset += 1

	// Now the rest of the message

	return msg, nil
}

// Map of strings for opcodes.
var OpcodeToString = map[int]string{
	0: "QUERY",  // A standard query
	1: "IQUERY", // An inverse query
	2: "STATUS", // A server status request
}

// Map of strings for rcodes.
var RcodeToString = map[int]string{
	0: "NOERROR",  // No error condition
	1: "FORMERR",  // Format error - The server was unable to interpret the query.
	2: "SERVFAIL", // Server failure
	3: "NXDOMAIN", // Name Error - Domain doesn't exist
	4: "NOTIMPL",  // Not implemented - The server doesn't support that query
	5: "REFUSED",  // The server refuses to process this query
}

// Convert a Message to a string, with dig-like headers:
//
//;; opcode: QUERY, status: NOERROR, id: 48404
//
//;; flags: qr aa rd ra;
func (m *Message) String() string {
	if m == nil {
		return "<nil> Message"
	}

	// Header fields
	s := ";; opcode: " + OpcodeToString[m.Opcode]
	s += ", status: " + RcodeToString[m.Rcode]
	s += ", id: " + string(m.Id) + "\n"

	s += ";; flags:"
	if m.IsResponse {
		s += " qr"
	}
	if m.IsAuthoritative {
		s += " aa"
	}
	if m.IsTruncated {
		s += " tc"
	}
	if m.IsRecursionDesired {
		s += " rd"
	}
	if m.IsRecursionAvailable {
		s += " ra"
	}
	if m.IsZero { // Hmm
		s += " z"
	}
	/*if m.IsAuthenticatedData {
		s += " ad"
	}
	if m.IsCheckingDisabled {
		s += " cd"
	}*/

	s += ";"

	// Message fields

	// All done
	return s
}

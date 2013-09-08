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
	"strconv"
)

//
// Main functions for starting up and listening for records start here
//

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

// Listen on a socket for multicast records and parse them
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
		msg := parseMessage(buff[:read])

		// Print out the source address and the message
		fmt.Printf("\n%s:\n%s\n", addr, msg.String())
	}
}

//
// Message parsing functions start here
//

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

	// Message values
	Question []Question       // Holds the resource records in the question section, usually one
	Answer   []ResourceRecord // Holds the resource records in the answer section
	Ns       []ResourceRecord // Holds the resource records in the authority section
	Extra    []ResourceRecord // Holds the resource records in the additional section
}

type Question struct {
	Name  string // The name of the domain
	Type  uint16 // The type of query
	Class uint16 // The class of the query (like 'IN' for the Internet)
}

type ResourceRecord struct {
	Name  string // The name of the domain
	Type  uint16 // The type of the RDATA field
	Class uint16 // The class of the RDATA field
	TTL   uint32 // Time to live of this record, in seconds. Discard when this passes. TODO: Convert this to an explicit expiry timestamp
	Rdata []byte // The data of the record
}

// Parse a bytestream into a Message struct
func parseMessage(buffer []byte) Message {
	offset := 0 // Point in the buffer that we are reading

	// Header first
	var msg Message

	msg.Id = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
	offset += 2

	msg.IsResponse = (buffer[offset]&(1<<7) != 0)
	msg.Opcode = int(buffer[offset]>>3) & 0xF
	msg.IsAuthoritative = (buffer[offset]&(1<<2) != 0)
	msg.IsTruncated = (buffer[offset]&(1<<1) != 0)
	msg.IsRecursionDesired = (buffer[offset] != 0)
	offset += 1

	msg.IsRecursionAvailable = (buffer[offset]&(1<<7) != 0)
	msg.IsZero = (buffer[offset]&(1<<6) != 0) // TODO: There's other stuff in here!
	msg.Rcode = int(buffer[offset] & 0xF)
	offset += 1

	// Now the rest of the message

	return msg
}

//
// Formatting of messages to strings starts here
//

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
	s += ", id: " + strconv.Itoa(int(m.Id)) + "\n"

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

	s += ";"

	// Message fields

	// All done
	return s
}

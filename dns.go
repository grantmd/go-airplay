//
// This is just enough DNS message parsing/construction/formatting to be
// able to handle the messages relevant to airplay service discovery.
// Plus, maybe a few other types that I find interesting or easy.
//
// The point is, you probably can't use this to do DNS on the wider internet.
//
// Relevant RFCs:
// http://www.ietf.org/rfc/rfc1035.txt - DNS
//

package main

import (
	"fmt"
	"strconv"
)

//
// Message parsing functions start here
//

// A representation of a full message, including header
type DNSMessage struct {
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
	Questions []Question       // Holds the resource records in the question section, usually one
	Answers   []ResourceRecord // Holds the resource records in the answer section
	Nss       []ResourceRecord // Holds the resource records in the authority section
	Extras    []ResourceRecord // Holds the resource records in the additional section
}

type Question struct {
	Name  string // The name of the domain
	Type  uint16 // The type of query
	Class uint16 // The class of the query (like 'IN' for the Internet)
}

type ResourceRecord struct {
	Name       string // The name of the domain
	Type       uint16 // The type of the RDATA field
	Class      uint16 // The class of the RDATA field
	CacheClear bool
	TTL        uint32 // Time to live of this record, in seconds. Discard when this passes. TODO: Convert this to an explicit expiry timestamp
	Rdata      []byte // The data of the record
}

// Parse a bytestream into a DNSMessage struct
func (msg *DNSMessage) Parse(buffer []byte) {
	//fmt.Printf("% #x\n", buffer)
	length := len(buffer)
	offset := 0 // Point in the buffer that we are reading

	// Header first
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
	qdcount := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
	msg.Questions = make([]Question, qdcount)
	offset += 2

	ancount := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
	msg.Answers = make([]ResourceRecord, ancount)
	offset += 2

	nscount := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
	msg.Nss = make([]ResourceRecord, nscount)
	offset += 2

	arcount := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
	msg.Extras = make([]ResourceRecord, arcount)
	offset += 2

	for i := 0; i < len(msg.Questions); i++ {
		name, offset1 := parseDomainName(buffer, offset)
		offset = offset1
		msg.Questions[i].Name = name

		msg.Questions[i].Type = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Questions[i].Class = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2
	}

	for i := 0; i < len(msg.Answers); i++ {
		name, offset1 := parseDomainName(buffer, offset)
		offset = offset1
		msg.Answers[i].Name = name

		msg.Answers[i].Type = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Answers[i].CacheClear = (buffer[offset]&0x80 == 0x80)
		if msg.Answers[i].CacheClear {
			msg.Answers[i].Class = uint16(buffer[offset]^0x80)<<8 | uint16(buffer[offset+1])
		} else {
			msg.Answers[i].Class = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		}
		offset += 2

		msg.Answers[i].TTL = uint32(uint32(buffer[offset])<<24 | uint32(buffer[offset+1])<<16 | uint32(buffer[offset+2])<<8 | uint32(buffer[offset+3]))
		offset += 4

		dataLength := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Answers[i].Rdata = buffer[offset : offset+int(dataLength)]
		offset += int(dataLength)
	}

	for i := 0; i < len(msg.Nss); i++ {
		name, offset1 := parseDomainName(buffer, offset)
		offset = offset1
		msg.Nss[i].Name = name

		msg.Nss[i].Type = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Nss[i].CacheClear = (buffer[offset]&0x80 == 0x80)
		if msg.Nss[i].CacheClear {
			msg.Nss[i].Class = uint16(buffer[offset]^0x80)<<8 | uint16(buffer[offset+1])
		} else {
			msg.Nss[i].Class = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		}
		offset += 2

		msg.Nss[i].TTL = uint32(uint32(buffer[offset])<<24 | uint32(buffer[offset+1])<<16 | uint32(buffer[offset+2])<<8 | uint32(buffer[offset+3]))
		offset += 4

		dataLength := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Nss[i].Rdata = buffer[offset : offset+int(dataLength)]
		offset += int(dataLength)
	}

	for i := 0; i < len(msg.Extras); i++ {
		name, offset1 := parseDomainName(buffer, offset)
		offset = offset1
		msg.Extras[i].Name = name

		msg.Extras[i].Type = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Extras[i].CacheClear = (buffer[offset]&0x80 == 0x80)
		if msg.Extras[i].CacheClear {
			msg.Extras[i].Class = uint16(buffer[offset]^0x80)<<8 | uint16(buffer[offset+1])
		} else {
			msg.Extras[i].Class = uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		}
		offset += 2

		msg.Extras[i].TTL = uint32(uint32(buffer[offset])<<24 | uint32(buffer[offset+1])<<16 | uint32(buffer[offset+2])<<8 | uint32(buffer[offset+3]))
		offset += 4

		dataLength := uint16(buffer[offset])<<8 | uint16(buffer[offset+1])
		offset += 2

		msg.Extras[i].Rdata = buffer[offset : offset+int(dataLength)]
		offset += int(dataLength)
	}

	// TODO: Make this an error and return it
	if length != offset {
		fmt.Printf("Expected %d, ended up with %d", length, offset)
	}
}

// Parse a domain name out of the message buffer. Requires access to the full message buffer in case it encounters a pointer
// to previously in the message. Takes an offset for where to start reading in the buffer.
// Returns string domain name and new offset
func parseDomainName(buffer []byte, offset int) (string, int) {
	name := ""
	for {
		// Pointer to somewhere else in the message?
		if buffer[offset]&0xC0 == 0xC0 {
			ptr := int(buffer[offset]^0xC0)<<8 | int(buffer[offset+1])
			ptrName, _ := parseDomainName(buffer, ptr)
			name += ptrName
			offset += 2
			break

		} else {
			// Nope, raw domain name
			labelLength := uint16(buffer[offset])
			offset += 1
			if labelLength == 0 {
				break
			}

			name += string(buffer[offset:offset+int(labelLength)]) + "."
			offset += int(labelLength)
		}
	}

	return name, offset
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

// Map of strings for each CLASS wire type.
var ClassToString = map[uint16]string{
	1:   "IN",
	2:   "CS",
	3:   "CH",
	4:   "HS",
	254: "NONE",
	255: "ANY",
}

// Map of strings for each RR wire type.
var TypeToString = map[uint16]string{
	1:  "A",
	2:  "NS",
	3:  "MD",
	4:  "MF",
	5:  "CNAME",
	6:  "SOA",
	7:  "MB",
	8:  "MG",
	9:  "MR",
	10: "NULL",
	11: "WKS",
	12: "PTR",
	13: "HINFO",
	14: "MINFO",
	15: "MX",
	16: "TXT",

	252: "AXFR",
	253: "MAILB",
	254: "MAILA",
	255: "ANY",
}

// Convert a Message to a string, with dig-like headers:
//
//;; opcode: QUERY, status: NOERROR, id: 48404
//
//;; flags: qr aa rd ra;
func (m *DNSMessage) String() string {
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
	s += " "
	s += "QUERY: " + strconv.Itoa(len(m.Questions)) + ", "
	s += "ANSWER: " + strconv.Itoa(len(m.Answers)) + ", "
	s += "AUTHORITY: " + strconv.Itoa(len(m.Nss)) + ", "
	s += "ADDITIONAL: " + strconv.Itoa(len(m.Extras)) + "\n"

	if len(m.Questions) > 0 {
		s += "\n;; QUESTION SECTION:\n"
		for i := 0; i < len(m.Questions); i++ {
			s += m.Questions[i].String() + "\n"
		}
	}

	if len(m.Answers) > 0 {
		s += "\n;; ANSWER SECTION:\n"
		for i := 0; i < len(m.Answers); i++ {
			s += m.Answers[i].String() + "\n"
		}
	}

	if len(m.Nss) > 0 {
		s += "\n;; AUTHORITY SECTION:\n"
		for i := 0; i < len(m.Nss); i++ {
			s += m.Nss[i].String() + "\n"
		}
	}

	if len(m.Extras) > 0 {
		s += "\n;; ADDITIONAL SECTION:\n"
		for i := 0; i < len(m.Extras); i++ {
			s += m.Extras[i].String() + "\n"
		}
	}

	// All done
	return s
}

func (q *Question) String() (s string) {
	// prefix with ; (as in dig)
	if len(q.Name) == 0 {
		s = ";.\t" // root label
	} else {
		s = ";" + q.Name + "\t"
	}
	s += ClassToString[q.Class] + "\t"
	s += " " + TypeToString[q.Type]
	return s
}

func (rr *ResourceRecord) String() string {
	var s string
	if len(rr.Name) == 0 {
		s += ".\t"
	} else {
		s += rr.Name + "\t"
	}
	s += strconv.FormatInt(int64(rr.TTL), 10) + "\t"
	s += ClassToString[rr.Class] + "\t"
	s += " " + TypeToString[rr.Type]
	return s
}

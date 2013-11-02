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

package airplay

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

var (
	deviceList []AirplayDevice
)

type AirplayDevice struct {
	Name     string
	Hostname string
	IP       net.IP
	Port     uint16
	Type     string
	Flags    map[string]string
}

//
// Main functions for starting up and listening for records start here
//

func Discover(devices chan []AirplayDevice) {
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
	msgs := make(chan DNSMessage)
	go listen(socket, msgs)

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

	// Write the payload
	_, err = socket.WriteToUDP(buffer, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		panic(err)
	}

	// Wait for a message from the listen goroutine
	for {
		msg = <-msgs

		//fmt.Println(msg.String())

		// Look for new devices
		for i := range msg.Answers {
			rr := &msg.Answers[i]

			// PTRs only
			if rr.Type != 12 {
				continue
			}

			// Figure out the name of this thing
			nameParts := strings.Split(rr.Rdata.(PTRRecord).Name, ".")

			deviceName := nameParts[0]
			deviceType := ""
			if nameParts[1] == "_raop" || nameParts[1] == "_airplay" {
				deviceType = "airplay"
			} else if nameParts[1] == "_touch-remote" {
				deviceType = "remote"
			} else {
				continue
			}

			// If this is a device we already know about, then update it
			// Otherwise, add it
			index := -1
			for i := range deviceList {
				if deviceList[i].Name == deviceName {
					index = i
					break
				}
			}

			if index == -1 {
				deviceList = append(deviceList, AirplayDevice{
					Name: deviceName,
					Type: deviceType,
				})

				index = len(deviceList) - 1
			} else {
				deviceList[index].Type = deviceType
			}
		}

		for i := range deviceList {
			deviceList[i].updateFromDNS(&msg)
		}

		// TODO: Ask for info on devices we don't have all the information about
		// TODO: Do this on a timer so we're not asking for things too often

		// Push it down the channel
		devices <- deviceList
	}
}

// Listen on a socket for multicast records and parse them
func listen(socket *net.UDPConn, msgs chan DNSMessage) {
	var msg DNSMessage
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

		// Does this have answers we are interested in? If so, return the whole message since the rest of it (Extras in particular)
		// is probably relevant
		for i := range msg.Answers {
			rr := &msg.Answers[i]

			// PTRs only
			if rr.Type != 12 {
				continue
			}

			// Is this an airplay address
			nameParts := strings.Split(rr.Name, ".")
			if nameParts[0] == "_raop" || nameParts[0] == "_airplay" {
				msgs <- msg
			} else if nameParts[0] == "_touch-remote" {
				msgs <- msg
			}
		}
	}
}

func (a *AirplayDevice) updateFromDNS(msg *DNSMessage) {
	//fmt.Println(msg)
	loop := true
	for loop {
		loop = false
		for i := range msg.Answers {
			rr := &msg.Answers[i]
			//fmt.Println(rr.String())

			// Figure out the name of this thing
			nameParts := strings.Split(rr.Name, ".")
			deviceName := nameParts[0]

			if a.Name == deviceName || a.Hostname == rr.Name {
				// Found it, now update it
				startOver := a.updateFromRR(rr)
				if startOver == true {
					loop = true
				}
			}
		}

		// See if the rest of the information for this device is in Extras
		for i := range msg.Extras {
			rr := &msg.Extras[i]
			//fmt.Println(rr.String())

			// Figure out the name of this thing
			nameParts := strings.Split(rr.Name, ".")
			deviceName := nameParts[0]

			// Find our existing device
			if a.Name == deviceName || a.Hostname == rr.Name {
				// Found it, now update it
				startOver := a.updateFromRR(rr)
				if startOver == true {
					loop = true
				}
			}
		}
	}
}

func (a *AirplayDevice) updateFromRR(rr *ResourceRecord) (startOver bool) {
	startOver = false
	switch rr.Type {
	case 1: // A
		if rr.Rdata.(ARecord).Address.IsGlobalUnicast() {
			a.IP = rr.Rdata.(ARecord).Address
		}
		break

	case 16: // TXT
		flags := rr.Rdata.(TXTRecord).CStrings
		a.Flags = make(map[string]string, len(flags))
		for _, flag := range flags {
			parts := strings.SplitN(flag, "=", 2)
			a.Flags[parts[0]] = parts[1]
		}
		break

	case 33: // SRV
		srv := rr.Rdata.(SRVRecord)
		if a.Hostname != srv.Target {
			startOver = true // We changed hostname, so we need to start over
		}
		a.Hostname = srv.Target
		a.Port = srv.Port
		break
	}

	return startOver
}

func (a *AirplayDevice) AudioChannels() int {
	c, err := strconv.Atoi(a.Flags["ch"])
	if err != nil {
		return 0
	}

	return c
}

func (a *AirplayDevice) AudioCodecs() []int {
	parts := strings.Split(a.Flags["cn"], ",")
	codecs := make([]int, len(parts))

	for i, c := range parts {
		c1, err := strconv.Atoi(c)
		if err != nil {
			c1 = -1
		}
		codecs[i] = c1
	}

	return codecs
}

func (a *AirplayDevice) EncryptionTypes() []int {
	parts := strings.Split(a.Flags["et"], ",")
	types := make([]int, len(parts))

	for i, t := range parts {
		t1, err := strconv.Atoi(t)
		if err != nil {
			t1 = -1
		}
		types[i] = t1
	}

	return types
}

func (a *AirplayDevice) MetadataTypes() []int {
	parts := strings.Split(a.Flags["md"], ",")
	types := make([]int, len(parts))

	for i, t := range parts {
		t1, err := strconv.Atoi(t)
		if err != nil {
			t1 = -1
		}
		types[i] = t1
	}

	return types
}

func (a *AirplayDevice) RequiresPassword() bool {
	if a.Flags["pw"] == "true" {
		return true
	}

	return false
}

func (a *AirplayDevice) AudioSampleRate() int {
	c, err := strconv.Atoi(a.Flags["sr"])
	if err != nil {
		return 0
	}

	return c
}

func (a *AirplayDevice) AudioSampleSize() int {
	c, err := strconv.Atoi(a.Flags["ss"])
	if err != nil {
		return 0
	}

	return c
}

func (a *AirplayDevice) Transports() []string {
	return strings.Split(a.Flags["tp"], ",")
}

func (a *AirplayDevice) ServerVersion() string {
	return a.Flags["vs"]
}

func (a *AirplayDevice) DeviceModel() string {
	return a.Flags["am"]
}

func (a *AirplayDevice) String() (str string) {
	str += fmt.Sprintf("%s (%s:%d)\n", a.Name, a.IP, a.Port)

	if a.Type == "airplay" {
		str += fmt.Sprintf("Device: %s v%s\n", a.DeviceModel(), a.ServerVersion())
		str += fmt.Sprintf("Audio Channels: %d, Sample: %dHz (%d-bit)\n", a.AudioChannels(), a.AudioSampleRate(), a.AudioSampleSize())

		str += "Supported Codecs: "
		for i, c := range a.AudioCodecs() {
			if i > 0 {
				str += ", "
			}

			switch c {
			case 0:
				str += "PCM"
				break
			case 1:
				str += "Apple Lossless (ALAC)"
				break
			case 2:
				str += "AAC"
				break
			case 3:
				str += "AAC ELD (Enhanced Low Delay)"
				break
			default:
				str += "Unknown"
				break
			}
		}
		str += "\n"

		str += "Supported Encryption Types: "
		for i, t := range a.EncryptionTypes() {
			if i > 0 {
				str += ", "
			}

			switch t {
			case 0:
				str += "None"
				break
			case 1:
				str += "RSA (AirPort Express)"
				break
			case 2:
				str += "FairPlay"
				break
			case 3:
				str += "MFiSAP (3rd-party devices)"
				break
			case 4:
				str += "FairPlay SAPv2.5"
				break
			default:
				str += "Unknown"
				break
			}
		}
		str += "\n"

		str += "Supported Metadata Types: "
		for i, t := range a.MetadataTypes() {
			if i > 0 {
				str += ", "
			}

			switch t {
			case 0:
				str += "text"
				break
			case 1:
				str += "artwork"
				break
			case 2:
				str += "progress"
				break
			default:
				str += "Unknown"
				break
			}
		}
		str += "\n"

		str += fmt.Sprintf("Transports: %s\n", strings.Join(a.Transports(), ", "))

		str += "Requires Password: "
		if a.RequiresPassword() {
			str += "Yes"
		} else {
			str += "No"
		}
	} else if a.Type == "remote" {
		str += fmt.Sprintf("Device: %s (%s)\n", a.Flags["DvNm"], a.Flags["DvTy"])
		str += fmt.Sprintf("Remote: %s v%s\n", a.Flags["RemN"], a.Flags["RemV"])
		str += fmt.Sprintf("Pair Code: %s", a.Flags["Pair"])
	} else {
		str += "Unsupported device"
	}

	return str
}

//
// http://blog.mycroes.nl/2008/08/pairing-itunes-remote-app-with-your-own.html
// http://jinxidoru.blogspot.com/2009/06/itunes-remote-pairing-code.html
// https://code.google.com/p/ytrack/wiki/DAAPDocumentation -- and other documentation
// http://www.awilco.net/doku/dacp -- this one has the pairing process explained well
//

package airplay

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	//"net"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrBadPin      = errors.New("Invalid pin")
	ErrInvalidDAAP = errors.New("Invalid DAAP response received")
)

type Remote struct {
	pin  string
	Name string
	Type string
	GUID string
}

type RemoteServer struct {
	Port    int
	Remotes []Remote
}

func StartRemoteServer() (rs RemoteServer, err error) {
	rs = RemoteServer{
		Port: 3690,
	}

	// Start our http server
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/server-info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	// Listen async
	go func() {
		err = http.ListenAndServe(":"+strconv.Itoa(rs.Port), serveMux)
		if err != nil {
			return
		}
	}()

	/*
		// Advertise ourselves on the network
		var msg DNSMessage

		rr := ResourceRecord{
			Name:  "_touch-able._tcp.local.",
			Type:  12, // PTR
			Class: 1,
		}
		msg.AddAnswer(rr)

		buffer, err := msg.Pack()
		if err != nil {
			panic(err)
		}

		// Write the payload
		socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.IPv4(224, 0, 0, 251),
			Port: 5353,
		})
		if err != nil {
			panic(err)
		}
		// Don't forget to close it!
		defer socket.Close()

		_, err = socket.WriteToUDP(buffer, &net.UDPAddr{
			IP:   net.IPv4(224, 0, 0, 251),
			Port: 5353,
		})
		if err != nil {
			panic(err)
		}
	*/

	return
}

func Pair(device AirplayDevice, pin string) (r Remote, err error) {
	// TODO: Validate pin?
	r.pin = pin

	// md5(pairingcode1-2-3-4-)
	codeBytes := []byte(device.Flags["Pair"])
	codeLength := len(codeBytes)
	pairingcode := make([]byte, codeLength+8)
	for i := range codeBytes {
		pairingcode[i] = codeBytes[i]
	}

	pinBytes := []byte(pin)
	pairingcode[codeLength] = pinBytes[0]
	pairingcode[codeLength+2] = pinBytes[1]
	pairingcode[codeLength+4] = pinBytes[2]
	pairingcode[codeLength+6] = pinBytes[3]

	hash := md5.New()
	hash.Write(pairingcode)

	// Immediately make a connection, just to make sure we can connect
	u := url.URL{
		Scheme:   "http",
		Host:     device.IP.String() + ":" + strconv.Itoa(int(device.Port)),
		Path:     "/pair",
		RawQuery: fmt.Sprintf("pairingcode=%s&servicename=%s", fmt.Sprintf("%X", hash.Sum(nil)), device.Name),
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("Viewer-Only-Client", "1")
	resp, err := client.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	if resp.StatusCode != 200 {
		return r, ErrBadPin
	}

	tags := DAAPParse(body)
	cmpa, ok := tags["cmpa"].(map[string]interface{})
	if ok == false {
		return r, ErrInvalidDAAP
	}

	r.Name = cmpa["cmnm"].(string)
	r.Type = cmpa["cmty"].(string)
	r.GUID = fmt.Sprintf("%X", cmpa["cmpg"])

	return r, nil
}

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
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrBadPin = errors.New("Invalid pin")
)

type Remote struct {
	pin string
}

func Pair(device AirplayDevice, pin string) (r Remote, err error) {
	// TODO: Validate pin
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
	resp, err := http.Get(u.String())
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

	fmt.Println("Body:")
	fmt.Println(string(body))

	return r, nil
}

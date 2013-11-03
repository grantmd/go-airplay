//
// http://blog.mycroes.nl/2008/08/pairing-itunes-remote-app-with-your-own.html
// http://jinxidoru.blogspot.com/2009/06/itunes-remote-pairing-code.html
// https://code.google.com/p/ytrack/wiki/DAAPDocumentation -- and other documentation
// http://www.awilco.net/doku/dacp -- this one has the pairing process explained well
//

package airplay

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type Remote struct {
	pin string
}

func Pair(ip net.IP, port uint16, pin string) (r Remote, err error) {
	r.pin = pin

	// Immediately make a connection, just to make sure we can connect
	u := url.URL{
		Scheme:   "http",
		Host:     ip.String() + ":" + strconv.Itoa(int(port)),
		Path:     "/pair",
		RawQuery: "pairingcode=75D809650423A40091193AA4944D1FBD&servicename=D19BB75C3773B485",
	}
	fmt.Println(u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	fmt.Println("Body:")
	fmt.Println(string(body))

	return r, nil
}

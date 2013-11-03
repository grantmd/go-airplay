package airplay

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	ErrPasswordRequired = errors.New("Password required")
	ErrPasswordInvalid  = errors.New("Password invalid")
	ErrAuthUnsupported  = errors.New("Authentication not supported")
	ErrNoOptions        = errors.New("Airplay server did not respond to OPTIONS request")
	ErrInvalidOptions   = errors.New("Airplay server reported invalid OPTIONS")
)

type Airplay struct {
	Password    string
	conn        *textproto.Conn
	reverseConn *textproto.Conn
	sessionID   string
	nonce       string
	realm       string
	cseq        int
}

func Dial(ip net.IP, port uint16, password string) (a Airplay, err error) {
	a.Password = password
	uuid, err := uuid.NewV4()
	if err != nil {
		return a, err
	}
	a.sessionID = uuid.String()
	a.cseq = 0

	// Immediately make a connection and ask for OPTIONS, just to make sure we can connect
	a.conn, err = textproto.Dial("tcp", ip.String()+":"+strconv.Itoa(int(port)))
	if err != nil {
		return a, err
	}

	resp, err := a.makeRTSPRequest("OPTIONS", "*")
	if err != nil {
		return a, err
	}

	if resp.StatusCode != 200 {
		return a, ErrNoOptions
	}

	methods := resp.Header.Get("Public")
	if methods == "" || strings.Contains(methods, "ANNOUNCE") == false {
		return a, ErrInvalidOptions
	}

	/*
		a.reverseConn, err = textproto.Dial("tcp", ip.String()+":"+strconv.Itoa(int(port)))
		if err != nil {
			return a, err
		}

		err = a.makeReverseRequest()
		if err != nil {
			return a, err
		}
	*/

	return a, nil
}

func (a *Airplay) IsConnected() bool {
	if a.conn == nil {
		return false
	}

	return true
}

func (a *Airplay) GetServerInfo() (err error) {
	resp, err := a.makeHTTPRequest("GET", "/server-info")
	if err != nil {
		return err
	}

	fmt.Println(resp.Body)
	return nil
}

func (a *Airplay) makeRTSPRequest(method string, path string) (resp http.Response, err error) {
	a.cseq++
	err = a.conn.PrintfLine("%s %s RTSP/1.0", method, path)
	if err != nil {
		return resp, err
	}

	a.conn.PrintfLine("Content-Length: 0")
	a.conn.PrintfLine("User-Agent: go-airplay/1.0")
	a.conn.PrintfLine("X-Apple-Session-ID: %s", a.sessionID)
	a.conn.PrintfLine("CSeq: %d", a.cseq)

	// Add auth headers, if necessary
	if a.realm != "" {
		username := ""
		if a.realm == "raop" {
			username = "iTunes"
		} else if a.realm == "Airplay" {
			username = "Airplay"
		}

		hash := md5.New()
		io.WriteString(hash, username+":"+a.realm+":"+a.Password)
		ha1 := fmt.Sprintf("%x", hash.Sum(nil))
		hash.Reset()

		io.WriteString(hash, method+":"+path)
		ha2 := fmt.Sprintf("%x", hash.Sum(nil))
		hash.Reset()

		io.WriteString(hash, ha1+":"+a.nonce+":"+ha2)
		response := fmt.Sprintf("%x", hash.Sum(nil))

		a.conn.PrintfLine("Authorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"", username, a.realm, a.nonce, path, response)
	}

	// Submit request
	err = a.conn.PrintfLine("")
	if err != nil {
		return resp, err
	}

	// Read response
	line, err := a.conn.ReadLine()
	if err != nil {
		return resp, err
	}

	//fmt.Println(line)
	f := strings.SplitN(line, " ", 3) // Proto, Code, Status
	reasonPhrase := ""
	if len(f) > 2 {
		reasonPhrase = f[2]
	}
	resp.Status = f[1] + " " + reasonPhrase
	resp.StatusCode, err = strconv.Atoi(f[1])
	if err != nil {
		return resp, err
	}

	resp.Proto = f[0]

	headers, err := a.conn.ReadMIMEHeader()
	if err != nil {
		return resp, err
	}

	resp.Header = http.Header(headers)

	//fmt.Println(headers)

	// Do auth
	if f[1] == "401" {
		if a.Password == "" {
			return resp, ErrPasswordRequired
		} else if a.realm == "" {
			// Parse the headers
			auth := headers.Get("WWW-Authenticate")
			if auth == "" {
				return resp, ErrAuthUnsupported
			}

			authParts := strings.Split(auth, " ")
			if authParts[0] != "Digest" {
				return resp, ErrAuthUnsupported
			}

			for i := range authParts {
				if i == 0 {
					continue
				}

				parts := strings.SplitN(authParts[i], "=", 2)
				value := strings.Trim(parts[1], "\"")
				if parts[0] == "nonce" {
					a.nonce = value

				} else if parts[0] == "realm" {
					a.realm = value
				}
			}

			// Make another request with the new auth information
			return a.makeRTSPRequest(method, path)

		} else {
			// We've already tried auth and failed
			return resp, ErrPasswordInvalid
		}
	}

	return resp, nil
}

func (a *Airplay) makeHTTPRequest(method string, path string) (resp http.Response, err error) {
	a.cseq++
	err = a.conn.PrintfLine("%s %s HTTP/1.1", method, path)
	if err != nil {
		return resp, err
	}

	a.conn.PrintfLine("Content-Length: 0")
	a.conn.PrintfLine("User-Agent: go-airplay/1.0")
	a.conn.PrintfLine("X-Apple-Session-ID: %s", a.sessionID)
	a.conn.PrintfLine("CSeq: %d", a.cseq)

	// Add auth headers, if necessary
	if a.realm != "" {
		username := ""
		if a.realm == "raop" {
			username = "iTunes"
		} else if a.realm == "Airplay" {
			username = "Airplay"
		}

		hash := md5.New()
		io.WriteString(hash, username+":"+a.realm+":"+a.Password)
		ha1 := fmt.Sprintf("%x", hash.Sum(nil))
		hash.Reset()

		io.WriteString(hash, method+":"+path)
		ha2 := fmt.Sprintf("%x", hash.Sum(nil))
		hash.Reset()

		io.WriteString(hash, ha1+":"+a.nonce+":"+ha2)
		response := fmt.Sprintf("%x", hash.Sum(nil))

		a.conn.PrintfLine("Authorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"", username, a.realm, a.nonce, path, response)
	}

	// Submit request
	err = a.conn.PrintfLine("")
	if err != nil {
		return resp, err
	}

	// Read response
	line, err := a.conn.ReadLine()
	if err != nil {
		return resp, err
	}

	fmt.Println(line)
	f := strings.SplitN(line, " ", 3) // Proto, Code, Status
	reasonPhrase := ""
	if len(f) > 2 {
		reasonPhrase = f[2]
	}
	resp.Status = f[1] + " " + reasonPhrase
	resp.StatusCode, err = strconv.Atoi(f[1])
	if err != nil {
		return resp, err
	}

	resp.Proto = f[0]

	headers, err := a.conn.ReadMIMEHeader()
	if err != nil {
		return resp, err
	}

	resp.Header = http.Header(headers)

	//fmt.Println(headers)

	// Do auth
	if f[1] == "401" {
		if a.Password == "" {
			return resp, ErrPasswordRequired
		} else if a.realm == "" {
			// Parse the headers
			auth := headers.Get("WWW-Authenticate")
			if auth == "" {
				return resp, ErrAuthUnsupported
			}

			authParts := strings.Split(auth, " ")
			if authParts[0] != "Digest" {
				return resp, ErrAuthUnsupported
			}

			for i := range authParts {
				if i == 0 {
					continue
				}

				parts := strings.SplitN(authParts[i], "=", 2)
				value := strings.Trim(parts[1], "\"")
				if parts[0] == "nonce" {
					a.nonce = value

				} else if parts[0] == "realm" {
					a.realm = value
				}
			}

			// Make another request with the new auth information
			return a.makeHTTPRequest(method, path)

		} else {
			// We've already tried auth and failed
			return resp, ErrPasswordInvalid
		}
	}

	return resp, nil
}

// Sets up the reverse HTTP connection
func (a *Airplay) makeReverseRequest() (err error) {
	err = a.reverseConn.PrintfLine("POST /reverse RTSP/1.0")
	if err != nil {
		return err
	}
	a.reverseConn.PrintfLine("Upgrade: PTTH/1.0")
	a.reverseConn.PrintfLine("Connection: Upgrade")
	a.reverseConn.PrintfLine("X-Apple-Purpose: Event")
	a.reverseConn.PrintfLine("Content-Length: 0")
	a.reverseConn.PrintfLine("User-Agent: go-airplay/1.0")
	a.reverseConn.PrintfLine("X-Apple-Session-ID: %s", a.sessionID)

	// Add auth headers, if necessary
	if a.realm != "" {
		username := ""
		if a.realm == "raop" {
			username = "iTunes"
		} else if a.realm == "Airplay" {
			username = "Airplay"
		}

		hash := md5.New()
		io.WriteString(hash, username+":"+a.realm+":"+a.Password)
		ha1 := fmt.Sprintf("%x", hash.Sum(nil))
		hash.Reset()

		io.WriteString(hash, "POST:/reverse")
		ha2 := fmt.Sprintf("%x", hash.Sum(nil))
		hash.Reset()

		io.WriteString(hash, ha1+":"+a.nonce+":"+ha2)
		response := fmt.Sprintf("%x", hash.Sum(nil))

		a.reverseConn.PrintfLine("Authorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"/reverse\", response=\"%s\"", username, a.realm, a.nonce, response)
	}

	// Submit request
	err = a.reverseConn.PrintfLine("")
	if err != nil {
		return err
	}

	// Read response
	line, err := a.reverseConn.ReadLine()
	if err != nil {
		return err
	}

	fmt.Println(line)
	f := strings.SplitN(line, " ", 3) // Proto, Code, Status

	headers, err := a.reverseConn.ReadMIMEHeader()
	if err != nil {
		return err
	}

	fmt.Println(headers)

	// Do auth
	if f[1] == "401" {
		if a.Password == "" {
			return ErrPasswordRequired
		} else if a.realm == "" {
			// Parse the headers
			auth := headers.Get("WWW-Authenticate")
			if auth == "" {
				return ErrAuthUnsupported
			}

			authParts := strings.Split(auth, " ")
			if authParts[0] != "Digest" {
				return ErrAuthUnsupported
			}

			for i := range authParts {
				if i == 0 {
					continue
				}

				parts := strings.SplitN(authParts[i], "=", 2)
				value := strings.Trim(parts[1], "\"")
				if parts[0] == "nonce" {
					a.nonce = value

				} else if parts[0] == "realm" {
					a.realm = value
				}
			}

			// Make another request with the new auth information
			return a.makeReverseRequest()

		} else {
			// We've already tried auth and failed
			return ErrPasswordInvalid
		}
	}

	return nil
}

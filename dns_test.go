//
// Tests formed from actual DNS records I have seen on my local network
//

package main

import (
	"encoding/hex"
	"testing"
)

func TestPTR1(t *testing.T) {
	bytes, err := hex.DecodeString("000084000000000100000000095f7365727669636573075f646e732d7364045f756470056c6f63616c00000c00010000119400150d5f6170706c652d6d6f62646576045f746370c023")
	if err != nil {
		t.Fatal(err)
	}

	////////
	var msg DNSMessage
	err = msg.Parse(bytes)
	if err != nil {
		t.Fatal(err)
	}

	////////
	if msg.Id != 0 {
		t.Errorf("Unexpected message id: %d", msg.Id)
	}
	if msg.Opcode != 0 {
		t.Errorf("Unexpected opcode: %d", msg.Opcode)
	}
	if OpcodeToString[msg.Opcode] != "QUERY" {
		t.Errorf("Unexpected opcode string: %s", OpcodeToString[msg.Opcode])
	}
	if msg.IsResponse != true {
		t.Error("Message was not a response")
	}
	if msg.IsAuthoritative != true {
		t.Error("Message was not authoritative")
	}
	if msg.IsTruncated != false {
		t.Error("Message was truncated")
	}
	if msg.IsRecursionDesired != true {
		t.Error("Message was not recursion desired")
	}
	if msg.IsRecursionAvailable != false {
		t.Error("Message was recursion desired")
	}
	if msg.IsZero != false {
		t.Error("Message had zero bit set")
	}
	if msg.Rcode != 0 {
		t.Errorf("Unexpected rcode: %d", msg.Rcode)
	}
	if RcodeToString[msg.Rcode] != "NOERROR" {
		t.Errorf("Unexpected rcode string: %s", RcodeToString[msg.Rcode])
	}

	////////
	if len(msg.Questions) != 0 {
		t.Errorf("Question length was not 0: %d", len(msg.Questions))
	}
	if len(msg.Answers) != 1 {
		t.Errorf("Answer length was not 1: %d", len(msg.Answers))
	}
	if len(msg.Nss) != 0 {
		t.Errorf("Authority length was not 0: %d", len(msg.Nss))
	}
	if len(msg.Extras) != 0 {
		t.Errorf("Extra length was not 0: %d", len(msg.Extras))
	}

	////////
	for i := 0; i < len(msg.Answers); i++ {
		rr := msg.Answers[i]
		if rr.Name != "_services._dns-sd._udp.local." {
			t.Errorf("Unexpected resource record name: %s", rr.Name)
		}
		if rr.Type != 12 {
			t.Errorf("Resource record type was not 12: %d", rr.Type)
		}
		if TypeToString[rr.Type] != "PTR" {
			t.Errorf("Unexpected resource record type string: %s", TypeToString[rr.Type])
		}
		if rr.Class != 1 {
			t.Errorf("Resource record class was not 1: %d", rr.Class)
		}
		if ClassToString[rr.Class] != "IN" {
			t.Errorf("Unexpected resource record class string: %s", ClassToString[rr.Class])
		}
		if rr.CacheClear != false {
			t.Error("Resource record cache clear was not false")
		}
		if rr.TTL != 4500 {
			t.Errorf("Resource record TTL was not 4500: %d", rr.TTL)
		}

		if rr.Rdata.(PTRRecord).Name != "_apple-mobdev._tcp.local." {
			t.Errorf("Unexpected resource record PTR domain name: %s", rr.Rdata.(PTRRecord).Name)
		}
	}
}

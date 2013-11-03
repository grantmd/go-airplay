//
// Tests formed from actual DNS records I have seen on my local network
//

package airplay

import (
	"encoding/hex"
	//"fmt"
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

func TestANY1(t *testing.T) {
	bytes, err := hex.DecodeString("0000000000030000000300002a30633a37343a63323a64353a32343a323440666538303a3a6537343a633266663a666564353a323432340d5f6170706c652d6d6f62646576045f746370056c6f63616c0000ff0001174d6f62696c652d436f6d707574696e672d446576696365c04a00ff0001c05500ff0001c00c0021000100000078000800000000f27ec055c055001c0001000000780010fe800000000000000e74c2fffed52424c055000100010000007800040a000110")
	if err != nil {
		t.Fatal(err)
	}

	////////
	var msg DNSMessage
	err = msg.Parse(bytes)
	if err != nil {
		t.Fatal(err)
	}

	//fmt.Println(msg.String())
}

func TestTXT1(t *testing.T) {

	bytes, err := hex.DecodeString("0000840000000005000000080b4c6976696e6720526f6f6d085f616972706f7274045f746370056c6f63616c00001080010000119400a6a577614d413d30302d32342d33362d39412d43382d38432c72614d413d30302d32342d33362d39412d43382d38442c72614e6d3d4861766f63472c726143683d3134392c726153743d302c72614e413d302c737944733d4170706c6520426173652053746174696f6e2056372e362e342c7379466c3d3078384138432c737941503d3130372c737956733d372e362e342c737263763d37363430302e31302c626a53643d3232c018000c0001000011940002c00c0b4c6976696e6720526f6f6d0c5f6465766963652d696e666fc02100100001000011940013126d6f64656c3d416972506f7274342c31303718303032343336394143383843404c6976696e6720526f6f6d055f72616f70c0210010800100001194008a09747874766572733d310463683d3206636e3d302c3104656b3d310665743d302c310873763d66616c73650764613d747275650873723d34343130300573733d31360770773d7472756508766e3d36353533370a74703d5443502c5544500876733d3130352e310f616d3d416972506f7274342c3130370b66763d37363430302e31300673663d307834c13c000c0001000011940002c1230b4c6976696e672d526f6f6dc026001c8001000000780010fe80000000000000022436fffe9ac88cc00c00218001000000780008000000001391c1e6c12300218001000000780008000000001388c1e6c1e600018001000000780004c0a80178c1e600018001000000780004a9fe74ffc00c002f8001000011940009c00c00050000800040c1e6002f8001000000780008c1e6000440000008c123002f8001000011940009c12300050000800040")
	if err != nil {
		t.Fatal(err)
	}

	////////
	var msg DNSMessage
	err = msg.Parse(bytes)
	if err != nil {
		t.Fatal(err)
	}

	//fmt.Println(msg.String())
}

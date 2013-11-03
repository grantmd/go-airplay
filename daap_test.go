package airplay

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestDAAPParse(t *testing.T) {
	bytes, err := hex.DecodeString("636d70610000003d636d7067000000083ae031c80b6318c9636d6e6d000000174d6f62696c6520436f6d707574696e6720446576696365636d7479000000066950686f6e65")
	if err != nil {
		t.Fatal(err)
	}

	////////
	tags := DAAPParse(bytes)

	fmt.Println(DAAPPrint(tags, ""))
}

package airplay

import (
	"fmt"
)

var (
	DAAPGroups = map[string]bool{
		"cmst": true,
		"mlog": true,
		"agal": true,
		"mlcl": true,
		"mshl": true,
		"mlit": true,
		"abro": true,
		"abar": true,
		"apso": true,
		"caci": true,
		"avdb": true,
		"cmgt": true,
		"aply": true,
		"adbs": true,
		"cmpa": true,
	}
)

func DAAPParse(buffer []byte) (tags map[string]interface{}) {
	tags = make(map[string]interface{}, 100) // TODO: Make this a better capacity ;-)

	length := len(buffer)
	offset := 0
	for offset < length {
		tag := string(buffer[offset : offset+4])
		offset += 4
		size := int(buffer[offset])<<24 | int(buffer[offset+1])<<16 | int(buffer[offset+2])<<8 | int(buffer[offset+3])
		offset += 4

		if DAAPGroups[tag] {
			data := DAAPParse(buffer[offset : offset+size])
			tags[tag] = data
		} else {
			data := string(buffer[offset : offset+size])
			tags[tag] = data
		}

		offset += size
	}

	return
}

func DAAPPrint(tags map[string]interface{}, indent string) (out string) {
	for tag := range tags {
		if DAAPGroups[tag] {
			out += fmt.Sprintf("%s%s:\n", indent, tag)
			out += DAAPPrint(tags[tag].(map[string]interface{}), indent+"\t")
		} else {
			out += fmt.Sprintf("%s%s: %s\n", indent, tag, tags[tag])
		}
	}

	return
}

package compact_time

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func demonstrateEncode() {
	location, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		// TODO: Handle error
	}
	goDate := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, location)
	compactDate := AsCompactTime(goDate)
	buffer := &bytes.Buffer{}
	encodedCount, err := compactDate.Encode(buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Encoded [%v] into %v bytes: %v\n", goDate, encodedCount, buffer.Bytes())
	// Prints: Encoded [2020-08-30 15:33:14.019577323 +0800 +08] into 21 bytes: [95 207 85 9 156 240 121 68 1 22 83 47 83 105 110 103 97 112 111 114 101]
}

func demonstrateDecode() {
	buffer := []byte{0x28, 0x9a, 0x12, 0x78, 0x08}
	compactDate, decodedCount, err := DecodeTimestamp(bytes.NewBuffer(buffer))
	if err != nil {
		// TODO: Handle error
	}
	goDate, err := compactDate.AsGoTime()
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into [%v]\n", decodedCount, buffer, goDate)
	// Prints: Decoded 5 bytes of [40 154 18 120 8] into [1966-12-01 05:13:05 +0000 UTC]
}

func TestReadmeExamples(t *testing.T) {
	demonstrateEncode()
	demonstrateDecode()
}

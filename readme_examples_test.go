package compact_time

import (
	"fmt"
	"testing"
	"time"
)

func demonstrateEncode() {
	location, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		// TODO: Deal with this
	}
	date := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, location)
	buffer := make([]byte, TimestampEncodedSize(date))
	encodedCount, ok, err := EncodeTimestamp(date, buffer)
	if err != nil {
		// TODO: Handle error
	}
	if !ok {
		// TODO: Not enough room in buffer to encode
	}
	fmt.Printf("Encoded [%v] into %v bytes: %v\n", date, encodedCount, buffer)
	// Prints: Encoded [2020-08-30 15:33:14.019577323 +0800 +08] into 21 bytes: [59 225 243 184 158 171 18 0 80 22 83 47 83 105 110 103 97 112 111 114 101]
}

func demonstrateDecode() {
	buffer := []byte{0x14, 0x4d, 0x09, 0x1c, 0x07}
	date, decodedCount, ok, err := DecodeTimestamp(buffer)
	if err != nil {
		// TODO: Handle error
	}
	if !ok {
		// TODO: Not enough bytes in buffer to decode
	}
	fmt.Printf("Decoded %v bytes of %v into [%v]\n", decodedCount, buffer, date)
	// Prints: Decoded 5 bytes of [20 77 9 28 7] into [1966-12-01 05:13:05 +0000 UTC]
}

func TestReadmeExamples(t *testing.T) {
	demonstrateEncode()
	demonstrateDecode()
}

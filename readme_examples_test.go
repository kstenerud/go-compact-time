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
	encodedCount, err := EncodeTimestamp(date, buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Encoded [%v] into %v bytes: %v\n", date, encodedCount, buffer)
}

func demonstrateDecode() {
	buffer := []byte{0x14, 0x4d, 0x09, 0x1c, 0x07}
	date, decodedCount, err := DecodeTimestamp(buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into [%v]\n", decodedCount, buffer, date)
}

func TestReadmeExamples(t *testing.T) {
	demonstrateEncode()
	demonstrateDecode()
}

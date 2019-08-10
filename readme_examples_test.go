package compact_date

import (
	"fmt"
	"testing"
	"time"
)

func demonstrateEncode() {
	date := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, time.UTC)
	buffer := make([]byte, EncodedSize(date))
	encodedCount, err := Encode(date, buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Encoded %v into %v bytes: %v\n", date, encodedCount, buffer)
}

func demonstrateDecode() {
	buffer := []byte{0x14, 0x4d, 0x07, 0x10, 0x03}
	decodedCount, date, err := Decode(buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into %v\n", decodedCount, buffer, date)
}

func TestReadmeExamples(t *testing.T) {
	demonstrateEncode()
	demonstrateDecode()
}

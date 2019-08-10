package compact_date

import (
	"bytes"
	"testing"
	"time"
)

func assertEncodedSize(t *testing.T, year int, nanoseconds int, expectedSize int) {
	d := time.Date(year, time.Month(1), 1, 0, 0, 0, nanoseconds, time.UTC)
	size := EncodedSize(d)
	if size != expectedSize {
		t.Errorf("Expected %v but got %v", expectedSize, size)
	}
}

func assertEncodeDecode(t *testing.T, year int, month int, day int, hour int, minute int, second int, nanosecond int, expected []byte) {
	actual := make([]byte, len(expected))
	expectedDate := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC)
	encodedCount, err := Encode(expectedDate, actual)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected %v encoded bytes but got %v", len(expected), encodedCount)
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected %v but got %v", expected, actual)
	}

	decodedCount, actualDate, err := Decode(expected)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected %v decoded bytes but got %v", len(expected), decodedCount)
	}
	if actualDate != expectedDate {
		t.Errorf("Expected %v but got %v", expectedDate, actualDate)
	}
}

func TestEncodedSize(t *testing.T) {
	for year := 1995; year < 2004; year++ {
		assertEncodedSize(t, year, 0, 5)
	}

	for year := 1995; year < 2004; year++ {
		assertEncodedSize(t, year, 10000000, 6)
	}

	for year := 1995; year < 2004; year++ {
		assertEncodedSize(t, year, 10000, 7)
	}

	for year := 1995; year < 2004; year++ {
		assertEncodedSize(t, year, 10, 9)
	}

	assertEncodedSize(t, 7000, 10, 10)
	assertEncodedSize(t, 1935, 10000, 8)
	assertEncodedSize(t, 2064, 10000, 8)
}

func TestEncodeDecode(t *testing.T) {
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, []byte{0x00, 0x00, 0x01, 0x10, 00})
	assertEncodeDecode(t, 2000, 1, 1, 1, 0, 0, 0, []byte{0x00, 0x00, 0x21, 0x10, 0x00})
	assertEncodeDecode(t, 2000, 1, 1, 0, 1, 0, 0, []byte{0x00, 0x04, 0x01, 0x10, 0x00})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 1, 0, []byte{0x01, 0x00, 0x01, 0x10, 0x00})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 1000000, []byte{0x40, 0x10, 0x00, 0x00, 0x44, 0x00})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000000, []byte{0x7e, 0x70, 0x00, 0x00, 0x44, 0x00})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000, []byte{0x80, 0x0f, 0x9c, 0x00, 0x00, 0x11, 0x00})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999, []byte{0xc0, 0x00, 0x03, 0xe7, 0x00, 0x00, 0x04, 0x40, 0x00})
	assertEncodeDecode(t, 2009, 1, 1, 0, 0, 0, 0, []byte{0x00, 0x00, 0x01, 0x10, 0x12})
	assertEncodeDecode(t, 3009, 1, 1, 0, 0, 0, 0, []byte{0x00, 0x00, 0x01, 0x1f, 0x62})
	assertEncodeDecode(t, -50000, 1, 1, 0, 0, 0, 0, []byte{0x00, 0x00, 0x01, 0x16, 0xac, 0x3f})

	// June 24, 2019, 17:53:04.180
	assertEncodeDecode(t, 2019, 6, 24, 17, 53, 4, 180000000, []byte{0x4b, 0x41, 0x35, 0x8e, 0x18, 0x26})

	// January 7, 1998, 08:19:20
	assertEncodeDecode(t, 1998, 1, 7, 8, 19, 20, 0, []byte{0x14, 0x4d, 0x07, 0x10, 0x03})

	// August 31, 3190, 00:54:47.394129
	assertEncodeDecode(t, 3190, 8, 31, 0, 54, 47, 394129000, []byte{0x98, 0x0e, 0x46, 0xfd, 0x81, 0xf8, 0x92, 0x4c})
}

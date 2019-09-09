package compact_date

import (
	"bytes"
	"testing"
	"time"
)

func assertEncodeDecode(t *testing.T, year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone string, expected []byte) {
	actual := make([]byte, len(expected))
	location := time.UTC
	if len(timezone) > 0 {
		var err error
		location, err = time.LoadLocation(timezone)
		if err != nil {
			t.Errorf("BUG IN TEST CODE. Error loading location %v: %v", timezone, err)
		}
	}
	expectedDate := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, location)
	actualSize := EncodedSize(expectedDate)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
	}
	encodedCount, err := EncodeTimestamp(expectedDate, actual)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", expected, actual)
	}

	actualDate, decodedCount, err := DecodeTimestamp(expected)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
	}
	if !actualDate.Equal(expectedDate) {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedDate, actualDate)
	}
}

func TestEncodeDecode(t *testing.T) {
	assertEncodeDecode(t, 2020, 8, 30, 15, 33, 14, 19577323, "Asia/Singapore", []byte{0x3b, 0xe1, 0xf3, 0xb8, 0x9e, 0xab, 0x12, 0x00, 0x50, 0x16, 'S', '/', 'S', 'i', 'n', 'g', 'a', 'p', 'o', 'r', 'e'})
	assertEncodeDecode(t, 1966, 12, 1, 5, 13, 5, 0, "", []byte{0x14, 0x4d, 0x09, 0x1c, 0x07})

	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "Europe/Berlin", []byte{0x00, 0x00, 0x08, 0x01, 00, 0x10, 'E', '/', 'B', 'e', 'r', 'l', 'i', 'n'})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0x01, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 1, 0, 0, 0, "", []byte{0x00, 0x40, 0x08, 0x01, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 0, 1, 0, 0, "", []byte{0x00, 0x01, 0x08, 0x01, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 1, 0, "", []byte{0x04, 0x00, 0x08, 0x01, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 1000000, "", []byte{0x01, 0x00, 0x08, 0x11, 0x00, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000000, "", []byte{0x01, 0x00, 0x08, 0x71, 0x3e, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000, "", []byte{0x02, 0x00, 0x08, 0x71, 0x3e, 0x00, 0x01})
	assertEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999, "", []byte{0x03, 0x00, 0x08, 0x71, 0x3e, 0x00, 0x00, 0x00, 0x01})
	assertEncodeDecode(t, 2009, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0x01, 0x25})
	assertEncodeDecode(t, 3009, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0x01, 0x9f, 0x45})
	assertEncodeDecode(t, -50000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0xc1, 0xd8, 0x7f})

	// June 24, 2019, 17:53:04.180
	assertEncodeDecode(t, 2019, 6, 24, 17, 53, 4, 180000000, "", []byte{0x11, 0x75, 0xc4, 0x46, 0x0b, 0x4d})

	// January 7, 1998, 08:19:20, Europe/Rome
	assertEncodeDecode(t, 1998, 1, 7, 8, 19, 20, 0, "Europe/Rome", []byte{0x50, 0x13, 0x3a, 0x01, 0x06, 0x0c, 'E', '/', 'R', 'o', 'm', 'e'})

	// August 31, 3190, 00:54:47.394129, location 59.94, 10.71
	// TEST_TIMESTAMP_TZ_LOC(, 3190,8,31,0,54,47,394129000, 5994, 1071, {0xbe, 0x36, 0xf8, 0x18, 0x39, 0x60, 0xa5, 0x18, 0xd5, 0xae, 0x17, 0x02})
}

package compact_time

import (
	"bytes"
	"testing"
	"time"
)

func assertDateEncodeDecode(t *testing.T, year int, month int, day int, expected []byte) {
	actual := make([]byte, len(expected))
	location := time.UTC
	hour := 0
	minute := 0
	second := 0
	nanosecond := 0
	expectedDate := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, location)
	actualSize := DateEncodedSize(expectedDate)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
	}
	encodedCount, ok := EncodeDate(expectedDate, actual)
	if !ok {
		t.Errorf("Not enough room to encode date %v", expectedDate)
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", expected, actual)
	}

	actualDate, decodedCount, ok := DecodeDate(expected)
	if !ok {
		t.Errorf("Not enough bytes to decode date")
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
	}
	if !actualDate.Equal(expectedDate) {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedDate, actualDate)
	}
}

func assertTimeEncodeDecode(t *testing.T, hour int, minute int, second int, nanosecond int, timezone string, expected []byte) {
	year := 0
	month := 1
	day := 1
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
	actualSize := TimeEncodedSize(expectedDate)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
	}
	encodedCount, ok, err := EncodeTime(expectedDate, actual)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if !ok {
		t.Errorf("Not enough room to encode %v", expectedDate)
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", expected, actual)
	}

	actualDate, decodedCount, ok, err := DecodeTime(expected)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if !ok {
		t.Errorf("Not enough data to decode time")
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
	}
	if !actualDate.Equal(expectedDate) {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedDate, actualDate)
	}
}

func assertTimestampEncodeDecode(t *testing.T, year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone string, expected []byte) {
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
	actualSize := TimestampEncodedSize(expectedDate)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
	}
	encodedCount, ok, err := EncodeTimestamp(expectedDate, actual)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if !ok {
		t.Errorf("Not enough room to encode timestamp %v", expectedDate)
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", expected, actual)
	}

	actualDate, decodedCount, ok, err := DecodeTimestamp(expected)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	if !ok {
		t.Errorf("Not enough data to decode timestamp")
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
	}
	if !actualDate.Equal(expectedDate) {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedDate, actualDate)
	}
}

func TestDate(t *testing.T) {
	assertDateEncodeDecode(t, 2000, 1, 1, []byte{0x21, 0x00, 0x00})
	assertDateEncodeDecode(t, -2000, 12, 21, []byte{0x95, 0x7d, 0x3f})
}

func TestTime(t *testing.T) {
	assertTimeEncodeDecode(t, 8, 41, 05, 999999999, "", []byte{0x47, 0x69, 0xf1, 0x9f, 0xac, 0xb9, 0x03})
	assertTimeEncodeDecode(t, 14, 18, 30, 43000000, "", []byte{0x73, 0x92, 0xb7, 0x02})
	assertTimeEncodeDecode(t, 23, 6, 55, 8000, "", []byte{0xbd, 0xc6, 0x8d, 0x00, 0x00})
	assertTimeEncodeDecode(t, 10, 10, 10, 0, "Asia/Tokyo", []byte{0x50, 0x8a, 0x02, 0x1c, 'S', '/', 'T', 'o', 'k', 'y', 'o'})
}

func TestTimestamp(t *testing.T) {
	assertTimestampEncodeDecode(t, 2020, 8, 30, 15, 33, 14, 19577323, "Asia/Singapore", []byte{0x3b, 0xe1, 0xf3, 0xb8, 0x9e, 0xab, 0x12, 0x00, 0x50, 0x2c, 'S', '/', 'S', 'i', 'n', 'g', 'a', 'p', 'o', 'r', 'e'})
	assertTimestampEncodeDecode(t, 1966, 12, 1, 5, 13, 5, 0, "", []byte{0x14, 0x4d, 0x09, 0x1c, 0x07})

	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "Europe/Berlin", []byte{0x00, 0x00, 0x08, 0x01, 00, 0x20, 'E', '/', 'B', 'e', 'r', 'l', 'i', 'n'})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0x01, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 1, 0, 0, 0, "", []byte{0x00, 0x40, 0x08, 0x01, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 1, 0, 0, "", []byte{0x00, 0x01, 0x08, 0x01, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 1, 0, "", []byte{0x04, 0x00, 0x08, 0x01, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 1000000, "", []byte{0x01, 0x00, 0x08, 0x11, 0x00, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000000, "", []byte{0x01, 0x00, 0x08, 0x71, 0x3e, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000, "", []byte{0x02, 0x00, 0x08, 0x71, 0x3e, 0x00, 0x01})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999, "", []byte{0x03, 0x00, 0x08, 0x71, 0x3e, 0x00, 0x00, 0x00, 0x01})
	assertTimestampEncodeDecode(t, 2009, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0x01, 0x25})
	assertTimestampEncodeDecode(t, 3009, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0x01, 0x9f, 0x45})
	assertTimestampEncodeDecode(t, -50000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x08, 0xc1, 0xd8, 0x7f})

	// June 24, 2019, 17:53:04.180
	assertTimestampEncodeDecode(t, 2019, 6, 24, 17, 53, 4, 180000000, "", []byte{0x11, 0x75, 0xc4, 0x46, 0x0b, 0x4d})

	// January 7, 1998, 08:19:20, Europe/Rome
	assertTimestampEncodeDecode(t, 1998, 1, 7, 8, 19, 20, 0, "Europe/Rome", []byte{0x50, 0x13, 0x3a, 0x01, 0x06, 0x18, 'E', '/', 'R', 'o', 'm', 'e'})

	// August 31, 3190, 00:54:47.394129, location 59.94, 10.71
	// TEST_TIMESTAMP_TZ_LOC(, 3190,8,31,0,54,47,394129000, 5994, 1071, {0xbe, 0x36, 0xf8, 0x18, 0x39, 0x60, 0xa5, 0x18, 0xd5, 0xae, 0x17, 0x02})
}

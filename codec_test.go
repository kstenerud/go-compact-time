// Copyright 2019 Karl Stenerud
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package compact_time

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/kstenerud/go-describe"
)

func assertDateEncodeDecode(t *testing.T, year int, month int, day int, expected []byte) {
	actual := make([]byte, len(expected))
	expectedDate, err := NewDate(year, month, day)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	actualSize := EncodedSize(expectedDate)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
		return
	}
	encodedCount, ok := Encode(expectedDate, actual)
	if !ok {
		t.Errorf("Not enough room to encode date %v at %v", expectedDate, encodedCount)
		return
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
		return
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", describe.D(expected), describe.D(actual))
		return
	}

	actualDate, decodedCount, err := DecodeDate(expected)
	if err != nil {
		t.Error(err)
		return
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
		return
	}
	if *actualDate != *expectedDate {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedDate, actualDate)
		return
	}
}

func assertTimeEncodeDecode(t *testing.T, hour int, minute int, second int, nanosecond int, timezone string, expected []byte) {
	actual := make([]byte, len(expected))
	if len(timezone) > 0 {
		_, err := time.LoadLocation(timezone)
		if err != nil {
			t.Errorf("BUG IN TEST CODE. Error loading location %v: %v", timezone, err)
			return
		}
	}
	expectedTime, err := NewTime(hour, minute, second, nanosecond, timezone)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	actualSize := EncodedSize(expectedTime)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
		return
	}
	encodedCount, ok := Encode(expectedTime, actual)
	if !ok {
		t.Errorf("Not enough room to encode %v at %v", expectedTime, encodedCount)
		return
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
		return
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", describe.D(expected), describe.D(actual))
		return
	}

	actualTime, decodedCount, err := DecodeTime(expected)
	if err != nil {
		t.Error(err)
		return
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
		return
	}
	if *actualTime != *expectedTime {
		t.Errorf("Expected decoded date of %v but got %v", expectedTime, actualTime)
		return
	}
}

func assertTimestampEncode(t *testing.T, year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone string, expected []byte) {
	actual := make([]byte, len(expected))
	if len(timezone) > 0 && timezone != "Z" && timezone != "Zero" && timezone != "L" && timezone != "Local" {
		_, err := time.LoadLocation(timezone)
		if err != nil {
			t.Errorf("BUG IN TEST CODE. Error loading location %v: %v", timezone, err)
			return
		}
	}
	expectedTimestamp, err := NewTimestamp(year, month, day, hour, minute, second, nanosecond, timezone)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	actualSize := EncodedSize(expectedTimestamp)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
		return
	}
	encodedCount, ok := Encode(expectedTimestamp, actual)
	if !ok {
		t.Errorf("Not enough room to encode timestamp %v at %v", expectedTimestamp, encodedCount)
		return
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
		return
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", describe.D(expected), describe.D(actual))
		return
	}
}

func assertTimestampEncodeDecode(t *testing.T, year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone string, expected []byte) {
	actual := make([]byte, len(expected))
	if len(timezone) > 0 && timezone != "Z" && timezone != "Zero" && timezone != "L" && timezone != "Local" {
		_, err := time.LoadLocation(timezone)
		if err != nil {
			t.Errorf("BUG IN TEST CODE. Error loading location %v: %v", timezone, err)
			return
		}
	}
	expectedTimestamp, err := NewTimestamp(year, month, day, hour, minute, second, nanosecond, timezone)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	actualSize := EncodedSize(expectedTimestamp)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
		return
	}
	encodedCount, ok := Encode(expectedTimestamp, actual)
	if !ok {
		t.Errorf("Not enough room to encode timestamp %v at %v", expectedTimestamp, encodedCount)
		return
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
		return
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", describe.D(expected), describe.D(actual))
		return
	}

	actualTimestamp, decodedCount, err := DecodeTimestamp(expected)
	if err != nil {
		t.Error(err)
		return
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
		return
	}
	if *actualTimestamp != *expectedTimestamp {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedTimestamp, actualTimestamp)
		return
	}
}

func assertTimestampLatLongEncodeDecode(t *testing.T, year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int, expected []byte) {
	actual := make([]byte, len(expected))
	expectedTimestamp, err := NewTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	actualSize := encodedSizeTimestamp(expectedTimestamp)
	if actualSize != len(expected) {
		t.Errorf("Expected encoded size of %v but got %v", len(expected), actualSize)
		return
	}
	encodedCount, ok := encodeTimestamp(expectedTimestamp, actual)
	if !ok {
		t.Errorf("Not enough room to encode timestamp %v at %v", expectedTimestamp, encodedCount)
		return
	}
	if encodedCount != len(expected) {
		t.Errorf("Expected encoded byte count of %v but got %v", len(expected), encodedCount)
		return
	}
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected encoded bytes %v but got %v", describe.D(expected), describe.D(actual))
		return
	}

	actualTimestamp, decodedCount, err := DecodeTimestamp(expected)
	if err != nil {
		t.Error(err)
		return
	}
	if decodedCount != len(expected) {
		t.Errorf("Expected decoded byte count of %v but got %v", len(expected), decodedCount)
		return
	}
	if *actualTimestamp != *expectedTimestamp {
		t.Errorf("Expected decoded date of [%v] but got [%v]", expectedTimestamp, actualTimestamp)
		return
	}
}

func TestDate(t *testing.T) {
	assertDateEncodeDecode(t, 2000, 1, 1, []byte{0x21, 0x00, 0x00})
	assertDateEncodeDecode(t, -2000, 12, 21, []byte{0x95, 0x7f, 0x3e})
}

func TestTime(t *testing.T) {
	assertTimeEncodeDecode(t, 8, 41, 05, 999999999, "", []byte{0xfe, 0x4f, 0xd6, 0xdc, 0x8b, 0x14, 0x01})
	assertTimeEncodeDecode(t, 14, 18, 30, 43000000, "", []byte{0x5a, 0xc1, 0x93, 0x1c})
	assertTimeEncodeDecode(t, 23, 6, 55, 8000, "", []byte{0x44, 0x00, 0x80, 0xdb, 0xb8})
	assertTimeEncodeDecode(t, 10, 10, 10, 0, "Asia/Tokyo", []byte{0x51, 0x14, 0x05, 0x0e, 'S', '/', 'T', 'o', 'k', 'y', 'o'})
}

func TestTimestamp(t *testing.T) {
	assertTimestampEncodeDecode(t, 2020, 8, 30, 15, 33, 14, 19577323, "Asia/Singapore", []byte{0x5f, 0xcf, 0x55, 0x09, 0x9c, 0xf0, 0x79, 0x44, 0x01, 0x16, 'S', '/', 'S', 'i', 'n', 'g', 'a', 'p', 'o', 'r', 'e'})
	assertTimestampEncodeDecode(t, 1966, 12, 1, 5, 13, 5, 0, "", []byte{0x28, 0x9a, 0x12, 0x78, 0x08})

	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "Europe/Berlin", []byte{0x01, 0x00, 0x10, 0x02, 00, 0x10, 'E', '/', 'B', 'e', 'r', 'l', 'i', 'n'})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 1, 0, 0, 0, "", []byte{0x00, 0x80, 0x10, 0x02, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 1, 0, 0, "", []byte{0x00, 0x02, 0x10, 0x02, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 1, 0, "", []byte{0x08, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 1000000, "", []byte{0x0a, 0x00, 0x00, 0x40, 0x08, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000000, "", []byte{0x3a, 0x1f, 0x00, 0x40, 0x08, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999000, "", []byte{0x3c, 0x1f, 0x00, 0x00, 0x00, 0x21, 0x00, 0x00})
	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 999, "", []byte{0x3e, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x84, 0x00, 0x00})
	assertTimestampEncodeDecode(t, 2009, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x10, 0x42, 0x02})

	assertTimestampEncodeDecode(t, 3009, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x10, 0x42, 0xfc, 0x01})
	assertTimestampEncodeDecode(t, -50000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x10, 0xe2, 0xc7, 0x65})

	// June 24, 2019, 17:53:04.180
	assertTimestampEncodeDecode(t, 2019, 6, 24, 17, 53, 4, 180000000, "", []byte{0xa2, 0x85, 0xa8, 0x23, 0x36, 0x13})

	// January 7, 1998, 08:19:20, Europe/Rome
	assertTimestampEncodeDecode(t, 1998, 1, 7, 8, 19, 20, 0, "Europe/Rome", []byte{0xa1, 0x26, 0x74, 0x62, 0x00, 0x0c, 'E', '/', 'R', 'o', 'm', 'e'})

	// August 31, 3190, 00:54:47.394129, location 59.94, 10.71
	assertTimestampLatLongEncodeDecode(t, 3190, 8, 31, 0, 54, 47, 394129000, 5994, 1071, []byte{0x8d, 0x1c, 0xb0, 0xd7, 0x06, 0x1f, 0x99, 0x12, 0xd5, 0x2e, 0x2f, 0x04})

	assertTimestampEncodeDecode(t, 2000, 1, 1, 0, 0, 0, 0, "Local", []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
}

func TestTimestampUTC(t *testing.T) {
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/GMT", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/GMT+0", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/GMT-0", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/GMT0", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/Greenwich", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/UCT", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/Universal", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/UTC", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Etc/Zulu", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Factory", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "GMT", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "GMT+0", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "GMT-0", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "GMT0", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Greenwich", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "UCT", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Universal", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "UTC", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Zulu", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Z", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Zero", []byte{0x00, 0x00, 0x10, 0x02, 0x00})
}

func TestTimestampLocal(t *testing.T) {
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "Local", []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
	assertTimestampEncode(t, 2000, 1, 1, 0, 0, 0, 0, "L", []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
}

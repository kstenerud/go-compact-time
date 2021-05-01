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
	gotime "time"

	"github.com/kstenerud/go-describe"
)

func isEqualGoTime(a gotime.Time, b gotime.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day() &&
		a.Hour() == b.Hour() &&
		a.Minute() == b.Minute() &&
		a.Second() == b.Second() &&
		a.Location().String() == b.Location().String()
}

func getGoTZ(tz Timezone) *gotime.Location {
	switch tz.Type {
	case TimezoneTypeAreaLocation:
		var err error
		goTZ, err := gotime.LoadLocation(tz.LongAreaLocation)
		if err != nil {
			panic(fmt.Errorf("BUG IN TEST CODE. Error loading location %v: %v", tz.LongAreaLocation, err))
		}
		return goTZ
	case TimezoneTypeLatitudeLongitude:
		return nil
	case TimezoneTypeLocal:
		return gotime.Local
	case TimezoneTypeUTC:
		return gotime.UTC
	case TimezoneTypeUTCOffset:
		return gotime.FixedZone("", int(tz.MinutesOffsetFromUTC)*60)
	default:
		panic(fmt.Errorf("%v: Unknown time zone type", tz.Type))
	}
}

func assertEncodeDecode(t *testing.T, expectedTime Time, expectValid bool, expectedBytes []byte) {

	// Validation
	if expectValid {
		if err := expectedTime.Validate(); err != nil {
			t.Errorf("Error validating expected time %v: %v", expectedTime, err)
		}
	}

	// Go time conversion
	if !expectedTime.IsZeroValue() && (expectedTime.Type == TimeTypeTime || expectedTime.Type == TimeTypeTimestamp) {
		goTZ := getGoTZ(expectedTime.Timezone)
		if goTZ != nil {
			expectedGoTime := gotime.Date(expectedTime.Year, gotime.Month(expectedTime.Month),
				int(expectedTime.Day), int(expectedTime.Hour), int(expectedTime.Minute),
				int(expectedTime.Second), int(expectedTime.Nanosecond), goTZ)
			actualGoTime, err := expectedTime.AsGoTime()
			if err != nil {
				t.Errorf("Error converting %v to go date", expectedTime)
				return
			}
			if !isEqualGoTime(expectedGoTime, actualGoTime) {
				t.Errorf("Expected %v to convert to go time %v but got %v", expectedTime, expectedGoTime, actualGoTime)
				return
			}
		}
	}

	// Encoded size
	expectedSize := len(expectedBytes)
	actualSize := expectedTime.EncodedSize()
	if actualSize != expectedSize {
		t.Errorf("Expected %v to have encoded size of %v but got %v", expectedTime, expectedSize, actualSize)
		// Fallthrough so that we see the actual bytes encoded later
	}

	// Encoding
	actualBytes := &bytes.Buffer{}
	encodedCount, err := expectedTime.Encode(actualBytes)
	if err != nil {
		t.Errorf("Error encoding %v: %v", expectedTime, err)
		return
	}
	if encodedCount != len(expectedBytes) {
		t.Errorf("Expected %v to have encoded byte count of %v but got %v",
			expectedTime, expectedSize, encodedCount)
		// Fallthrough so that we see the actual bytes encoded later
	}
	if !bytes.Equal(expectedBytes, actualBytes.Bytes()) {
		t.Errorf("Expected %v to encode to %v but got %v", expectedTime,
			describe.D(expectedBytes), describe.D(actualBytes.Bytes()))
		return
	}

	// Decoding
	var actualTime Time
	var decodedCount int
	switch expectedTime.Type {
	case TimeTypeDate:
		actualTime, decodedCount, err = DecodeDate(bytes.NewBuffer(expectedBytes))
	case TimeTypeTime:
		actualTime, decodedCount, err = DecodeTime(bytes.NewBuffer(expectedBytes))
	case TimeTypeTimestamp:
		actualTime, decodedCount, err = DecodeTimestamp(bytes.NewBuffer(expectedBytes))
	}
	if err != nil {
		t.Errorf("Error attempting to decode %v to %v: %v",
			describe.D(expectedBytes), expectedTime, err)
		return
	}
	if decodedCount != len(expectedBytes) {
		t.Errorf("Expected %v (%v) to have decoded byte count of %v but got %v",
			expectedTime, describe.D(expectedBytes), len(expectedBytes), decodedCount)
		return
	}
	if !expectedTime.IsEquivalentTo(actualTime) {
		t.Errorf("Expected %v to decode to date of %v but got %v",
			describe.D(expectedBytes), expectedTime, actualTime)
		return
	}
	if expectValid {
		if err = actualTime.Validate(); err != nil {
			t.Errorf("Error validating expected time %v (decoded from %v): %v",
				expectedTime, describe.D(expectedBytes), err)
		}
	}
}

func TestDate(t *testing.T) {
	assertEncodeDecode(t, NewDate(2000, 1, 1), true, []byte{0x21, 0x00, 0x00})
	assertEncodeDecode(t, NewDate(2001, 1, 1), true, []byte{0x21, 0x04, 0x00})
	assertEncodeDecode(t, NewDate(-2000, 12, 21), true, []byte{0x95, 0x7f, 0x3e})
}

func TestTime(t *testing.T) {
	assertEncodeDecode(t, NewTime(8, 41, 05, 999999999, TZAtUTC()), true, []byte{0xfe, 0x4f, 0xd6, 0xdc, 0x8b, 0x14, 0xfd})
	assertEncodeDecode(t, NewTime(14, 18, 30, 43000000, TZAtUTC()), true, []byte{0x5a, 0xc1, 0x93, 0xdc})
	assertEncodeDecode(t, NewTime(23, 6, 55, 8000, TZAtUTC()), true, []byte{0x44, 0x00, 0x80, 0xdb, 0xb8})
	assertEncodeDecode(t, NewTime(10, 10, 10, 0, TZAtAreaLocation("Asia/Tokyo")), true, []byte{0x51, 0x14, 0xf5, 0x0e, 'S', '/', 'T', 'o', 'k', 'y', 'o'})
	assertEncodeDecode(t, NewTime(8, 41, 05, 999999999, TZWithMiutesOffsetFromUTC(1000)), true, []byte{0xff, 0x4f, 0xd6, 0xdc, 0x8b, 0x14, 0xfd, 0x00, 0xe8, 0x03})
	assertEncodeDecode(t, NewTime(8, 41, 05, 999999999, TZWithMiutesOffsetFromUTC(-500)), true, []byte{0xff, 0x4f, 0xd6, 0xdc, 0x8b, 0x14, 0xfd, 0x00, 0x0c, 0xfe})
}

func TestTimestamp(t *testing.T) {
	assertEncodeDecode(t, NewTimestamp(2020, 8, 30, 15, 33, 14, 19577323, TZAtAreaLocation("Asia/Singapore")), true, []byte{0x5f, 0xcf, 0x55, 0x09, 0x9c, 0xf0, 0x79, 0x44, 0x01, 0x16, 'S', '/', 'S', 'i', 'n', 'g', 'a', 'p', 'o', 'r', 'e'})
	assertEncodeDecode(t, NewTimestamp(1966, 12, 1, 5, 13, 5, 0, TZAtUTC()), true, []byte{0x28, 0x9a, 0x12, 0x78, 0x08})

	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Europe/Berlin")), true, []byte{0x01, 0x00, 0x10, 0x02, 00, 0x10, 'E', '/', 'B', 'e', 'r', 'l', 'i', 'n'})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtUTC()), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 1, 0, 0, 0, TZAtUTC()), true, []byte{0x00, 0x80, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 1, 0, 0, TZAtUTC()), true, []byte{0x00, 0x02, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 1, 0, TZAtUTC()), true, []byte{0x08, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 1000000, TZAtUTC()), true, []byte{0x0a, 0x00, 0x00, 0x40, 0x08, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 999000000, TZAtUTC()), true, []byte{0x3a, 0x1f, 0x00, 0x40, 0x08, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 999000, TZAtUTC()), true, []byte{0x3c, 0x1f, 0x00, 0x00, 0x00, 0x21, 0x00, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 999, TZAtUTC()), true, []byte{0x3e, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x84, 0x00, 0x00})
	assertEncodeDecode(t, NewTimestamp(2009, 1, 1, 0, 0, 0, 0, TZAtUTC()), true, []byte{0x00, 0x00, 0x10, 0x42, 0x02})

	assertEncodeDecode(t, NewTimestamp(3009, 1, 1, 0, 0, 0, 0, TZAtUTC()), true, []byte{0x00, 0x00, 0x10, 0x42, 0xfc, 0x01})
	assertEncodeDecode(t, NewTimestamp(-50000, 1, 1, 0, 0, 0, 0, TZAtUTC()), true, []byte{0x00, 0x00, 0x10, 0xe2, 0xc7, 0x65})

	// June 24, 2019, 17:53:04.180
	assertEncodeDecode(t, NewTimestamp(2019, 6, 24, 17, 53, 4, 180000000, TZAtUTC()), true, []byte{0xa2, 0x85, 0xa8, 0x23, 0x36, 0x13})

	// January 7, 1998, 08:19:20, Europe/Rome
	assertEncodeDecode(t, NewTimestamp(1998, 1, 7, 8, 19, 20, 0, TZAtAreaLocation("Europe/Rome")), true, []byte{0xa1, 0x26, 0x74, 0x62, 0x00, 0x0c, 'E', '/', 'R', 'o', 'm', 'e'})

	// August 31, 3190, 00:54:47.394129, location 59.94, 10.71
	assertEncodeDecode(t, NewTimestamp(3190, 8, 31, 0, 54, 47, 394129000, TZAtLatLong(5994, 1071)), true, []byte{0x8d, 0x1c, 0xb0, 0xd7, 0x06, 0x1f, 0x99, 0x12, 0xd5, 0x2e, 0x2f, 0x04})

	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZLocal()), true, []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZWithMiutesOffsetFromUTC(1000)), true, []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x00, 0xe8, 0x03})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZWithMiutesOffsetFromUTC(-60)), true, []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x00, 0xc4, 0xff})
}

func TestTimestampUTC(t *testing.T) {
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtUTC()), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/GMT")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/GMT+0")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/GMT-0")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/GMT0")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/Greenwich")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/UCT")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/Universal")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/UTC")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Etc/Zulu")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Factory")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("GMT")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("GMT+0")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("GMT-0")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("GMT0")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Greenwich")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("UCT")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Universal")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("UTC")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Zulu")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Z")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Zero")), true, []byte{0x00, 0x00, 0x10, 0x02, 0x00})
}

func TestTimestampLocal(t *testing.T) {
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZLocal()), true, []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("Local")), true, []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
	assertEncodeDecode(t, NewTimestamp(2000, 1, 1, 0, 0, 0, 0, TZAtAreaLocation("L")), true, []byte{0x01, 0x00, 0x10, 0x02, 0x00, 0x02, 0x4c})
}

func TestZeroValues(t *testing.T) {
	assertEncodeDecode(t, ZeroDate(), false, []byte{0x00, 0x00, 0x00})
	assertEncodeDecode(t, ZeroTime(), false, []byte{0x00, 0x00, 0x00})
	assertEncodeDecode(t, ZeroTimestamp(), false, []byte{0x00, 0x00, 0x00, 0x00, 0x00})
}

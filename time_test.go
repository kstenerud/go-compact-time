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
	"fmt"
	"testing"
)

func newDate(year, month, day int) Time {
	v, err := NewDate(year, month, day)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	return v
}

func newTime(hour, minute, second, nanosecond int, areaLocation string) Time {
	v, err := NewTime(hour, minute, second, nanosecond, areaLocation)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	return v
}

func newTimeLL(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) Time {
	v, err := NewTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	return v
}

func newTimestamp(year, month, day, hour, minute, second, nanosecond int, areaLocation string) Time {
	v, err := NewTimestamp(year, month, day, hour, minute, second, nanosecond, areaLocation)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	return v
}

func newTimestampLL(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) Time {
	v, err := NewTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	if err != nil {
		panic(fmt.Errorf("BUG: Unexpected error %v", err))
	}
	return v
}

func assertEquivalentTime(t *testing.T, a, b Time) {
	if !a.IsEquivalentTo(&b) {
		t.Errorf("Expected %v to be equivalent to %v", a, b)
	}
}

func assertNotEquivalentTime(t *testing.T, a, b Time) {
	if a.IsEquivalentTo(&b) {
		t.Errorf("Expected %v to not be equivalent to %v", a, b)
	}
}

func TestEquivalence(t *testing.T) {
	assertEquivalentTime(t, newDate(2100, 1, 1), newDate(2100, 1, 1))
	assertNotEquivalentTime(t, newDate(2100, 1, 2), newDate(2100, 1, 1))
	assertNotEquivalentTime(t, newDate(2100, 6, 1), newDate(2100, 1, 1))
	assertNotEquivalentTime(t, newDate(2101, 1, 1), newDate(2100, 1, 1))

	assertEquivalentTime(t, newTime(12, 1, 4, 91, ""), newTime(12, 1, 4, 91, ""))
	assertEquivalentTime(t, newTime(12, 1, 4, 91, ""), newTime(12, 1, 4, 91, "UTC"))
	assertEquivalentTime(t, newTime(12, 1, 4, 91, "Etc/UTC"), newTime(12, 1, 4, 91, "UTC"))
	assertEquivalentTime(t, newTime(12, 1, 4, 91, "Etc/UTC"), newTime(12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTime(12, 1, 4, 9, ""), newTime(12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTime(12, 1, 14, 91, ""), newTime(12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTime(12, 12, 4, 91, ""), newTime(12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTime(1, 1, 4, 91, ""), newTime(12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTime(12, 1, 4, 91, "L"), newTime(12, 1, 4, 91, ""))

	assertEquivalentTime(t, newTimeLL(12, 1, 4, 91, 0, 0), newTimeLL(12, 1, 4, 91, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(12, 1, 4, 91, 0, 1), newTimeLL(12, 1, 4, 91, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(12, 1, 4, 91, 1, 0), newTimeLL(12, 1, 4, 91, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(12, 1, 4, 910, 0, 0), newTimeLL(12, 1, 4, 91, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(12, 1, 14, 91, 0, 0), newTimeLL(12, 1, 4, 91, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(12, 11, 4, 91, 0, 0), newTimeLL(12, 1, 4, 91, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(11, 1, 4, 91, 0, 0), newTimeLL(12, 1, 4, 91, 0, 0))

	assertEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 1, 4, 91, "Z"), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 1, 4, 91, "Etc/GMT"), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 1, 4, 91, "America/Los_Angeles"), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 1, 4, 9, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 1, 1, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 8, 5, 12, 2, 4, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 8, 5, 1, 1, 4, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 8, 4, 12, 1, 4, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2050, 3, 5, 12, 1, 4, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))
	assertNotEquivalentTime(t, newTimestamp(2051, 8, 5, 12, 1, 4, 91, ""), newTimestamp(2050, 8, 5, 12, 1, 4, 91, ""))

	assertEquivalentTime(t, newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 2), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 2, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 5, 12, 1, 4, 92, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 5, 12, 1, 2, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 5, 12, 2, 4, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 5, 2, 1, 4, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 8, 2, 12, 1, 4, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2050, 2, 5, 12, 1, 4, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))
	assertNotEquivalentTime(t, newTimestampLL(2052, 8, 5, 12, 1, 4, 91, 1, 1), newTimestampLL(2050, 8, 5, 12, 1, 4, 91, 1, 1))

	assertNotEquivalentTime(t, newDate(2100, 1, 2), newTime(12, 4, 1, 0, ""))
	assertNotEquivalentTime(t, newDate(2100, 1, 2), newTimeLL(12, 5, 1, 0, 1, 1))
	assertNotEquivalentTime(t, newDate(2100, 1, 2), newTimestamp(2100, 1, 2, 0, 0, 0, 0, ""))
	assertNotEquivalentTime(t, newDate(2100, 1, 2), newTimestampLL(2100, 1, 2, 0, 0, 0, 0, 0, 0))
	assertNotEquivalentTime(t, newTime(12, 4, 1, 0, ""), newTimeLL(12, 4, 1, 0, 0, 0))
	assertNotEquivalentTime(t, newTime(12, 4, 1, 0, ""), newTimestamp(2100, 1, 2, 0, 0, 0, 0, ""))
	assertNotEquivalentTime(t, newTime(12, 4, 1, 0, ""), newTimestampLL(2100, 1, 2, 0, 0, 0, 0, 0, 0))
	assertNotEquivalentTime(t, newTimeLL(12, 5, 1, 0, 1, 1), newTimestamp(2100, 1, 2, 0, 0, 0, 0, ""))
	assertNotEquivalentTime(t, newTimeLL(12, 5, 1, 0, 1, 1), newTimestampLL(2100, 1, 2, 0, 0, 0, 0, 0, 0))
	assertNotEquivalentTime(t, newTimestamp(2100, 1, 2, 0, 0, 0, 0, ""), newTimestampLL(2100, 1, 2, 0, 0, 0, 0, 0, 0))
}

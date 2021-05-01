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
	"testing"
)

func assertEquivalentTime(t *testing.T, a, b Time) {
	if !a.IsEquivalentTo(b) {
		t.Errorf("Expected %v to be equivalent to %v", a, b)
	}
}

func assertNotEquivalentTime(t *testing.T, a, b Time) {
	if a.IsEquivalentTo(b) {
		t.Errorf("Expected %v to not be equivalent to %v", a, b)
	}
}

func TestEquivalence(t *testing.T) {
	assertEquivalentTime(t, NewDate(2100, 1, 1), NewDate(2100, 1, 1))
	assertNotEquivalentTime(t, NewDate(2100, 1, 2), NewDate(2100, 1, 1))
	assertNotEquivalentTime(t, NewDate(2100, 6, 1), NewDate(2100, 1, 1))
	assertNotEquivalentTime(t, NewDate(2101, 1, 1), NewDate(2100, 1, 1))

	assertEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtUTC()), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtUTC()), NewTime(12, 1, 4, 91, TZAtAreaLocation("UTC")))
	assertEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtAreaLocation("Etc/UTC")), NewTime(12, 1, 4, 91, TZAtAreaLocation("UTC")))
	assertEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtAreaLocation("Etc/UTC")), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertEquivalentTime(t, NewTime(12, 1, 4, 91, TZWithMiutesOffsetFromUTC(1)), NewTime(12, 1, 4, 91, TZWithMiutesOffsetFromUTC(1)))
	assertNotEquivalentTime(t, NewTime(12, 1, 4, 9, TZAtUTC()), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(12, 1, 14, 91, TZAtUTC()), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(12, 12, 4, 91, TZAtUTC()), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(1, 1, 4, 91, TZAtUTC()), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(12, 1, 4, 91, TZLocal()), NewTime(12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(12, 1, 4, 91, TZWithMiutesOffsetFromUTC(1)), NewTime(12, 1, 4, 91, TZWithMiutesOffsetFromUTC(0)))

	assertEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtLatLong(0, 1)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 1, 4, 91, TZAtLatLong(1, 0)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 1, 4, 910, TZAtLatLong(0, 0)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 1, 14, 91, TZAtLatLong(0, 0)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 11, 4, 91, TZAtLatLong(0, 0)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(11, 1, 4, 91, TZAtLatLong(0, 0)), NewTime(12, 1, 4, 91, TZAtLatLong(0, 0)))

	assertEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtAreaLocation("Z")), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtAreaLocation("Etc/GMT")), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZWithMiutesOffsetFromUTC(-100)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZWithMiutesOffsetFromUTC(-100)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtAreaLocation("America/Los_Angeles")), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 9, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 1, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 2, 4, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 1, 1, 4, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 4, 12, 1, 4, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 3, 5, 12, 1, 4, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2051, 8, 5, 12, 1, 4, 91, TZAtUTC()), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtUTC()))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZWithMiutesOffsetFromUTC(-300)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZWithMiutesOffsetFromUTC(300)))

	assertEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 2)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(2, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 4, 92, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 1, 2, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 12, 2, 4, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 5, 2, 1, 4, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 8, 2, 12, 1, 4, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2050, 2, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewTimestamp(2052, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)), NewTimestamp(2050, 8, 5, 12, 1, 4, 91, TZAtLatLong(1, 1)))

	assertNotEquivalentTime(t, NewDate(2100, 1, 2), NewTime(12, 4, 1, 0, TZAtUTC()))
	assertNotEquivalentTime(t, NewDate(2100, 1, 2), NewTime(12, 5, 1, 0, TZAtLatLong(1, 1)))
	assertNotEquivalentTime(t, NewDate(2100, 1, 2), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtUTC()))
	assertNotEquivalentTime(t, NewDate(2100, 1, 2), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 4, 1, 0, TZAtUTC()), NewTime(12, 4, 1, 0, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 4, 1, 0, TZAtUTC()), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(12, 4, 1, 0, TZAtUTC()), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTime(12, 5, 1, 0, TZAtLatLong(1, 1)), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtUTC()))
	assertNotEquivalentTime(t, NewTime(12, 5, 1, 0, TZAtLatLong(1, 1)), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtLatLong(0, 0)))
	assertNotEquivalentTime(t, NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtUTC()), NewTimestamp(2100, 1, 2, 0, 0, 0, 0, TZAtLatLong(0, 0)))
}

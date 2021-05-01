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

func assertStringRep(t *testing.T, time Time, expected string) {
	actual := fmt.Sprintf("%v", time)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestStringRepresentation(t *testing.T) {
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZAtUTC()), "2020-01-15/13:41:00.000599")
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZLocal()), "2020-01-15/13:41:00.000599/Local")
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZAtAreaLocation("America/New_York")), "2020-01-15/13:41:00.000599/America/New_York")
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZAtLatLong(50, -50)), "2020-01-15/13:41:00.000599/0.50/-0.50")
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZAtLatLong(500, -500)), "2020-01-15/13:41:00.000599/5.00/-5.00")
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZWithMiutesOffsetFromUTC(60)), "2020-01-15/13:41:00.000599+0100")
	assertStringRep(t, NewTimestamp(2020, 1, 15, 13, 41, 0, 599000, TZWithMiutesOffsetFromUTC(-1)), "2020-01-15/13:41:00.000599-0001")
}

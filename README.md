Compact Time
============

A go implementation of [compact time](https://github.com/kstenerud/compact-time/blob/master/compact-time-specification.md).



Library Usage
-------------

```golang
func demonstrateEncode() {
	location, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		// TODO: Handle error
	}
	date := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, location)
	buffer := make([]byte, encodedSizeTimestamp(AsCompactTime(date)))
	encodedCount, ok := encodeTimestamp(AsCompactTime(date), buffer)
	if !ok {
		// TODO: Not enough room in buffer to encode
	}
	fmt.Printf("Encoded [%v] into %v bytes: %v\n", date, encodedCount, buffer)
	// Prints: Encoded [2020-08-30 15:33:14.019577323 +0800 +08] into 21 bytes: [59 225 243 184 158 171 18 0 80 22 83 47 83 105 110 103 97 112 111 114 101]
}

func demonstrateDecode() {
	buffer := []byte{0x14, 0x4d, 0x09, 0x1c, 0x07}
	compactDate, decodedCount, ok := DecodeTimestamp(buffer)
	if err := compactDate.Validate(); err != nil {
		// TODO: Handle error
	}
	if !ok {
		// TODO: Not enough bytes in buffer to decode
	}
	date, err := compactDate.AsGoTime()
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into [%v]\n", decodedCount, buffer, date)
	// Prints: Decoded 5 bytes of [20 77 9 28 7] into [1966-12-01 05:13:05 +0000 UTC]
}
```



License
-------

MIT License:

Copyright 2019 Karl Stenerud

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

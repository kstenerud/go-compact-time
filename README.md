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
	goDate := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, location)
	compactDate, err := AsCompactTime(goDate)
	if err != nil {
		// TODO: Handle error
	}
	buffer := make([]byte, EncodedSize(compactDate))
	encodedCount, ok := Encode(compactDate, buffer)
	if !ok {
		// TODO: Not enough room in buffer to encode
	}
	fmt.Printf("Encoded [%v] into %v bytes: %v\n", goDate, encodedCount, buffer)
	// Prints: Encoded [2020-08-30 15:33:14.019577323 +0800 +08] into 21 bytes: [95 207 85 9 156 240 121 68 1 22 83 47 83 105 110 103 97 112 111 114 101]
}

func demonstrateDecode() {
	buffer := []byte{0x28, 0x9a, 0x12, 0x78, 0x08}
	compactDate, decodedCount, err := DecodeTimestamp(buffer)
	if err != nil {
		// TODO: Check if is compact_time.ErrorIncomplete or something else
	}
	goDate, err := compactDate.AsGoTime()
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into [%v]\n", decodedCount, buffer, goDate)
	// Prints: Decoded 5 bytes of [40 154 18 120 8] into [1966-12-01 05:13:05 +0000 UTC]
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

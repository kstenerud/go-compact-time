Compact Date
============

A go implementation of [compact date](https://github.com/kstenerud/compact-date/blob/master/compact-date-specification.md).



Library Usage
-------------

```golang

func demonstrateEncode() {
	date := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, time.UTC)
	buffer := make([]byte, EncodedSize(date))
	encodedCount, err := Encode(date, buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Encoded %v into %v bytes: %v\n", date, encodedCount, buffer)
}
```

Output:

	Encoded 2020-08-30 15:33:14.019577323 +0000 UTC into 9 bytes: [193 42 185 235 58 23 250 0 40]


```golang
func demonstrateDecode() {
	buffer := []byte{0x14, 0x4d, 0x07, 0x10, 0x03}
	decodedCount, date, err := Decode(buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into %v\n", decodedCount, buffer, date)
}
```

Output:

	Decoded 5 bytes of [20 77 7 16 3] into 1998-01-07 08:19:20 +0000 UTC

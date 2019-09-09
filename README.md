Compact Time
============

A go implementation of [compact time](https://github.com/kstenerud/compact-time/blob/master/compact-time-specification.md).



Library Usage
-------------

```golang
func demonstrateEncode() {
	location, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		// TODO: Deal with this
	}
	date := time.Date(2020, time.Month(8), 30, 15, 33, 14, 19577323, location)
	buffer := make([]byte, EncodedSize(date))
	encodedCount, err := EncodeTimestamp(date, buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Encoded [%v] into %v bytes: %v\n", date, encodedCount, buffer)
}
```

Output:

    Encoded [2020-08-30 15:33:14.019577323 +0800 +08] into 21 bytes: [59 225 243 184 158 171 18 0 80 22 83 47 83 105 110 103 97 112 111 114 101]


```golang
func demonstrateDecode() {
	buffer := []byte{0x14, 0x4d, 0x09, 0x1c, 0x07}
	date, decodedCount, err := DecodeTimestamp(buffer)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("Decoded %v bytes of %v into [%v]\n", decodedCount, buffer, date)
}
```

Output:

    Decoded 5 bytes of [20 77 9 28 7] into [1966-12-01 05:13:05 +0000 UTC]

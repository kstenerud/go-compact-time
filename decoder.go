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

// Package compact_time provides encoding and decoding mechanisms for the
// compact time format (https://github.com/kstenerud/compact-time), as well as
// basic conversion functions to/from the go time package.
//
// Basic validation is performed when decoding data, enough to ensure that it
// isn't blatantly wrong (such as invalid area/location values, latitude 500,
// december 54th, etc). However, it does not do more nuanced checks such as on
// which years February 29th is valid, or when leap seconds are allowed. It
// also doesn't check for impossible timestamp values such as
// 2011-03-13/02:10:00/Los_Angeles.
package compact_time

import (
	"fmt"
	"io"

	"github.com/kstenerud/go-uleb128"
)

var ErrorIncomplete = fmt.Errorf("Compact time value is incomplete")

const RequiredBufferSize = 127

// Decode a date.
func DecodeDate(reader io.Reader) (time Time, bytesDecoded int, err error) {
	return DecodeDateWithBuffer(reader, makeRequiredBuffer())
}

func DecodeDateWithBuffer(reader io.Reader, buffer []byte) (time Time, bytesDecoded int, err error) {
	var year int
	var month int
	var day int

	if err = fillSlice(reader, buffer[:2]); err != nil {
		return
	}
	bytesDecoded = 2
	accumulator := int(decode16LE(buffer))
	day = int(accumulator & maskDay)
	accumulator >>= sizeDay
	month = int(accumulator & maskMonth)
	accumulator >>= sizeMonth
	asUint, asBig, byteCount, err := uleb128.DecodeWithByteBuffer(reader, buffer)
	if err != nil {
		return
	}
	bytesDecoded += byteCount
	if asBig != nil {
		err = fmt.Errorf("Year is too big")
		return
	}
	encodedYear := (asUint << 7) | uint64(accumulator)
	if encodedYear > 0xffffffff || byteCount*7+yearLowBitCountDate > 64 {
		err = fmt.Errorf("Year is too big")
		return
	}
	year = decodeYear(uint32(encodedYear))
	if year == 2000 && month == 0 && day == 0 {
		time = ZeroDate()
		return
	}

	time = NewDate(year, month, day)
	return
}

// Decode a time value.
func DecodeTime(reader io.Reader) (time Time, bytesDecoded int, err error) {
	return DecodeTimeWithBuffer(reader, makeRequiredBuffer())
}

func DecodeTimeWithBuffer(reader io.Reader, buffer []byte) (time Time, bytesDecoded int, err error) {
	var hour int
	var minute int
	var second int
	var nanosecond int
	var tz Timezone

	if _, err = reader.Read(buffer[:1]); err != nil {
		return
	}
	header := buffer[0]

	magnitude := int((header >> 1) & maskMagnitude)
	baseByteCount := baseByteCountsTime[magnitude]
	if err = fillSlice(reader, buffer[1:baseByteCount]); err != nil {
		return
	}
	bytesDecoded = baseByteCount

	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubseconds := uint(sizeSubsecond * magnitude)
	maskSubsecond := bitMask(int(sizeSubseconds))

	accumulator := decodeLE(buffer[:baseByteCount], baseByteCount)
	hasTimezone := accumulator&1 == 1
	accumulator >>= 1
	accumulator >>= sizeMagnitude
	nanosecond = int(accumulator&maskSubsecond) * subsecondMultiplier
	accumulator >>= sizeSubseconds
	second = int(accumulator & maskSecond)
	accumulator >>= sizeSecond
	minute = int(accumulator & maskMinute)
	accumulator >>= sizeMinute
	hour = int(accumulator & maskHour)
	accumulator >>= sizeHour

	expectedReservedBits := reservedBitsTime[magnitude]
	if accumulator != expectedReservedBits {
		if accumulator == 0 {
			time = ZeroTime()
		} else {
			err = fmt.Errorf("Expected reserved bits %b but got %b", expectedReservedBits, accumulator)
		}
		return
	}

	if !hasTimezone {
		time.InitTime(hour, minute, second, nanosecond, timezoneUTC)
		return
	}

	var byteCount int
	tz, byteCount, err = decodeTimezone(reader, buffer)
	bytesDecoded += byteCount
	time.InitTime(hour, minute, second, nanosecond, tz)
	return
}

// Decode a timestamp.
func DecodeTimestamp(reader io.Reader) (time Time, bytesDecoded int, err error) {
	return DecodeTimestampWithBuffer(reader, makeRequiredBuffer())
}

func DecodeTimestampWithBuffer(reader io.Reader, buffer []byte) (time Time, bytesDecoded int, err error) {
	var year int
	var month int
	var day int
	var hour int
	var minute int
	var second int
	var nanosecond int
	var tz Timezone

	if _, err = reader.Read(buffer[:1]); err != nil {
		return
	}
	header := buffer[0]

	magnitude := int((header >> 1) & maskMagnitude)
	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubseconds := uint(sizeSubsecond * magnitude)
	maskSubsecond := bitMask(int(sizeSubseconds))
	baseByteCount := baseByteCountsTimestamp[magnitude]
	if err = fillSlice(reader, buffer[1:baseByteCount]); err != nil {
		return
	}
	bytesDecoded = baseByteCount

	accumulator := decodeLE(buffer[:baseByteCount], baseByteCount)
	hasTimezone := accumulator&1 == 1
	accumulator >>= 1
	accumulator >>= sizeMagnitude
	nanosecond = int(accumulator&maskSubsecond) * subsecondMultiplier
	accumulator >>= sizeSubseconds
	second = int(accumulator & maskSecond)
	accumulator >>= sizeSecond
	minute = int(accumulator & maskMinute)
	accumulator >>= sizeMinute
	hour = int(accumulator & maskHour)
	accumulator >>= sizeHour
	day = int(accumulator & maskDay)
	accumulator >>= sizeDay
	month = int(accumulator & maskMonth)
	accumulator >>= sizeMonth

	yearLowBitCount := yearLowBitCountsTimestamp[magnitude]
	asUint, asBig, byteCount, err := uleb128.DecodeWithByteBuffer(reader, buffer)
	if err != nil {
		return
	}
	bytesDecoded += byteCount
	if asBig != nil {
		err = fmt.Errorf("Year is too big")
		return
	}
	encodedYear := (asUint << yearLowBitCount) | uint64(accumulator)
	if encodedYear > 0xffffffff || byteCount*7+yearLowBitCount > 64 {
		err = fmt.Errorf("Year is too big")
		return
	}
	year = decodeYear(uint32(encodedYear))

	if !hasTimezone {
		if year == 2000 && month == 0 && day == 0 {
			time = ZeroTimestamp()
			return
		}
		tz = timezoneUTC
		byteCount = 0
	} else {
		tz, byteCount, err = decodeTimezone(reader, buffer)
	}

	bytesDecoded += byteCount

	time.InitTimestamp(year, month, day, hour, minute, second, nanosecond, tz)
	return
}

// =============================================================================

func makeRequiredBuffer() []byte {
	return make([]byte, RequiredBufferSize)
}

func fillSlice(reader io.Reader, dst []byte) (err error) {
	var bytesRead int
	for len(dst) > 0 {
		if bytesRead, err = reader.Read(dst); err != nil {
			return
		}
		dst = dst[bytesRead:]
	}
	return
}

func decodeLE(src []byte, byteCount int) uint64 {
	accumulator := uint64(0)
	for i := 0; i < byteCount; i++ {
		accumulator |= uint64(src[i]) << (uint(i) * 8)
	}
	return accumulator
}

func decode16LE(src []byte) uint16 {
	return uint16(src[0]) | (uint16(src[1]) << 8)
}

func decode32LE(src []byte) uint32 {
	return uint32(src[0]) | (uint32(src[1]) << 8) |
		(uint32(src[2]) << 16) | (uint32(src[3]) << 24)
}

func decodeZigzag32(value uint32) int32 {
	return int32((value >> 1) ^ -(value & 1))
}

func decodeYear(encodedYear uint32) int {
	return int(decodeZigzag32(uint32(encodedYear))) + yearBias
}

func decodeTimezone(reader io.Reader, buffer []byte) (tz Timezone, bytesDecoded int, err error) {
	if _, err = reader.Read(buffer[:1]); err != nil {
		return
	}
	header := buffer[0]

	if header&maskLatLong != 0 {
		if err = fillSlice(reader, buffer[1:4]); err != nil {
			return
		}
		latLong := decode32LE(buffer)
		bytesDecoded = 4
		longitudeHundredths := int(int32(latLong) >> shiftLongitude)
		latitudeHundredths := int(int32(latLong<<16) >> 17)
		tz.InitWithLatLong(latitudeHundredths, longitudeHundredths)
		return
	}

	stringLength := int(header >> 1)
	if stringLength == 0 {
		if err = fillSlice(reader, buffer[0:2]); err != nil {
			return
		}
		bytesDecoded = 3
		minutesRaw := decode16LE(buffer)
		const maskNegative = 0xf000
		const maskPositive = 0x0fff
		var minutes int16
		if minutesRaw&0x800 != 0 {
			minutes = int16(minutesRaw | maskNegative)
		} else {
			minutes = int16(minutesRaw & maskPositive)
		}
		tz.InitWithMinutesOffsetFromUTC(int(minutes))
		return
	}

	if err = fillSlice(reader, buffer[:stringLength]); err != nil {
		return
	}
	bytesDecoded = stringLength + 1
	if stringLength == 1 {
		// Avoid a string allocation where possible
		switch buffer[0] {
		case 'L':
			tz = TZLocal()
			return
		case 'Z':
			tz = TZAtUTC()
			return
		}
	}
	tz.InitWithAreaLocation(string(buffer[:stringLength]))
	return
}

var reservedBitsTime = [...]uint64{0x0f, 0x03, 0x00, 0x3f}

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
	gotime "time"

	"github.com/kstenerud/go-uleb128"
)

var ErrorIncomplete = fmt.Errorf("Compact time value is incomplete")

// Decode a date.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
func DecodeDate(src []byte) (time Time, bytesDecoded int, err error) {
	var year int
	var month int
	var day int

	year, month, day, bytesDecoded, err = decodeDateFields(src)
	if err != nil {
		return
	}
	if year == 2000 && month == 0 && day == 0 {
		time = ZeroDate()
		return
	}
	time, err = NewDate(year, month, day)
	return
}

// Decode a go date.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
func DecodeGoDate(src []byte) (time gotime.Time, bytesDecoded int, err error) {
	var year int
	var month int
	var day int

	year, month, day, bytesDecoded, err = decodeDateFields(src)
	if err != nil {
		return
	}
	time = gotime.Date(year, gotime.Month(month), day, 0, 0, 0, 0, gotime.UTC)
	return
}

// Decode a time value.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
func DecodeTime(src []byte) (time Time, bytesDecoded int, err error) {
	var hour int
	var minute int
	var second int
	var nanosecond int
	var latitudeHundredths int
	var longitudeHundredths int
	var areaLocation string
	var tzType TimezoneType
	hour, minute, second, nanosecond,
		latitudeHundredths, longitudeHundredths,
		areaLocation, tzType, bytesDecoded, err = decodeTimeFields(src)
	if err != nil {
		return
	}

	switch tzType {
	case TypeZeroValue:
		time = ZeroTime()
	case TypeLatitudeLongitude:
		time, err = NewTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	default:
		time, err = NewTime(hour, minute, second, nanosecond, areaLocation)
	}

	return
}

// Decode a go time value.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
func DecodeGoTime(src []byte) (time gotime.Time, bytesDecoded int, err error) {
	var hour int
	var minute int
	var second int
	var nanosecond int
	var areaLocation string
	var tzType TimezoneType
	hour, minute, second, nanosecond, _, _,
		areaLocation, tzType, bytesDecoded, err = decodeTimeFields(src)
	if err != nil {
		return
	}

	if tzType == TypeLatitudeLongitude {
		err = fmt.Errorf("Go time doesn't support latitude/longitude")
		return
	} else {
		_, longAreaLocation := splitAreaLocation(areaLocation)
		if longAreaLocation == "Etc/UTC" {
			longAreaLocation = "UTC"
		}
		var location *gotime.Location
		location, err = gotime.LoadLocation(longAreaLocation)
		if err != nil {
			return
		}
		time = gotime.Date(0, 0, 0, hour, minute, second, nanosecond, location)
	}

	return
}

// Decode a timestamp.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
func DecodeTimestamp(src []byte) (time Time, bytesDecoded int, err error) {
	var year int
	var month int
	var day int
	var hour int
	var minute int
	var second int
	var nanosecond int
	var latitudeHundredths int
	var longitudeHundredths int
	var areaLocation string
	var tzType TimezoneType

	year, month, day, hour, minute, second, nanosecond,
		latitudeHundredths, longitudeHundredths,
		areaLocation, tzType, bytesDecoded, err = decodeTimestampFields(src)
	if err != nil {
		return
	}

	switch tzType {
	case TypeZeroValue:
		time = ZeroTimestamp()
	case TypeLatitudeLongitude:
		time, err = NewTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	default:
		time, err = NewTimestamp(year, month, day, hour, minute, second, nanosecond, areaLocation)
	}
	return
}

// Decode a go timestamp.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
func DecodeGoTimestamp(src []byte) (time gotime.Time, bytesDecoded int, err error) {
	var year int
	var month int
	var day int
	var hour int
	var minute int
	var second int
	var nanosecond int
	var areaLocation string
	var tzType TimezoneType

	year, month, day, hour, minute, second, nanosecond, _, _,
		areaLocation, tzType, bytesDecoded, err = decodeTimestampFields(src)
	if err != nil {
		return
	}

	if tzType == TypeLatitudeLongitude {
		err = fmt.Errorf("Go time doesn't support latitude/longitude")
		return
	} else {
		_, longAreaLocation := splitAreaLocation(areaLocation)
		if longAreaLocation == "Etc/UTC" {
			longAreaLocation = "UTC"
		}
		var location *gotime.Location
		location, err = gotime.LoadLocation(longAreaLocation)
		if err != nil {
			return
		}
		time = gotime.Date(year, gotime.Month(month), day, hour, minute, second, nanosecond, location)
	}
	return
}

// =============================================================================

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

func decodeTimezone(src []byte) (latitudeHundredths int, longitudeHundredths int,
	areaLocation string, tzType TimezoneType, bytesDecoded int, err error) {

	if len(src) == 0 {
		err = ErrorIncomplete
		return
	}

	if src[0]&maskLatLong != 0 {
		bytesDecoded = byteCountLatLong
		if bytesDecoded > len(src) {
			err = ErrorIncomplete
			return
		}
		latLong := decode32LE(src)
		longitudeHundredths = int(int32(latLong) >> shiftLongitude)
		latitudeHundredths = int((int32(latLong<<16) >> 17) & maskLatitude)
		tzType = TypeLatitudeLongitude
		return
	}

	stringLength := int(src[0] >> 1)
	bytesDecoded = stringLength + 1
	if bytesDecoded > len(src) {
		err = ErrorIncomplete
		return
	}

	areaLocation = string(src[1:bytesDecoded])
	tzType = TypeAreaLocation
	return
}

func decodeDateFields(src []byte) (year, month, day int, bytesDecoded int, err error) {
	if len(src) < minByteCountDate {
		err = ErrorIncomplete
		return
	}

	accumulator := int(decode16LE(src))
	bytesDecoded = 2
	day = int(accumulator & maskDay)
	accumulator >>= sizeDay
	month = int(accumulator & maskMonth)
	accumulator >>= sizeMonth
	asUint, asBig, byteCount, ok := uleb128.Decode(uint64(accumulator), yearLowBitCountDate, src[bytesDecoded:])
	bytesDecoded += byteCount
	if !ok {
		err = ErrorIncomplete
		return
	}
	if asBig != nil {
		err = fmt.Errorf("Year (%v) is too big", asBig)
		return
	}
	if asUint > 0xffffffff {
		err = fmt.Errorf("Year (%v) is too big", asUint)
		return
	}
	year = decodeYear(uint32(asUint))
	return
}

func decodeTimeFields(src []byte) (hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int,
	areaLocation string, tzType TimezoneType, bytesDecoded int, err error) {
	if len(src) == 0 {
		err = ErrorIncomplete
		return
	}

	magnitude := int((src[0] >> 1) & maskMagnitude)
	baseByteCount := baseByteCountsTime[magnitude]
	bytesDecoded = baseByteCount
	if len(src) < baseByteCount {
		err = ErrorIncomplete
		return
	}

	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubseconds := uint(sizeSubsecond * magnitude)
	maskSubsecond := bitMask(int(sizeSubseconds))

	accumulator := decodeLE(src, baseByteCount)
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
			tzType = TypeZeroValue
			return
		}
		err = fmt.Errorf("Expected reserved bits %b but got %b", expectedReservedBits, accumulator)
	}

	if !hasTimezone {
		tzType = TypeZero
		return
	}

	var byteCount int

	latitudeHundredths,
		longitudeHundredths,
		areaLocation,
		tzType,
		byteCount,
		err = decodeTimezone(src[bytesDecoded:])

	bytesDecoded += byteCount
	return
}

func decodeTimestampFields(src []byte) (year, month, day, hour, minute, second, nanosecond,
	latitudeHundredths, longitudeHundredths int, areaLocation string, tzType TimezoneType,
	bytesDecoded int, err error) {

	if len(src) == 0 {
		err = ErrorIncomplete
		return
	}

	magnitude := int((src[0] >> 1) & maskMagnitude)
	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubseconds := uint(sizeSubsecond * magnitude)
	maskSubsecond := bitMask(int(sizeSubseconds))

	baseByteCount := baseByteCountsTimestamp[magnitude]
	bytesDecoded = baseByteCount
	if len(src) < baseByteCount {
		err = ErrorIncomplete
		return
	}

	accumulator := decodeLE(src, baseByteCount)
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
	asUint, asBig, byteCount, ok := uleb128.Decode(uint64(accumulator), yearLowBitCount, src[baseByteCount:])
	bytesDecoded += byteCount
	if !ok {
		err = ErrorIncomplete
		return
	}
	if asBig != nil {
		err = fmt.Errorf("Year (%v) is too big", asBig)
		return
	}
	if asUint > 0xffffffff {
		err = fmt.Errorf("Year (%v) is too big", asUint)
		return
	}
	year = decodeYear(uint32(asUint))

	if !hasTimezone {
		if year == 2000 && month == 0 && day == 0 {
			tzType = TypeZeroValue
			return
		}
		tzType = TypeZero
		return
	}

	latitudeHundredths,
		longitudeHundredths,
		areaLocation,
		tzType,
		byteCount,
		err = decodeTimezone(src[bytesDecoded:])

	bytesDecoded += byteCount
	return
}

var reservedBitsTime = [...]uint64{0x0f, 0x03, 0x00, 0x3f}

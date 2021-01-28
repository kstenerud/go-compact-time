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

// Get the number of bytes that would be required to encode this time value.
func (this *Time) EncodedSize() int {
	if this.IsZeroValue() {
		return byteCountsZeroValue[this.TimeType]
	}
	switch this.TimeType {
	case TypeDate:
		return encodedSizeDate(this.Year)
	case TypeTime:
		return encodedSizeTime(int(this.Nanosecond), this.TimezoneType, this.ShortAreaLocation)
	case TypeTimestamp:
		return encodedSizeTimestamp(this.Year, int(this.Nanosecond), this.TimezoneType, this.ShortAreaLocation)
	default:
		panic(fmt.Errorf("%v: Unknown time type", this.TimeType))
	}
}

// Encode a time value (date, time, or timestamp).
// Returns the number of bytes encoded, or the number of bytes it attempted to encode.
// Returns isComplete=true if there was enough room in dst.
// Returns an error if something went wrong other than there not being enough room.
func (this *Time) Encode(dst []byte) (bytesEncoded int, isComplete bool) {
	switch this.TimeType {
	case TypeDate:
		return this.encodeDate(dst)
	case TypeTime:
		return this.encodeTime(dst)
	case TypeTimestamp:
		return this.encodeTimestamp(dst)
	default:
		panic(fmt.Errorf("%v: Unknown time type", this.TimeType))
	}
}

func (this *Time) encodeDate(dst []byte) (bytesEncoded int, isComplete bool) {
	if this.IsZeroValue() {
		return encodeZeroDate(dst)
	}

	return encodeDate(this.Year, int(this.Month), int(this.Day), dst)
}

func (this *Time) encodeTime(dst []byte) (bytesEncoded int, isComplete bool) {
	if this.IsZeroValue() {
		return encodeZeroTime(dst)
	}

	isZeroTS := this.TimezoneType == TypeZero
	bytesEncoded, isComplete = encodeTime(int(this.Hour), int(this.Minute),
		int(this.Second), int(this.Nanosecond), isZeroTS, dst)
	if isComplete && !isZeroTS {
		accumEncoded := bytesEncoded
		bytesEncoded, isComplete = this.encodeTimezone(dst[bytesEncoded:])
		bytesEncoded += accumEncoded
	}
	return
}

func (this *Time) encodeTimestamp(dst []byte) (bytesEncoded int, isComplete bool) {
	if this.IsZeroValue() {
		return encodeZeroTimestamp(dst)
	}

	isZeroTS := this.TimezoneType == TypeZero
	bytesEncoded, isComplete = encodeTimestamp(this.Year, int(this.Month),
		int(this.Day), int(this.Hour), int(this.Minute), int(this.Second),
		int(this.Nanosecond), isZeroTS, dst)
	if isComplete && !isZeroTS {
		accumEncoded := bytesEncoded
		bytesEncoded, isComplete = this.encodeTimezone(dst[bytesEncoded:])
		bytesEncoded += accumEncoded
	}
	return
}

func (this *Time) encodeTimezone(dst []byte) (bytesEncoded int, isComplete bool) {
	switch this.TimezoneType {
	case TypeZero:
		isComplete = true
	case TypeAreaLocation, TypeLocal:
		return encodeTimezoneAreaLoc(this.ShortAreaLocation, dst)
	case TypeLatitudeLongitude:
		return encodeTimezoneLatLong(int(this.LatitudeHundredths), int(this.LongitudeHundredths), dst)
	default:
		panic(fmt.Errorf("%v: Unknown timezone type", this.TimezoneType))
	}
	return
}

// =============================================================================

func EncodedSizeGoDate(time gotime.Time) int {
	return encodedSizeDate(time.Year())
}

func EncodedSizeGoTime(time gotime.Time) int {
	shortAreaLocation, _ := splitAreaLocation(time.Location().String())
	tzType := getTZTypeFromShortAreaLocation(shortAreaLocation)
	return encodedSizeTime(time.Nanosecond(), tzType, shortAreaLocation)
}

func EncodedSizeGoTimestamp(time gotime.Time) int {
	shortAreaLocation, _ := splitAreaLocation(time.Location().String())
	tzType := getTZTypeFromShortAreaLocation(shortAreaLocation)
	return encodedSizeTimestamp(time.Year(), time.Nanosecond(), tzType, shortAreaLocation)
}

func EncodeGoDate(time gotime.Time, dst []byte) (bytesEncoded int, isComplete bool) {
	return encodeDate(time.Year(), int(time.Month()), int(time.Day()), dst)
}

func EncodeGoTime(time gotime.Time, dst []byte) (bytesEncoded int, isComplete bool) {
	shortAreaLocation, _ := splitAreaLocation(time.Location().String())
	isZeroTS := shortAreaLocation == "Z"
	bytesEncoded, isComplete = encodeTime(time.Hour(), time.Minute(),
		time.Second(), time.Nanosecond(), isZeroTS, dst)
	if isComplete && !isZeroTS {
		accumEncoded := bytesEncoded
		bytesEncoded, isComplete = encodeTimezoneAreaLoc(shortAreaLocation, dst[bytesEncoded:])
		bytesEncoded += accumEncoded
	}
	return
}

func EncodeGoTimestamp(time gotime.Time, dst []byte) (bytesEncoded int, isComplete bool) {
	shortAreaLocation, _ := splitAreaLocation(time.Location().String())
	isZeroTS := shortAreaLocation == "Z"
	bytesEncoded, isComplete = encodeTimestamp(time.Year(), int(time.Month()),
		time.Day(), time.Hour(), time.Minute(), time.Second(),
		time.Nanosecond(), isZeroTS, dst)
	if isComplete && !isZeroTS {
		accumEncoded := bytesEncoded
		bytesEncoded, isComplete = encodeTimezoneAreaLoc(shortAreaLocation, dst[bytesEncoded:])
		bytesEncoded += accumEncoded
	}
	return
}

// =============================================================================

func encodedSizeDate(year int) int {
	encodedYear := encodeYear(year)
	return byteCountDate + getYearGroupCount(encodedYear, yearLowBitCountDate)
}

func encodedSizeTime(nanosecond int, tzType TimezoneType, shortAreaLocation string) int {
	magnitude := getSubsecondMagnitude(nanosecond)
	baseByteCount := baseByteCountsTime[magnitude]

	return baseByteCount + encodedSizeTimezone(tzType, shortAreaLocation)
}

func encodedSizeTimestamp(year, nanosecond int, tzType TimezoneType, shortAreaLocation string) int {
	magnitude := getSubsecondMagnitude(nanosecond)
	baseByteCount := baseByteCountsTimestamp[magnitude]
	encodedYear := encodeYear(year)
	yearGroupCount := getYearGroupCount(encodedYear, yearLowBitCountsTimestamp[magnitude])

	return baseByteCount + yearGroupCount + encodedSizeTimezone(tzType, shortAreaLocation)
}

func encodedSizeTimezone(tzType TimezoneType, shortAreaLocation string) int {
	switch tzType {
	case TypeZero:
		return 0
	case TypeAreaLocation, TypeLocal:
		return 1 + len(shortAreaLocation)
	case TypeLatitudeLongitude:
		return byteCountLatLong
	default:
		panic(fmt.Errorf("%v: Unknown timezone type", tzType))
	}
}

func encodeLE(value uint64, dst []byte, byteCount int) {
	for i := 0; i < byteCount; i++ {
		dst[i] = uint8(value)
		value >>= 8
	}
}

func encode16LE(value uint16, dst []byte) {
	dst[0] = uint8(value)
	dst[1] = uint8(value >> 8)
}

func encode32LE(value uint32, dst []byte) {
	dst[0] = uint8(value)
	dst[1] = uint8(value >> 8)
	dst[2] = uint8(value >> 16)
	dst[3] = uint8(value >> 24)
}

func encodeZigzag32(value int32) uint32 {
	return uint32((value >> 31) ^ (value << 1))
}

func getSubsecondMagnitude(nanosecond int) int {
	if nanosecond == 0 {
		return 0
	}
	if (nanosecond % 1000) != 0 {
		return 3
	}
	if (nanosecond % 1000000) != 0 {
		return 2
	}
	return 1
}

func encodeYear(year int) uint32 {
	return encodeZigzag32(int32(year) - yearBias)
}

func getYearGroupCount(encodedYear uint32, uncountedBits int) int {
	year := encodedYear >> uint32(uncountedBits)
	if year == 0 {
		return 1
	}

	size := 0
	for year != 0 {
		size++
		year >>= bitsPerYearGroup
	}
	return size
}

var zeroBytes = [...]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func encodeZeroBytes(count int, dst []byte) (bytesEncoded int, isComplete bool) {
	if len(dst) < count {
		return
	}
	copy(dst[:count], zeroBytes[:])
	bytesEncoded = count
	isComplete = true
	return
}

func encodeZeroDate(dst []byte) (bytesEncoded int, isComplete bool) {
	return encodeZeroBytes(byteCountsZeroValue[TypeDate], dst)
}

func encodeZeroTime(dst []byte) (bytesEncoded int, isComplete bool) {
	return encodeZeroBytes(byteCountsZeroValue[TypeTime], dst)
}

func encodeZeroTimestamp(dst []byte) (bytesEncoded int, isComplete bool) {
	return encodeZeroBytes(byteCountsZeroValue[TypeTimestamp], dst)
}

func encodeDate(year, month, day int, dst []byte) (bytesEncoded int, isComplete bool) {
	if len(dst) < byteCountDate {
		return
	}

	encodedYear := encodeYear(year)
	yearGroupedMask := uint32(bitMask(yearLowBitCountDate))

	accumulator := uint16(encodedYear & yearGroupedMask)
	accumulator = (accumulator << uint(sizeMonth)) | uint16(month)
	accumulator = (accumulator << uint(sizeDay)) | uint16(day)

	encode16LE(accumulator, dst)
	byteCount := 0
	byteCount, isComplete = uleb128.EncodeUint64(uint64(encodedYear>>yearLowBitCountDate), dst[byteCountDate:])
	bytesEncoded = byteCount + byteCountDate
	return
}

func encodeTime(hour, minute, second, nanosecond int, isZeroTS bool, dst []byte) (bytesEncoded int, isComplete bool) {
	magnitude := getSubsecondMagnitude(nanosecond)
	baseByteCount := baseByteCountsTime[magnitude]
	if len(dst) < baseByteCount {
		return
	}

	subsecond := nanosecond / subsecMultipliers[magnitude]

	accumulator := ^uint64(0)
	accumulator = (accumulator << uint(sizeHour)) | uint64(hour)
	accumulator = (accumulator << uint(sizeMinute)) | uint64(minute)
	accumulator = (accumulator << uint(sizeSecond)) | uint64(second)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) | uint64(subsecond)
	accumulator = (accumulator << uint(sizeMagnitude)) | uint64(magnitude)
	accumulator <<= 1
	if !isZeroTS {
		accumulator |= 1
	}

	encodeLE(accumulator, dst, baseByteCount)
	bytesEncoded = baseByteCount
	isComplete = true
	return
}

func encodeTimestamp(year, month, day, hour, minute, second, nanosecond int,
	isZeroTS bool, dst []byte) (bytesEncoded int, isComplete bool) {
	magnitude := getSubsecondMagnitude(nanosecond)
	baseByteCount := baseByteCountsTimestamp[magnitude]
	if len(dst) < baseByteCount {
		return
	}

	subsecond := nanosecond / subsecMultipliers[magnitude]
	encodedYear := encodeYear(year)
	yearLowBitCount := yearLowBitCountsTimestamp[magnitude]

	accumulator := uint64(encodedYear)
	accumulator = (accumulator << uint(sizeMonth)) | uint64(month)
	accumulator = (accumulator << uint(sizeDay)) | uint64(day)
	accumulator = (accumulator << uint(sizeHour)) | uint64(hour)
	accumulator = (accumulator << uint(sizeMinute)) | uint64(minute)
	accumulator = (accumulator << uint(sizeSecond)) | uint64(second)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) | uint64(subsecond)
	accumulator = (accumulator << uint(sizeMagnitude)) | uint64(magnitude)
	accumulator <<= 1
	if !isZeroTS {
		accumulator |= 1
	}

	encodeLE(accumulator, dst, baseByteCount)

	byteCount := 0
	byteCount, isComplete = uleb128.EncodeUint64(uint64(encodedYear>>uint(yearLowBitCount)), dst[baseByteCount:])
	bytesEncoded = byteCount + baseByteCount
	return
}

func encodeTimezoneAreaLoc(areaLocation string, dst []byte) (bytesEncoded int, isComplete bool) {
	bytesEncoded = len(areaLocation) + 1
	if len(dst) < bytesEncoded {
		return
	}
	dst[0] = byte(len(areaLocation) << shiftLength)
	copy(dst[1:], areaLocation)
	isComplete = true
	return
}

func encodeTimezoneLatLong(latitudeHundredths, longitudeHundredths int, dst []byte) (bytesEncoded int, isComplete bool) {
	bytesEncoded = byteCountLatLong
	if len(dst) < bytesEncoded {
		return
	}
	latLong := ((longitudeHundredths & maskLongitude) << shiftLongitude) |
		((latitudeHundredths & maskLatitude) << shiftLatitude) | maskLatLong
	encode32LE(uint32(latLong), dst)
	isComplete = true
	return
}

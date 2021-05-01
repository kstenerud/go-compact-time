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
	gotime "time"

	"github.com/kstenerud/go-uleb128"
)

// Get the number of bytes that would be required to encode this time value.
func (this *Time) EncodedSize() int {
	if this.IsZeroValue() {
		return byteCountsZeroValue[this.Type]
	}
	switch this.Type {
	case TimeTypeDate:
		return encodedSizeDate(this.Year)
	case TimeTypeTime:
		return encodedSizeTime(int(this.Nanosecond), this.Timezone.Type, this.Timezone.ShortAreaLocation)
	case TimeTypeTimestamp:
		return encodedSizeTimestamp(this.Year, int(this.Nanosecond), this.Timezone.Type, this.Timezone.ShortAreaLocation)
	default:
		panic(fmt.Errorf("%v: Unknown time type", this.Type))
	}
}

// Encode a time value (date, time, or timestamp).
func (this *Time) Encode(writer io.Writer) (bytesEncoded int, err error) {
	buffer := make([]byte, this.EncodedSize())
	bytesEncoded = this.EncodeToBytes(buffer)
	_, err = writer.Write(buffer[:bytesEncoded])
	return
}

// Encode a time value (date, time, or timestamp) to a byte array.
// Assumes that the buffer is big enough.
func (this *Time) EncodeToBytes(buffer []byte) (bytesEncoded int) {
	switch this.Type {
	case TimeTypeDate:
		return this.encodeDate(buffer)
	case TimeTypeTime:
		return this.encodeTime(buffer)
	case TimeTypeTimestamp:
		return this.encodeTimestamp(buffer)
	default:
		panic(fmt.Errorf("%v: Unknown time type", this.Type))
	}
}

func (this *Time) encodeDate(buffer []byte) (bytesEncoded int) {
	if this.IsZeroValue() {
		return encodeZeroDate(buffer)
	}

	return encodeDate(this.Year, int(this.Month), int(this.Day), buffer)
}

func (this *Time) encodeTime(buffer []byte) (bytesEncoded int) {
	if this.IsZeroValue() {
		return encodeZeroTime(buffer)
	}

	isZeroTS := this.Timezone.Type == TimezoneTypeUTC
	bytesEncoded = encodeTime(int(this.Hour), int(this.Minute),
		int(this.Second), int(this.Nanosecond), isZeroTS, buffer)
	if !isZeroTS {
		bytesEncoded += this.encodeTimezone(buffer[bytesEncoded:])
	}
	return
}

func (this *Time) encodeTimestamp(buffer []byte) (bytesEncoded int) {
	if this.IsZeroValue() {
		return encodeZeroTimestamp(buffer)
	}

	isZeroTS := this.Timezone.Type == TimezoneTypeUTC
	bytesEncoded = encodeTimestamp(this.Year, int(this.Month),
		int(this.Day), int(this.Hour), int(this.Minute), int(this.Second),
		int(this.Nanosecond), isZeroTS, buffer)
	if !isZeroTS {
		bytesEncoded += this.encodeTimezone(buffer[bytesEncoded:])
	}
	return
}

func (this *Time) encodeTimezone(buffer []byte) (bytesEncoded int) {
	switch this.Timezone.Type {
	case TimezoneTypeUTC:
		return
	case TimezoneTypeAreaLocation, TimezoneTypeLocal:
		return encodeTimezoneAreaLoc(this.Timezone.ShortAreaLocation, buffer)
	case TimezoneTypeLatitudeLongitude:
		return encodeTimezoneLatLong(int(this.Timezone.LatitudeHundredths),
			int(this.Timezone.LongitudeHundredths), buffer)
	case TimezoneTypeUTCOffset:
		return encodeTimezoneUTCOffset(int(this.Timezone.MinutesOffsetFromUTC), buffer)
	default:
		panic(fmt.Errorf("%v: Unknown timezone type", this.Timezone.Type))
	}
}

// =============================================================================

func EncodedSizeGoDate(time gotime.Time) int {
	return encodedSizeDate(time.Year())
}

func EncodedSizeGoTime(time gotime.Time) int {
	tz := TZAtAreaLocation(time.Location().String())
	return encodedSizeTime(time.Nanosecond(), tz.Type, tz.ShortAreaLocation)
}

func EncodedSizeGoTimestamp(time gotime.Time) int {
	tz := TZAtAreaLocation(time.Location().String())
	return encodedSizeTimestamp(time.Year(), time.Nanosecond(), tz.Type, tz.ShortAreaLocation)
}

func EncodeGoDate(time gotime.Time, writer io.Writer) (bytesEncoded int, err error) {
	buffer := make([]byte, EncodedSizeGoDate(time))
	bytesEncoded = EncodeGoDateToBytes(time, buffer)
	_, err = writer.Write(buffer[:bytesEncoded])
	return
}

func EncodeGoDateToBytes(time gotime.Time, buffer []byte) (bytesEncoded int) {
	return encodeDate(time.Year(), int(time.Month()), int(time.Day()), buffer)
}

func EncodeGoTime(time gotime.Time, writer io.Writer) (bytesEncoded int, err error) {
	buffer := make([]byte, EncodedSizeGoTime(time))
	bytesEncoded = EncodeGoTimeToBytes(time, buffer)
	_, err = writer.Write(buffer[:bytesEncoded])
	return
}

func EncodeGoTimeToBytes(time gotime.Time, buffer []byte) (bytesEncoded int) {
	tz := TZAtAreaLocation(time.Location().String())
	bytesEncoded = encodeTime(time.Hour(), time.Minute(),
		time.Second(), time.Nanosecond(), tz.Type == TimezoneTypeUTC, buffer)
	if tz.Type != TimezoneTypeUTC {
		bytesEncoded += encodeTimezoneAreaLoc(tz.ShortAreaLocation, buffer[bytesEncoded:])
	}
	return
}

func EncodeGoTimestamp(time gotime.Time, writer io.Writer) (bytesEncoded int, err error) {
	buffer := make([]byte, EncodedSizeGoTimestamp(time))
	bytesEncoded = EncodeGoTimestampToBytes(time, buffer)
	_, err = writer.Write(buffer[:bytesEncoded])
	return
}

func EncodeGoTimestampToBytes(time gotime.Time, buffer []byte) (bytesEncoded int) {
	tz := TZAtAreaLocation(time.Location().String())
	bytesEncoded = encodeTimestamp(time.Year(), int(time.Month()),
		time.Day(), time.Hour(), time.Minute(), time.Second(),
		time.Nanosecond(), tz.Type == TimezoneTypeUTC, buffer)
	if tz.Type != TimezoneTypeUTC {
		bytesEncoded += encodeTimezoneAreaLoc(tz.ShortAreaLocation, buffer[bytesEncoded:])
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
	case TimezoneTypeUTC:
		return 0
	case TimezoneTypeAreaLocation, TimezoneTypeLocal:
		return 1 + len(shortAreaLocation)
	case TimezoneTypeLatitudeLongitude:
		return byteCountLatLong
	case TimezoneTypeUTCOffset:
		return byteCountUTCOffset
	default:
		panic(fmt.Errorf("%v: Unknown timezone type", tzType))
	}
}

func encodeLE(value uint64, buffer []byte, byteCount int) (bytesEncoded int) {
	for i := 0; i < byteCount; i++ {
		buffer[i] = uint8(value)
		value >>= 8
	}
	return byteCount
}

func encode16LE(value uint16, buffer []byte) (bytesEncoded int) {
	buffer[0] = uint8(value)
	buffer[1] = uint8(value >> 8)
	return 2
}

func encode32LE(value uint32, buffer []byte) (bytesEncoded int) {
	buffer[0] = uint8(value)
	buffer[1] = uint8(value >> 8)
	buffer[2] = uint8(value >> 16)
	buffer[3] = uint8(value >> 24)
	return 4
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

func encodeZeroBytes(count int, buffer []byte) (bytesEncoded int) {
	if len(buffer) < count {
		panic(fmt.Errorf("Attempt to copy %v bytes into %v buffer", count, len(buffer)))
	}
	return copy(buffer, zeroBytes[:count])
}

func encodeZeroDate(buffer []byte) (bytesEncoded int) {
	return encodeZeroBytes(byteCountsZeroValue[TimeTypeDate], buffer)
}

func encodeZeroTime(buffer []byte) (bytesEncoded int) {
	return encodeZeroBytes(byteCountsZeroValue[TimeTypeTime], buffer)
}

func encodeZeroTimestamp(buffer []byte) (bytesEncoded int) {
	return encodeZeroBytes(byteCountsZeroValue[TimeTypeTimestamp], buffer)
}

func encodeDate(year, month, day int, buffer []byte) (bytesEncoded int) {
	encodedYear := encodeYear(year)
	yearGroupedMask := uint32(bitMask(yearLowBitCountDate))

	accumulator := uint16(encodedYear & yearGroupedMask)
	accumulator = (accumulator << uint(sizeMonth)) | uint16(month)
	accumulator = (accumulator << uint(sizeDay)) | uint16(day)

	bytesEncoded = encode16LE(accumulator, buffer)
	bytesEncoded += uleb128.EncodeUint64ToBytes(uint64(encodedYear>>yearLowBitCountDate), buffer[bytesEncoded:])
	return
}

func encodeTime(hour, minute, second, nanosecond int, isZeroTS bool, buffer []byte) (bytesEncoded int) {
	magnitude := getSubsecondMagnitude(nanosecond)
	baseByteCount := baseByteCountsTime[magnitude]

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

	return encodeLE(accumulator, buffer, baseByteCount)
}

func encodeTimestamp(year, month, day, hour, minute, second, nanosecond int,
	isZeroTS bool, buffer []byte) (bytesEncoded int) {
	magnitude := getSubsecondMagnitude(nanosecond)
	baseByteCount := baseByteCountsTimestamp[magnitude]

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

	bytesEncoded = encodeLE(accumulator, buffer, baseByteCount)
	bytesEncoded += uleb128.EncodeUint64ToBytes(uint64(encodedYear>>uint(yearLowBitCount)), buffer[bytesEncoded:])
	return
}

func encodeTimezoneAreaLoc(areaLocation string, buffer []byte) (bytesEncoded int) {
	buffer[0] = byte(len(areaLocation) << shiftLength)
	return copy(buffer[1:], areaLocation) + 1
}

func encodeTimezoneLatLong(latitudeHundredths, longitudeHundredths int, buffer []byte) (bytesEncoded int) {
	latLong := ((longitudeHundredths & maskLongitude) << shiftLongitude) |
		((latitudeHundredths & maskLatitude) << shiftLatitude) | maskLatLong
	return encode32LE(uint32(latLong), buffer)
}

func encodeTimezoneUTCOffset(minutesOffsetFromUTC int, buffer []byte) (bytesEncoded int) {
	return encodeLE(uint64(minutesOffsetFromUTC)<<8, buffer, byteCountUTCOffset)
}

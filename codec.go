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
	"strings"

	"github.com/kstenerud/go-uleb128"
)

var ErrorIncomplete = fmt.Errorf("Compact time value is incomplete")

// Decode a date.
// Warning: The date fields will not be validated! Please call time.Validate()!
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
// Returns isComplete=true if there was enough data in src.
// If isComplete == false, the resulting date is invalid.
func DecodeDate(src []byte) (time *Time, bytesDecoded int, err error) {
	if len(src) < byteCountDate {
		err = ErrorIncomplete
		return
	}

	time = new(Time)
	time.TimeIs = TypeDate
	time.TimezoneIs = TypeUTC
	accumulator := int(decode16LE(src))
	bytesDecoded = 2
	time.Day = uint8(accumulator & maskDay)
	accumulator >>= sizeDay
	time.Month = uint8(accumulator & maskMonth)
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
	time.Year = decodeYear(uint32(asUint))
	return
}

// Decode a time value.
// Warning: The date fields will not be validated! Please call time.Validate()!
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
// Returns isComplete=true if there was enough data in src.
// If isComplete == false, the resulting time value is invalid.
func DecodeTime(src []byte) (time *Time, bytesDecoded int, isComplete bool) {
	if len(src) == 0 {
		return
	}

	magnitude := int((src[0] >> 1) & maskMagnitude)
	baseByteCount := baseByteCountsTime[magnitude]
	if len(src) < baseByteCount {
		return
	}

	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubseconds := uint(sizeSubsecond * magnitude)
	maskSubsecond := bitMask(int(sizeSubseconds))

	time = new(Time)
	time.TimeIs = TypeTime
	accumulator := decodeLE(src, baseByteCount)
	hasTimezone := accumulator&1 == 1
	accumulator >>= 1
	accumulator >>= sizeMagnitude
	time.Nanosecond = uint32(accumulator&maskSubsecond) * uint32(subsecondMultiplier)
	accumulator >>= sizeSubseconds
	time.Second = uint8(accumulator & maskSecond)
	accumulator >>= sizeSecond
	time.Minute = uint8(accumulator & maskMinute)
	accumulator >>= sizeMinute
	time.Hour = uint8(accumulator & maskHour)

	if byteCount, ok := decodeTimezone(src[baseByteCount:], time, hasTimezone); ok {
		bytesDecoded = baseByteCount + byteCount
		isComplete = true
	}

	return
}

// Decode a timestamp.
// Warning: The date fields will not be validated! Please call time.Validate()!
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
// Returns isComplete=true if there was enough data in src.
// If isComplete == false, the resulting timestamp is invalid.
func DecodeTimestamp(src []byte) (time *Time, bytesDecoded int, err error) {
	if len(src) == 0 {
		err = ErrorIncomplete
		return
	}

	magnitude := int((src[0] >> 1) & maskMagnitude)
	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubseconds := uint(sizeSubsecond * magnitude)
	maskSubsecond := bitMask(int(sizeSubseconds))

	baseByteCount := baseByteCountsTimestamp[magnitude]
	if len(src) < baseByteCount {
		err = ErrorIncomplete
		return
	}

	time = new(Time)
	time.TimeIs = TypeTimestamp
	accumulator := decodeLE(src, baseByteCount)
	hasTimezone := accumulator&1 == 1
	accumulator >>= 1
	accumulator >>= sizeMagnitude
	time.Nanosecond = uint32(accumulator&maskSubsecond) * uint32(subsecondMultiplier)
	accumulator >>= sizeSubseconds
	time.Second = uint8(accumulator & maskSecond)
	accumulator >>= sizeSecond
	time.Minute = uint8(accumulator & maskMinute)
	accumulator >>= sizeMinute
	time.Hour = uint8(accumulator & maskHour)
	accumulator >>= sizeHour
	time.Day = uint8(accumulator & maskDay)
	accumulator >>= sizeDay
	time.Month = uint8(accumulator & maskMonth)
	accumulator >>= sizeMonth

	yearLowBitCount := yearLowBitCountsTimestamp[magnitude]
	asUint, asBig, byteCount, ok := uleb128.Decode(uint64(accumulator), yearLowBitCount, src[baseByteCount:])
	bytesDecoded = byteCount + baseByteCount
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
	time.Year = decodeYear(uint32(asUint))

	if byteCount, ok = decodeTimezone(src[bytesDecoded:], time, hasTimezone); !ok {
		err = ErrorIncomplete
		return
	}
	bytesDecoded += byteCount

	return
}

// Get the number of bytes that would be required to encode this time value.
func EncodedSize(time *Time) int {
	switch time.TimeIs {
	case TypeDate:
		return encodedSizeDate(time)
	case TypeTime:
		return encodedSizeTime(time)
	case TypeTimestamp:
		return encodedSizeTimestamp(time)
	default:
		panic(fmt.Errorf("%v: Unknown time type", time.TimeIs))
	}
}

// Encode a time value (date, time, or timestamp).
// Returns the number of bytes encoded, or the number of bytes it attempted to encode.
// Returns isComplete=true if there was enough room in dst.
// Returns an error if something went wrong other than there not being enough room.
func Encode(time *Time, dst []byte) (bytesEncoded int, isComplete bool) {
	switch time.TimeIs {
	case TypeDate:
		bytesEncoded, isComplete = encodeDate(time, dst)
		return
	case TypeTime:
		return encodeTime(time, dst)
	case TypeTimestamp:
		return encodeTimestamp(time, dst)
	default:
		panic(fmt.Errorf("%v: Unknown time type", time.TimeIs))
	}
}

// Maximum byte length that this library will encode
const MaxEncodeLength = 50

const yearBias = 2000
const bitsPerYearGroup = 7

const sizeUtc = 1
const sizeMagnitude = 2
const sizeSubsecond = 10
const sizeSecond = 6
const sizeMinute = 6
const sizeHour = 5
const sizeDay = 5
const sizeMonth = 4
const sizeLatitude = 15
const sizeLongitude = 16
const yearLowBitCountDate = 7

const baseSizeTime = sizeUtc + sizeMagnitude + sizeSecond + sizeMinute + sizeHour
const baseSizeTimestamp = sizeMagnitude + sizeSecond + sizeMinute + sizeHour + sizeDay + sizeMonth

const byteCountDate = 2
const byteCountLatLong = 4

const maskLatLong = 1
const maskMagnitude = ((1 << sizeMagnitude) - 1)
const maskSecond = ((1 << sizeSecond) - 1)
const maskMinute = ((1 << sizeMinute) - 1)
const maskHour = ((1 << sizeHour) - 1)
const maskDay = ((1 << sizeDay) - 1)
const maskMonth = ((1 << sizeMonth) - 1)
const maskLatitude = ((1 << sizeLatitude) - 1)
const maskLongitude = ((1 << sizeLongitude) - 1)
const maskDateYearUpperBits = ((1 << yearLowBitCountDate) - 1)

const shiftLength = 1
const shiftLatitude = 1
const shiftLongitude = 16

var yearLowBitCountsTimestamp = [...]int{3, 1, 7, 5}
var subsecMultipliers = [...]int{1, 1000000, 1000, 1}
var baseByteCountsTime = [...]int{3, 4, 5, 7}
var baseByteCountsTimestamp = [...]int{4, 5, 7, 8}

var abbrevToTimezone = map[rune]string{
	'F': "Africa",
	'M': "America",
	'N': "Antarctica",
	'R': "Arctic",
	'S': "Asia",
	'T': "Atlantic",
	'U': "Australia",
	'C': "Etc",
	'E': "Europe",
	'I': "Indian",
	'P': "Pacific",
	'Z': "Etc/UTC",
	'L': "Local",
}

var timezoneToAbbrev = map[string]string{
	"Africa":     "F",
	"America":    "M",
	"Antarctica": "N",
	"Arctic":     "R",
	"Asia":       "S",
	"Atlantic":   "T",
	"Australia":  "U",
	"Etc":        "C",
	"Europe":     "E",
	"Indian":     "I",
	"Pacific":    "P",
	"Etc/UTC":    "Z",
	"Local":      "L",
}

func getFullTimezoneString(tz string) string {
	if len(tz) == 0 {
		return tz
	}
	firstChar := rune(tz[0])

	if len(tz) == 1 {
		switch firstChar {
		case 'L':
			return "Local"
		case 'Z':
			return "Etc/UTC"
		}
		return tz
	}
	if tz[1] != '/' {
		return tz
	}

	remainder := tz[1:]
	if value, exists := abbrevToTimezone[firstChar]; exists {
		return value + remainder
	}
	return tz
}

func getAbbreviatedTimezoneString(tz string) string {
	if len(tz) == 0 {
		return "Z"
	}

	if tz == "Local" {
		return "L"
	}

	index := strings.Index(tz, "/")
	if index < 1 {
		return tz
	}

	area := tz[:index]

	if value, exists := timezoneToAbbrev[area]; exists {
		return value + tz[index:]
	}

	return tz
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

func zigzagEncode32(value int32) uint32 {
	return uint32((value >> 31) ^ (value << 1))
}

func zigzagDecode32(value uint32) int32 {
	return int32((value >> 1) ^ -(value & 1))
}

func bitMask(bitCount int) uint64 {
	return uint64(1)<<uint(bitCount) - 1
}

func encodeYear(year int) uint32 {
	return zigzagEncode32(int32(year) - yearBias)
}

func decodeYear(encodedYear uint32) int {
	return int(zigzagDecode32(uint32(encodedYear))) + yearBias
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

func encodedSizeTimezone(time *Time) int {
	switch time.TimezoneIs {
	case TypeUTC:
		return 0
	case TypeAreaLocation:
		return 1 + len(getAbbreviatedTimezoneString(time.AreaLocation))
	case TypeLatitudeLongitude:
		return byteCountLatLong
	default:
		panic(fmt.Errorf("%v: Unknown timezone type", time.TimezoneIs))
	}
}

func encodeTimezone(time *Time, dst []byte) (bytesEncoded int, isComplete bool) {
	switch time.TimezoneIs {
	case TypeUTC:
		isComplete = true
	case TypeAreaLocation:
		areaLocation := getAbbreviatedTimezoneString(time.AreaLocation)
		bytesEncoded = len(areaLocation) + 1
		if len(dst) < bytesEncoded {
			return
		}
		dst[0] = byte(len(areaLocation) << shiftLength)
		copy(dst[1:], areaLocation)
		isComplete = true
	case TypeLatitudeLongitude:
		bytesEncoded = byteCountLatLong
		if len(dst) < bytesEncoded {
			return
		}
		latLong := ((int(time.LongitudeHundredths) & maskLongitude) << shiftLongitude) |
			((int(time.LatitudeHundredths) & maskLatitude) << shiftLatitude) | maskLatLong
		encode32LE(uint32(latLong), dst)
		isComplete = true
	default:
		panic(fmt.Errorf("%v: Unknown timezone type", time.TimezoneIs))
	}
	return
}

func decodeTimezone(src []byte, time *Time, hasTimezone bool) (bytesDecoded int, isComplete bool) {
	if !hasTimezone {
		time.TimezoneIs = TypeUTC
		isComplete = true
		return
	}

	if len(src) == 0 {
		return
	}

	if src[0]&maskLatLong != 0 {
		time.TimezoneIs = TypeLatitudeLongitude
		bytesDecoded = byteCountLatLong
		if bytesDecoded > len(src) {
			return
		}
		latLong := decode32LE(src)
		time.LongitudeHundredths = int16(int32(latLong) >> shiftLongitude)
		time.LatitudeHundredths = int16((int32(latLong<<16) >> 17) & maskLatitude)
		isComplete = true
		return
	}

	time.TimezoneIs = TypeAreaLocation
	stringLength := int(src[0] >> 1)
	bytesDecoded = stringLength + 1
	if bytesDecoded > len(src) {
		return
	}

	areaLocation := string(src[1:bytesDecoded])
	time.AreaLocation = getFullTimezoneString(areaLocation)
	isComplete = true
	return
}

func encodedSizeDate(time *Time) int {
	encodedYear := encodeYear(time.Year)
	return byteCountDate + getYearGroupCount(encodedYear, yearLowBitCountDate)
}

func encodeDate(time *Time, dst []byte) (bytesEncoded int, isComplete bool) {
	if len(dst) < byteCountDate {
		return
	}

	encodedYear := encodeYear(time.Year)
	yearGroupedMask := uint32(bitMask(yearLowBitCountDate))

	accumulator := uint16(encodedYear & yearGroupedMask)
	accumulator = (accumulator << uint(sizeMonth)) | uint16(time.Month)
	accumulator = (accumulator << uint(sizeDay)) | uint16(time.Day)

	encode16LE(accumulator, dst)
	byteCount := 0
	byteCount, isComplete = uleb128.EncodeUint64(uint64(encodedYear>>yearLowBitCountDate), dst[byteCountDate:])
	bytesEncoded = byteCount + byteCountDate
	return
}

func encodedSizeTime(time *Time) int {
	magnitude := getSubsecondMagnitude(int(time.Nanosecond))
	baseByteCount := baseByteCountsTime[magnitude]

	return baseByteCount + encodedSizeTimezone(time)
}

func encodeTime(time *Time, dst []byte) (bytesEncoded int, isComplete bool) {
	magnitude := getSubsecondMagnitude(int(time.Nanosecond))
	baseByteCount := baseByteCountsTime[magnitude]
	if len(dst) < baseByteCount {
		return
	}

	subsecond := int(time.Nanosecond) / subsecMultipliers[magnitude]

	accumulator := uint64(time.Hour)
	accumulator = (accumulator << uint(sizeMinute)) | uint64(time.Minute)
	accumulator = (accumulator << uint(sizeSecond)) | uint64(time.Second)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) | uint64(subsecond)
	accumulator = (accumulator << uint(sizeMagnitude)) | uint64(magnitude)
	accumulator <<= 1
	if time.TimezoneIs != TypeUTC {
		accumulator |= 1
	}

	encodeLE(accumulator, dst, baseByteCount)
	byteCount := 0
	byteCount, isComplete = encodeTimezone(time, dst[baseByteCount:])
	bytesEncoded = byteCount + baseByteCount
	return
}

func encodedSizeTimestamp(time *Time) int {
	magnitude := getSubsecondMagnitude(int(time.Nanosecond))
	baseByteCount := baseByteCountsTimestamp[magnitude]
	encodedYear := encodeYear(time.Year)
	yearGroupCount := getYearGroupCount(encodedYear, yearLowBitCountsTimestamp[magnitude])

	return baseByteCount + yearGroupCount + encodedSizeTimezone(time)
}

func encodeTimestamp(time *Time, dst []byte) (bytesEncoded int, isComplete bool) {
	magnitude := getSubsecondMagnitude(int(time.Nanosecond))
	baseByteCount := baseByteCountsTimestamp[magnitude]
	if len(dst) < baseByteCount {
		return
	}

	subsecond := int(time.Nanosecond) / subsecMultipliers[magnitude]
	encodedYear := encodeYear(time.Year)
	yearLowBitCount := yearLowBitCountsTimestamp[magnitude]

	accumulator := uint64(encodedYear)
	accumulator = (accumulator << uint(sizeMonth)) | uint64(time.Month)
	accumulator = (accumulator << uint(sizeDay)) | uint64(time.Day)
	accumulator = (accumulator << uint(sizeHour)) | uint64(time.Hour)
	accumulator = (accumulator << uint(sizeMinute)) | uint64(time.Minute)
	accumulator = (accumulator << uint(sizeSecond)) | uint64(time.Second)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) | uint64(subsecond)
	accumulator = (accumulator << uint(sizeMagnitude)) | uint64(magnitude)
	accumulator <<= 1
	if time.TimezoneIs != TypeUTC {
		accumulator |= 1
	}

	encodeLE(accumulator, dst, baseByteCount)

	byteCount := 0
	byteCount, isComplete = uleb128.EncodeUint64(uint64(encodedYear>>uint(yearLowBitCount)), dst[baseByteCount:])
	bytesEncoded = byteCount + baseByteCount
	if !isComplete {
		return
	}

	byteCount, isComplete = encodeTimezone(time, dst[bytesEncoded:])
	bytesEncoded += byteCount
	return
}

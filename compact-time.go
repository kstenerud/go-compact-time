package compact_time

// Package compact_time provides encoding and decoding mechanisms for the
// compact time format (https://github.com/kstenerud/compact-time), as well as
// basic conversion functions to/from the go time package.
//
// Basic validation is performed when decoding data, enough to ensure that it
// isn't blatantly wrong (such as invalid area/location values, latitude 500,
// december 54th, etc). However, it does not do more nuanced checks such as on which
// years February 29th is valid, or when leap seconds are allowed. It also doesn't
// check for impossible timestamp values such as 2011-03-13/02:10:00/Los_Angeles.
//
// If a function returns with err != nil, none of the other fields can be trusted.
// A best effort will be made to set bytesEncoded or bytesDecoded to a position
// in the vicinity of where the error occurred, but it is not guaranteed to be
// exact.

import (
	"fmt"
	"strings"
	gotime "time"

	"github.com/kstenerud/go-vlq"
)

type TimeType int

const (
	TypeUnset = TimeType(iota)
	TypeDate
	TypeTime
	TypeTimestamp
)

type TimezoneType int

const (
	TypeUTC = TimezoneType(iota)
	TypeAreaLocation
	TypeLatitudeLongitude
)

type Time struct {
	TimeIs              TimeType
	TimezoneIs          TimezoneType
	Year                int
	Month               int
	Day                 int
	Hour                int
	Minute              int
	Second              int
	Nanosecond          int
	LatitudeHundredths  int
	LongitudeHundredths int
	AreaLocation        string
}

const (
	monthMin      = 1
	monthMax      = 12
	dayMin        = 1
	hourMin       = 0
	hourMax       = 23
	minuteMin     = 0
	minuteMax     = 59
	secondMin     = 0
	secondMax     = 60
	nanosecondMin = 0
	nanosecondMax = 999999999
	latitudeMin   = -9000
	latitudeMax   = 9000
	longitudeMin  = -18000
	longitudeMax  = 18000
)

var dayMax = [...]int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

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
const sizeDateYearUpperBits = 7

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
const maskDateYearUpperBits = ((1 << sizeDateYearUpperBits) - 1)

const shiftLength = 1
const shiftLatitude = 1
const shiftLongitude = 16

var timestampYearUpperBits = [...]int{4, 2, 0, 6}
var subsecMultipliers = [...]int{1, 1000000, 1000, 1}

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
	if val, ok := abbrevToTimezone[firstChar]; ok {
		return val + remainder
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

	if value, ok := timezoneToAbbrev[area]; ok {
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

func zigzagEncode(value int32) uint32 {
	return uint32((value >> 31) ^ (value << 1))
}

func zigzagDecode(value uint32) int32 {
	return int32((value >> 1) ^ -(value & 1))
}

func encodeYear(year int) uint32 {
	return zigzagEncode(int32(year) - yearBias)
}

func decodeYear(encodedYear uint32) int {
	return int(zigzagDecode(uint32(encodedYear))) + yearBias
}

func getBaseByteCount(baseSize int, magnitude int) int {
	size := baseSize + sizeSubsecond*magnitude
	remainder := int8(size & 7)
	extraByte := ((remainder | (-remainder)) >> 7) & 1
	return size/8 + int(extraByte)
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

func validateDate(time *Time) error {
	if time.Month < monthMin || time.Month > monthMax {
		return fmt.Errorf("%v: Invalid month", time.Month)
	}
	if time.Day < dayMin || time.Day > dayMax[time.Month] {
		return fmt.Errorf("%v: Invalid day", time.Day)
	}
	return nil
}

func validateTime(time *Time) error {
	if time.Hour < hourMin || time.Hour > hourMax {
		return fmt.Errorf("%v: Invalid hour", time.Hour)
	}
	if time.Minute < minuteMin || time.Minute > minuteMax {
		return fmt.Errorf("%v: Invalid minute", time.Minute)
	}
	if time.Second < secondMin || time.Second > secondMax {
		return fmt.Errorf("%v: Invalid second", time.Second)
	}
	if time.Nanosecond < nanosecondMin || time.Nanosecond > nanosecondMax {
		return fmt.Errorf("%v: Invalid nanosecond", time.Nanosecond)
	}
	return nil
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
		panic(fmt.Errorf("BUG: %v: Unhandled timezone type", time.TimezoneIs))
	}
}

func encodeTimezone(time *Time, dst []byte) (bytesEncoded int, ok bool, err error) {
	switch time.TimezoneIs {
	case TypeUTC:
		ok = true
	case TypeAreaLocation:
		areaLocation := getAbbreviatedTimezoneString(time.AreaLocation)
		bytesEncoded = len(areaLocation) + 1
		if len(dst) < bytesEncoded {
			return
		}
		dst[0] = byte(len(areaLocation) << shiftLength)
		copy(dst[1:], areaLocation)
		ok = true
	case TypeLatitudeLongitude:
		bytesEncoded = byteCountLatLong
		if bytesEncoded > len(dst) {
			return
		}
		latLong := ((time.LongitudeHundredths & maskLongitude) << shiftLongitude) |
			((time.LatitudeHundredths & maskLatitude) << shiftLatitude) | maskLatLong
		encode32LE(uint32(latLong), dst)
		ok = true
	default:
		err = fmt.Errorf("BUG: %v: Unhandled timezone type", time.TimezoneIs)
	}
	return
}

func decodeTimezone(src []byte, time *Time, isUTC bool) (bytesDecoded int, ok bool, err error) {
	if isUTC {
		time.TimezoneIs = TypeUTC
		ok = true
		return
	}

	if len(src) < 1 {
		bytesDecoded = 1
		return
	}

	if src[0]&maskLatLong != 0 {
		time.TimezoneIs = TypeLatitudeLongitude
		bytesDecoded = byteCountLatLong
		if bytesDecoded > len(src) {
			return
		}
		latLong := decode32LE(src)
		time.LongitudeHundredths = int(int32(latLong) >> shiftLongitude)
		time.LatitudeHundredths = int((int32(latLong<<16) >> 17) & maskLatitude)
		if time.LongitudeHundredths < longitudeMin || time.LongitudeHundredths > longitudeMax {
			err = fmt.Errorf("Lontitude %v is out of range (must be %v to %v)", time.LongitudeHundredths, longitudeMin, longitudeMax)
			return
		}
		if time.LatitudeHundredths < latitudeMin || time.LatitudeHundredths > latitudeMax {
			err = fmt.Errorf("Latitude %v is out of range (must be %v to %v)", time.LatitudeHundredths, latitudeMin, latitudeMax)
			return
		}
		ok = true
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
	if time.AreaLocation == "Local" {
		ok = true
		return
	}

	_, err = gotime.LoadLocation(time.AreaLocation)
	if err == nil {
		ok = true
	}
	return
}

// Get the number of bytes that would be required to encode this date.
func EncodedSizeDate(time *Time) int {
	encodedYear := encodeYear(time.Year)
	return byteCountDate + getYearGroupCount(encodedYear, sizeDateYearUpperBits)
}

// Encode a date.
// Returns the number of bytes encoded, or the number of bytes it attempted to encode.
// Returns ok=true if there was enough room in dst.
func EncodeDate(time *Time, dst []byte) (bytesEncoded int, ok bool) {
	encodedYear := encodeYear(time.Year)
	yearGroupCount := getYearGroupCount(encodedYear, sizeDateYearUpperBits)
	yearGroupBitCount := yearGroupCount * bitsPerYearGroup
	yearGroupedMask := uint32(1<<uint(yearGroupBitCount) - 1)

	accumulator := uint16(encodedYear >> uint(yearGroupBitCount))
	accumulator = (accumulator << uint(sizeMonth)) + uint16(time.Month)
	accumulator = (accumulator << uint(sizeDay)) + uint16(time.Day)

	bytesEncoded = byteCountDate
	if bytesEncoded > len(dst) {
		return
	}
	encode16LE(accumulator, dst)
	byteCount := 0
	byteCount, ok = vlq.Rvlq(encodedYear & yearGroupedMask).EncodeTo(dst[bytesEncoded:])
	bytesEncoded += byteCount
	return
}

// Decode a date.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
// Returns ok=true if there was enough data in src.
// If ok == false, the resulting date is invalid.
func DecodeDate(src []byte) (time *Time, bytesDecoded int, ok bool, err error) {
	bytesDecoded = byteCountDate
	if bytesDecoded >= len(src) {
		return
	}

	time = new(Time)
	time.TimeIs = TypeDate
	time.TimezoneIs = TypeUTC
	accumulator := int(decode16LE(src))
	time.Day = accumulator & maskDay
	accumulator >>= sizeDay
	time.Month = accumulator & maskMonth
	accumulator >>= sizeMonth
	yearEncoded := vlq.Rvlq(accumulator)

	byteCount := 0
	byteCount, ok = yearEncoded.DecodeFrom(src[bytesDecoded:])
	bytesDecoded += byteCount
	time.Year = decodeYear(uint32(yearEncoded))

	err = validateDate(time)
	if err != nil {
		return
	}

	ok = true
	return
}

// Get the number of bytes that would be required to encode this time value.
func EncodedSizeTime(time *Time) int {
	magnitude := getSubsecondMagnitude(time.Nanosecond)
	baseByteCount := getBaseByteCount(baseSizeTime, magnitude)

	return baseByteCount + encodedSizeTimezone(time)
}

// Encode a time value.
// Returns the number of bytes encoded, or the number of bytes it attempted to encode.
// Returns ok=true if there was enough room in dst.
// Returns an error if something went wrong other than there not being enough room.
func EncodeTime(time *Time, dst []byte) (bytesEncoded int, ok bool, err error) {
	magnitude := getSubsecondMagnitude(time.Nanosecond)
	subsecond := time.Nanosecond / subsecMultipliers[magnitude]

	accumulator := uint64(subsecond)
	accumulator = (accumulator << uint(sizeSecond)) + uint64(time.Second)
	accumulator = (accumulator << uint(sizeMinute)) + uint64(time.Minute)
	accumulator = (accumulator << uint(sizeHour)) + uint64(time.Hour)
	accumulator = (accumulator << uint(sizeMagnitude)) + uint64(magnitude)
	accumulator <<= 1
	if time.TimezoneIs == TypeUTC {
		accumulator += 1
	}

	bytesEncoded = getBaseByteCount(baseSizeTime, magnitude)
	if bytesEncoded > len(dst) {
		ok = false
		return
	}

	encodeLE(accumulator, dst, bytesEncoded)
	byteCount := 0
	byteCount, ok, err = encodeTimezone(time, dst[bytesEncoded:])
	bytesEncoded += byteCount
	return
}

// Decode a time value.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
// Returns ok=true if there was enough data in src.
// Returns an error if something went wrong other than there not being enough data.
// If ok == false or err != nil, the resulting time value is invalid.
func DecodeTime(src []byte) (time *Time, bytesDecoded int, ok bool, err error) {
	if len(src) < 1 {
		bytesDecoded = 1
		return
	}

	magnitude := int((src[0] >> 1) & maskMagnitude)
	bytesDecoded = getBaseByteCount(baseSizeTime, magnitude)
	if bytesDecoded > len(src) {
		return
	}

	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubsecond := uint(sizeSubsecond * magnitude)
	maskSubsecond := (1 << sizeSubsecond) - 1

	time = new(Time)
	time.TimeIs = TypeTime
	accumulator := decodeLE(src, bytesDecoded)
	isUTC := accumulator&1 == 1
	accumulator >>= 1
	accumulator >>= sizeMagnitude
	time.Hour = int(accumulator & maskHour)
	accumulator >>= sizeHour
	time.Minute = int(accumulator & maskMinute)
	accumulator >>= sizeMinute
	time.Second = int(accumulator & maskSecond)
	accumulator >>= sizeSecond
	time.Nanosecond = (int(accumulator) & maskSubsecond) * subsecondMultiplier

	err = validateTime(time)
	if err != nil {
		return
	}

	byteCount := 0
	byteCount, ok, err = decodeTimezone(src[bytesDecoded:], time, isUTC)
	bytesDecoded += byteCount
	return
}

// Get the number of bytes that would be required to encode this timestamp.
func EncodedSizeTimestamp(time *Time) int {
	magnitude := getSubsecondMagnitude(time.Nanosecond)
	baseByteCount := getBaseByteCount(baseSizeTimestamp, magnitude)
	encodedYear := encodeYear(time.Year)
	yearGroupCount := getYearGroupCount(encodedYear<<1, timestampYearUpperBits[magnitude])

	return baseByteCount + yearGroupCount + encodedSizeTimezone(time)
}

// Encode a timestamp.
// Returns the number of bytes encoded, or the number of bytes it attempted to encode.
// Returns ok=true if there was enough room in dst.
// Returns an error if something went wrong other than there not being enough room.
func EncodeTimestamp(time *Time, dst []byte) (bytesEncoded int, ok bool, err error) {
	magnitude := getSubsecondMagnitude(time.Nanosecond)
	encodedYear := encodeYear(time.Year) << 1
	if time.TimezoneIs == TypeUTC {
		encodedYear |= 1
	}
	yearGroupCount := getYearGroupCount(encodedYear, timestampYearUpperBits[magnitude])
	yearGroupBitCount := yearGroupCount * bitsPerYearGroup
	yearGroupedMask := uint32(1<<uint(yearGroupBitCount) - 1)
	subsecond := time.Nanosecond / subsecMultipliers[magnitude]

	accumulator := uint64(encodedYear) >> uint(yearGroupBitCount)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) + uint64(subsecond)
	accumulator = (accumulator << uint(sizeMonth)) + uint64(time.Month)
	accumulator = (accumulator << uint(sizeDay)) + uint64(time.Day)
	accumulator = (accumulator << uint(sizeHour)) + uint64(time.Hour)
	accumulator = (accumulator << uint(sizeMinute)) + uint64(time.Minute)
	accumulator = (accumulator << uint(sizeSecond)) + uint64(time.Second)
	accumulator = (accumulator << uint(sizeMagnitude)) + uint64(magnitude)

	bytesEncoded = getBaseByteCount(baseSizeTimestamp, magnitude)
	if bytesEncoded > len(dst) {
		return
	}
	encodeLE(accumulator, dst, bytesEncoded)

	byteCount := 0
	byteCount, ok = vlq.Rvlq(encodedYear & yearGroupedMask).EncodeTo(dst[bytesEncoded:])
	bytesEncoded += byteCount
	if !ok {
		return
	}

	byteCount, ok, err = encodeTimezone(time, dst[bytesEncoded:])
	bytesEncoded += byteCount
	return
}

// Decode a timestamp.
// Returns the number of bytes decoded, or the number of bytes it attempted to decode.
// Returns ok=true if there was enough data in src.
// Returns an error if something went wrong other than there not being enough data.
// If ok == false or err != nil, the resulting timestamp is invalid.
func DecodeTimestamp(src []byte) (time *Time, bytesDecoded int, ok bool, err error) {
	if len(src) < 1 {
		bytesDecoded = 1
		return
	}

	magnitude := int(src[0] & maskMagnitude)
	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubsecond := uint(sizeSubsecond * magnitude)
	maskSubsecond := (1 << sizeSubsecond) - 1

	bytesDecoded = getBaseByteCount(baseSizeTimestamp, magnitude)
	if bytesDecoded > len(src) {
		return
	}

	time = new(Time)
	time.TimeIs = TypeTimestamp
	accumulator := decodeLE(src, bytesDecoded)
	accumulator >>= sizeMagnitude
	time.Second = int(accumulator & maskSecond)
	accumulator >>= sizeSecond
	time.Minute = int(accumulator & maskMinute)
	accumulator >>= sizeMinute
	time.Hour = int(accumulator & maskHour)
	accumulator >>= sizeHour
	time.Day = int(accumulator & maskDay)
	accumulator >>= sizeDay
	time.Month = int(accumulator & maskMonth)
	accumulator >>= sizeMonth
	time.Nanosecond = (int(accumulator) & maskSubsecond) * subsecondMultiplier
	accumulator >>= sizeSubsecond
	yearEncoded := vlq.Rvlq(accumulator)

	byteCount := 0
	byteCount, ok = yearEncoded.DecodeFrom(src[bytesDecoded:])
	bytesDecoded += byteCount
	if !ok {
		return
	}

	isUTC := yearEncoded&1 == 1
	yearEncoded >>= 1
	time.Year = decodeYear(uint32(yearEncoded))

	err = validateDate(time)
	if err != nil {
		return
	}
	err = validateTime(time)
	if err != nil {
		return
	}

	byteCount, ok, err = decodeTimezone(src[bytesDecoded:], time, isUTC)
	bytesDecoded += byteCount
	return
}

func AsCompactTime(src gotime.Time) (result *Time) {
	result = NewTimestamp(src.Year(), int(src.Month()), src.Day(), src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location().String())
	if src.Location() == gotime.Local {
		result.TimezoneIs = TypeAreaLocation
		result.AreaLocation = "Local"
	}
	return result
}

func AsGoTime(src *Time) (result gotime.Time, err error) {
	location := gotime.UTC
	switch src.TimezoneIs {
	case TypeUTC:
	// Nothing to do
	case TypeLatitudeLongitude:
		err = fmt.Errorf("Latitude/Longitude time zones are not supported by time.Time")
		return
	case TypeAreaLocation:
		location, err = gotime.LoadLocation(src.AreaLocation)
		if err != nil {
			return
		}
	}
	result = gotime.Date(src.Year, gotime.Month(src.Month), src.Day, src.Hour, src.Minute, src.Second, src.Nanosecond, location)
	return
}

func NewDate(year, month, day int) *Time {
	this := new(Time)
	this.TimeIs = TypeDate
	this.Year = year
	this.Month = month
	this.Day = day
	return this
}

// Create a new time. If areaLocation is empty, UTC is assumed.
func NewTime(hour, minute, second, nanosecond int, areaLocation string) *Time {
	this := new(Time)
	this.TimeIs = TypeTime
	this.Hour = hour
	this.Minute = minute
	this.Second = second
	this.Nanosecond = nanosecond
	this.AreaLocation = areaLocation
	if len(areaLocation) == 0 {
		this.TimezoneIs = TypeUTC
	} else {
		this.TimezoneIs = TypeAreaLocation
	}
	return this
}

func NewTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) *Time {
	this := new(Time)
	this.TimeIs = TypeTime
	this.Hour = hour
	this.Minute = minute
	this.Second = second
	this.Nanosecond = nanosecond
	this.LatitudeHundredths = latitudeHundredths
	this.LongitudeHundredths = longitudeHundredths
	this.TimezoneIs = TypeLatitudeLongitude
	return this
}

// Create a new timestamp. If areaLocation is empty, UTC is assumed.
func NewTimestamp(year, month, day, hour, minute, second, nanosecond int, areaLocation string) *Time {
	this := new(Time)
	this.TimeIs = TypeTimestamp
	this.Year = year
	this.Month = month
	this.Day = day
	this.Hour = hour
	this.Minute = minute
	this.Second = second
	this.Nanosecond = nanosecond
	this.AreaLocation = areaLocation
	if len(areaLocation) == 0 {
		this.TimezoneIs = TypeUTC
	} else {
		this.TimezoneIs = TypeAreaLocation
	}
	return this
}

func NewTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) *Time {
	this := new(Time)
	this.TimeIs = TypeTimestamp
	this.Year = year
	this.Month = month
	this.Day = day
	this.Hour = hour
	this.Minute = minute
	this.Second = second
	this.Nanosecond = nanosecond
	this.LatitudeHundredths = latitudeHundredths
	this.LongitudeHundredths = longitudeHundredths
	this.TimezoneIs = TypeLatitudeLongitude
	return this
}

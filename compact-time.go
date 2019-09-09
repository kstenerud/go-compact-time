package compact_time

import (
	"fmt"
	"strings"
	"time"

	"github.com/kstenerud/go-vlq"
)

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
const sizeLatitude = 14
const sizeLongitude = 15
const sizeDateYearUpperBits = 7

const baseSizeTime = sizeUtc + sizeMagnitude + sizeSecond + sizeMinute + sizeHour
const baseSizeTimestamp = sizeMagnitude + sizeSecond + sizeMinute + sizeHour + sizeDay + sizeMonth

const byteCountDate = 2

const maskMagnitude = ((1 << sizeMagnitude) - 1)
const maskSecond = ((1 << sizeSecond) - 1)
const maskMinute = ((1 << sizeMinute) - 1)
const maskHour = ((1 << sizeHour) - 1)
const maskDay = ((1 << sizeDay) - 1)
const maskMonth = ((1 << sizeMonth) - 1)
const maskLatitude = ((1 << sizeLatitude) - 1)
const maskLongitude = ((1 << sizeLongitude) - 1)
const maskDateYearUpperBits = ((1 << sizeDateYearUpperBits) - 1)

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

func getSubsecondMagnitude(time time.Time) int {
	if time.Nanosecond() == 0 {
		return 0
	}
	if (time.Nanosecond() % 1000) != 0 {
		return 3
	}
	if (time.Nanosecond() % 1000000) != 0 {
		return 2
	}
	return 1
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

func encodeYearAndUtcFlag(thetime time.Time) uint32 {
	utcflag := uint32(0)
	if thetime.Location() == time.UTC {
		utcflag = 1
	}
	return (encodeYear(thetime.Year()) << 1) | utcflag
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

func writeLocationString(location string, dst []byte) (bytesEncoded int, err error) {
	location = getAbbreviatedTimezoneString(location)
	byteCount := len(location) + 1
	if len(dst) < byteCount {
		return 0, fmt.Errorf("Require %v bytes to store location [%v], but only %v bytes available", byteCount, location, len(dst))
	}
	dst[0] = byte(len(location) << 1)
	copy(dst[1:], location)
	return byteCount, nil
}

func encodeTimezone(location *time.Location, dst []byte) (bytesEncoded int, err error) {
	// TODO: lat-long support?
	switch location {
	case time.UTC:
		return 0, nil
	case time.Local:
		return writeLocationString("L", dst)
	default:
		_, err := time.LoadLocation(location.String())
		if err != nil {
			return 0, fmt.Errorf("%v is not an IANA time zone, or time zone database not found", location)
		}
		return writeLocationString(location.String(), dst)
	}
}

func getFullTimezoneString(tz string) string {
	if len(tz) == 0 {
		return tz
	}
	firstChar := rune(tz[0])

	if len(tz) == 1 {
		switch firstChar {
		case 'l':
			return "Local"
		case 'z':
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

func decodeTimezone(src []byte, timezoneIsUtc bool) (location *time.Location, bytesDecoded int, err error) {
	if timezoneIsUtc {
		return time.UTC, 0, nil
	}

	if len(src) < 1 {
		return nil, 0, fmt.Errorf("Require %v bytes to read location, but only %v bytes available", 1, len(src))
	}

	isLatlong := src[0] & 1
	if isLatlong == 1 {
		return nil, 0, fmt.Errorf("TODO: latlong not supported")
	}

	offset := 0
	length := int(src[offset] >> 1)
	offset++
	if offset+length > len(src) {
		return nil, 0, fmt.Errorf("Require %v bytes to read location, but only %v bytes available", length, len(src))
	}
	name := string(src[offset : offset+length])
	offset += length
	location, err = time.LoadLocation(getFullTimezoneString(name))
	if err != nil {
		return nil, 0, err
	}
	return location, offset, nil
}

func timezoneEncodedSize(location *time.Location) int {
	// TODO: lat-long support?
	if location == time.UTC {
		return 0
	}
	if location == time.Local {
		return 2 // length 1 + "L"
	}
	return 1 + len(getAbbreviatedTimezoneString(location.String()))
}

func EncodedSize(time time.Time) int {
	magnitude := getSubsecondMagnitude(time)
	baseByteCount := getBaseByteCount(baseSizeTimestamp, magnitude)
	encodedYear := encodeYear(time.Year())
	yearGroupCount := getYearGroupCount(encodedYear<<1, timestampYearUpperBits[magnitude])

	return baseByteCount + yearGroupCount + timezoneEncodedSize(time.Location())
}

func encodeLE(value uint64, dst []byte, byteCount int) {
	for i := 0; i < byteCount; i++ {
		dst[i] = uint8(value)
		value >>= 8
	}
}

func decodeLE(src []byte, byteCount int) uint64 {
	accumulator := uint64(0)
	for i := 0; i < byteCount; i++ {
		accumulator |= uint64(src[i]) << (uint(i) * 8)
	}
	return accumulator
}

func EncodeTimestamp(time time.Time, dst []byte) (bytesEncoded int, err error) {
	magnitude := getSubsecondMagnitude(time)
	encodedYear := encodeYearAndUtcFlag(time)
	yearGroupCount := getYearGroupCount(encodedYear, timestampYearUpperBits[magnitude])
	yearGroupBitCount := yearGroupCount * bitsPerYearGroup
	yearGroupedMask := uint32(1<<uint(yearGroupBitCount) - 1)
	subsecond := time.Nanosecond() / subsecMultipliers[magnitude]

	accumulator := uint64(encodedYear) >> uint(yearGroupBitCount)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) + uint64(subsecond)
	accumulator = (accumulator << uint(sizeMonth)) + uint64(time.Month())
	accumulator = (accumulator << uint(sizeDay)) + uint64(time.Day())
	accumulator = (accumulator << uint(sizeHour)) + uint64(time.Hour())
	accumulator = (accumulator << uint(sizeMinute)) + uint64(time.Minute())
	accumulator = (accumulator << uint(sizeSecond)) + uint64(time.Second())
	accumulator = (accumulator << uint(sizeMagnitude)) + uint64(magnitude)

	offset := 0
	accumulatorSize := getBaseByteCount(baseSizeTimestamp, magnitude)
	if accumulatorSize > len(dst) {
		return 0, fmt.Errorf("Require %v bytes to store [%v], but only %v bytes available", accumulatorSize, time, len(dst))
	}
	encodeLE(accumulator, dst[offset:], accumulatorSize)
	offset += accumulatorSize

	bytesEncoded, err = vlq.Rvlq(encodedYear & yearGroupedMask).EncodeTo(dst[offset:])
	if err != nil {
		return 0, err
	}
	offset += bytesEncoded

	bytesEncoded, err = encodeTimezone(time.Location(), dst[offset:])
	if err != nil {
		return 0, err
	}
	offset += bytesEncoded

	return offset, nil
}

func DecodeTimestamp(src []byte) (result time.Time, bytesDecoded int, err error) {
	if len(src) < 1 {
		return result, 0, fmt.Errorf("Require %v bytes to decode timestamp, but only %v available", 1, len(src))
	}

	magnitude := int(src[0] & maskMagnitude)
	subsecondMultiplier := subsecMultipliers[magnitude]
	sizeSubsecond := uint(sizeSubsecond * magnitude)
	maskSubsecond := (1 << sizeSubsecond) - 1

	offset := getBaseByteCount(baseSizeTimestamp, magnitude)
	if offset > len(src) {
		return result, 0, fmt.Errorf("Require %v bytes to decode timestamp, but only %v available", offset, len(src))
	}

	accumulator := decodeLE(src, offset)
	accumulator >>= sizeMagnitude
	second := int(accumulator & maskSecond)
	accumulator >>= sizeSecond
	minute := int(accumulator & maskMinute)
	accumulator >>= sizeMinute
	hour := int(accumulator & maskHour)
	accumulator >>= sizeHour
	day := int(accumulator & maskDay)
	accumulator >>= sizeDay
	month := time.Month(accumulator & maskMonth)
	accumulator >>= sizeMonth
	nanosecond := (int(accumulator) & maskSubsecond) * subsecondMultiplier
	accumulator >>= sizeSubsecond
	yearEncoded := vlq.Rvlq(accumulator)

	isComplete := false
	bytesDecoded, isComplete = yearEncoded.DecodeFrom(src[offset:])
	if !isComplete {
		return result, 0, fmt.Errorf("Not enough data to decode RVLQ")
	}
	offset += bytesDecoded

	timezoneIsUtc := yearEncoded&1 == 1
	yearEncoded >>= 1
	year := decodeYear(uint32(yearEncoded))

	location, bytesDecoded, err := decodeTimezone(src[offset:], timezoneIsUtc)
	if err != nil {
		return result, 0, err
	}
	offset += bytesDecoded

	result = time.Date(year, month, day, hour, minute, second, nanosecond, location)

	return result, offset, nil
}

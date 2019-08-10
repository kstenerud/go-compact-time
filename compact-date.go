package compact_date

import (
	"fmt"
	"time"

	"github.com/kstenerud/go-vlq"
)

const yearBias = 2000
const bitsPerYearGroup = 7

const sizeMagnitude = 2
const sizeSubsecond = 10
const sizeSecond = 6
const sizeMinute = 6
const sizeHour = 5
const sizeDay = 5
const sizeMonth = 4

const maskSecond = ((1 << sizeSecond) - 1)
const maskMinute = ((1 << sizeMinute) - 1)
const maskHour = ((1 << sizeHour) - 1)
const maskDay = ((1 << sizeDay) - 1)
const maskMonth = ((1 << sizeMonth) - 1)

var baseSizes = [...]int{4, 5, 6, 8}
var yearHighBits = [...]int{4, 2, 0, 6}
var subsecMultipliers = [...]int{1, 1000000, 1000, 1}

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

func decodeYear(encodedYear uint32) int {
	return int(zigzagDecode(uint32(encodedYear))) + yearBias
}

func getBaseByteCount(magnitude int) int {
	return (sizeMagnitude + sizeSubsecond*magnitude + sizeSecond + sizeMinute + sizeHour + sizeDay + sizeMonth + yearHighBits[magnitude]) / 8
}

func getYearGroupCount(encodedYear uint32, subsecondMagnitude int) int {
	extraBitCount := yearHighBits[subsecondMagnitude]
	year := encodedYear >> bitsPerYearGroup
	if year == 0 {
		return 1
	}

	size := 1
	for year != 0 {
		size++
		year >>= bitsPerYearGroup
	}

	extraMask := (uint32(1) << uint(extraBitCount)) - 1
	lastGroupBits := encodedYear >> uint(bitsPerYearGroup*(size-1))
	if lastGroupBits & ^extraMask != 0 {
		return size
	}
	return size - 1
}

func EncodedSize(time time.Time) int {
	magnitude := getSubsecondMagnitude(time)
	baseByteCount := getBaseByteCount(magnitude)
	encodedYear := encodeYear(time.Year())
	yearGroupCount := getYearGroupCount(encodedYear, magnitude)

	return baseByteCount + yearGroupCount
}

func Encode(time time.Time, dst []byte) (bytesEncoded int, err error) {
	magnitude := getSubsecondMagnitude(time)
	baseByteCount := getBaseByteCount(magnitude)
	encodedYear := encodeYear(time.Year())
	yearGroupCount := getYearGroupCount(encodedYear, magnitude)

	if baseByteCount+yearGroupCount > len(dst) {
		return 0, fmt.Errorf("Require %v bytes to store [%v], but only %v bytes available", baseByteCount+yearGroupCount, time, len(dst))
	}

	subsecond := time.Nanosecond() / subsecMultipliers[magnitude]
	yearGroupBitCount := yearGroupCount * bitsPerYearGroup
	yearGroupedMask := (1 << uint(yearGroupBitCount)) - 1
	// KSLOG_TRACE("subs %d, year group bits %d, year mask %x", subsecond, yearGroupBitCount, yearGroupedMask);

	// int b = 0;
	// KSLOG_TRACE("y: %016lx %02d %d", (uint64_t)(encodedYear >> yearGroupBitCount) << b, b, encodedYear >> yearGroupBitCount); b += yearHighBits[magnitude];
	// KSLOG_TRACE("M: %016lx %02d %d", (uint64_t)date->month << b, b, date->month); b += SIZE_MONTH;
	// KSLOG_TRACE("d: %016lx %02d %d", (uint64_t)date->day << b, b, date->day); b += SIZE_DAY;
	// KSLOG_TRACE("h: %016lx %02d %d", (uint64_t)date->hour << b, b, date->hour); b += SIZE_HOUR;
	// KSLOG_TRACE("m: %016lx %02d %d", (uint64_t)date->minute << b, b, date->minute); b += SIZE_MINUTE;
	// KSLOG_TRACE("s: %016lx %02d %d", (uint64_t)date->second << b, b, date->second); b += SIZE_SECOND;
	// KSLOG_TRACE("S: %016lx %02d %d", (uint64_t)subsecond << b, b, subsecond); b += SIZE_SUBSECOND * magnitude;
	// KSLOG_TRACE("a: %016lx %02d %d", (uint64_t)magnitude << b, b, magnitude);

	accumulator := uint64(magnitude)
	accumulator = (accumulator << uint(sizeSubsecond*magnitude)) + uint64(subsecond)
	accumulator = (accumulator << uint(sizeSecond)) + uint64(time.Second())
	accumulator = (accumulator << uint(sizeMinute)) + uint64(time.Minute())
	accumulator = (accumulator << uint(sizeHour)) + uint64(time.Hour())
	accumulator = (accumulator << uint(sizeDay)) + uint64(time.Day())
	accumulator = (accumulator << uint(sizeMonth)) + uint64(time.Month())
	accumulator = (accumulator << uint(yearHighBits[magnitude])) + uint64(encodedYear>>uint(yearGroupBitCount))

	// KSLOG_DEBUG("Accumulator: %016lx", (uint64_t)accumulator);

	encodedYear &= uint32(yearGroupedMask)

	offset := 0
	for i := baseByteCount - 1; i >= 0; i-- {
		//     KSLOG_TRACE("Write %02x", (uint8_t)(accumulator >> (8*i)));
		dst[offset] = uint8(accumulator >> uint(8*i))
		offset++
	}

	// KSLOG_TRACE("encoded year %d (%02x)", encodedYear, encodedYear);
	bytesEncoded, err = vlq.Rvlq(encodedYear).EncodeTo(dst[offset:])
	if err != nil {
		return 0, err
	}
	return bytesEncoded + offset, nil
}

func Decode(src []byte) (bytesDecoded int, result time.Time, err error) {
	// KSLOG_DEBUG("decode");
	if len(src) < 1 {
		return 0, result, fmt.Errorf("Destination buffer has length 0")
	}

	shiftMagnitude := 6
	maskMagnitude := (1 << uint(shiftMagnitude)) - 1
	// KSLOG_TRACE("mask mag %02x", maskMagnitude);
	nextByte := src[0]
	//     KSLOG_TRACE("Read %d: %02x", 0, src[0]);
	srcIndex := 1

	magnitude := nextByte >> uint(shiftMagnitude)
	nextByte &= uint8(maskMagnitude)
	// KSLOG_TRACE("next byte masked %02x: %02x", ~maskMagnitude, nextByte);

	remainingBytes := baseSizes[magnitude] - 1
	srcIndexEnd := srcIndex + remainingBytes
	// KSLOG_TRACE("rem bytes %d, src index %d, src length %d", remainingBytes, srcIndex, srcLength);
	if srcIndexEnd >= len(src) {
		return 0, result, fmt.Errorf("Require %v bytes to decode compact date, but only %v bytes available", srcIndexEnd, len(src))
	}

	accumulator := uint64(nextByte)
	// KSLOG_TRACE("Accum %016lx", accumulator);
	for srcIndex < srcIndexEnd {
		// KSLOG_TRACE("Read %d: %02x", srcIndex, src[srcIndex]);
		accumulator = (accumulator << 8) | uint64(src[srcIndex])
		srcIndex++
		// KSLOG_TRACE("Accum %016lx", accumulator);
	}

	yearHighBits := yearHighBits[magnitude]
	yearHighBitsMask := (1 << uint(yearHighBits)) - 1

	yearEncoded := uint(accumulator & uint64(yearHighBitsMask))
	accumulator >>= uint(yearHighBits)
	month := int(accumulator & maskMonth)
	accumulator >>= uint(sizeMonth)
	day := int(accumulator & uint64(maskDay))
	accumulator >>= uint(sizeDay)
	hour := int(accumulator & uint64(maskHour))
	accumulator >>= uint(sizeHour)
	minute := int(accumulator & uint64(maskMinute))
	accumulator >>= uint(sizeMinute)
	second := int(accumulator & uint64(maskSecond))
	accumulator >>= uint(sizeSecond)
	nanosecond := int(accumulator * uint64(subsecMultipliers[magnitude]))

	yearEncodedRvlq := vlq.Rvlq(yearEncoded)
	isComplete := false
	bytesDecoded, isComplete = yearEncodedRvlq.DecodeFrom(src[srcIndex:])
	if !isComplete {
		return 0, result, fmt.Errorf("Not enough bytes to decode this compact date")
	}

	year := decodeYear(uint32(yearEncodedRvlq))

	result = time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC)
	bytesDecoded += srcIndex

	return bytesDecoded, result, err
}

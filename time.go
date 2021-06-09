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

package compact_time

import (
	"fmt"
	"strings"
	gotime "time"
)

type TimeType uint8

const (
	TimeTypeDate = TimeType(iota)
	TimeTypeTime
	TimeTypeTimestamp
)

type TimezoneType uint8

const (
	TimezoneTypeUnset = TimezoneType(iota)
	TimezoneTypeUTC
	TimezoneTypeLocal
	TimezoneTypeAreaLocation
	TimezoneTypeLatitudeLongitude
	TimezoneTypeUTCOffset
)

type Timezone struct {
	ShortAreaLocation    string
	LongAreaLocation     string
	LatitudeHundredths   int16
	LongitudeHundredths  int16
	MinutesOffsetFromUTC int16
	Type                 TimezoneType
}

var (
	timezoneUTC = Timezone{
		Type:              TimezoneTypeUTC,
		ShortAreaLocation: "Z",
		LongAreaLocation:  "Etc/UTC",
	}
	timezoneLocal = Timezone{
		Type:              TimezoneTypeLocal,
		ShortAreaLocation: "L",
		LongAreaLocation:  "Local",
	}
)

func TZAtUTC() Timezone {
	return timezoneUTC
}

func TZLocal() Timezone {
	return timezoneLocal
}

func TZAtAreaLocation(areaLocation string) Timezone {
	var this Timezone
	this.InitWithAreaLocation(areaLocation)
	return this
}

func TZAtLatLong(latitudeHundredths, longitudeHundredths int) Timezone {
	var this Timezone
	this.InitWithLatLong(latitudeHundredths, longitudeHundredths)
	return this
}

func TZWithMiutesOffsetFromUTC(minutesOffsetFromUTC int) Timezone {
	var this Timezone
	this.InitWithMinutesOffsetFromUTC(minutesOffsetFromUTC)
	return this
}

func (this *Timezone) InitWithAreaLocation(areaLocation string) {
	switch areaLocationToTimezoneType[areaLocation] {
	case internalTZUTC:
		*this = timezoneUTC
	case internalTZLocal:
		*this = timezoneLocal
	case internalTZUTCPreserve:
		this.Type = TimezoneTypeUTC
		this.ShortAreaLocation = "Z"
		this.LongAreaLocation = areaLocation
	default:
		this.Type = TimezoneTypeAreaLocation
		this.ShortAreaLocation, this.LongAreaLocation = splitAreaLocation(areaLocation)
	}
}

func (this *Timezone) InitWithLatLong(latitudeHundredths, longitudeHundredths int) {
	this.LatitudeHundredths = int16(latitudeHundredths)
	this.LongitudeHundredths = int16(longitudeHundredths)
	this.Type = TimezoneTypeLatitudeLongitude
}

func (this *Timezone) InitWithMinutesOffsetFromUTC(minutesOffsetFromUTC int) {
	minutes := int16(minutesOffsetFromUTC)
	if minutes == 0 {
		*this = timezoneUTC
	} else {
		this.MinutesOffsetFromUTC = minutes
		this.Type = TimezoneTypeUTCOffset
	}
}

func (this *Timezone) Validate() error {
	switch this.Type {
	case TimezoneTypeAreaLocation:
		length := len(this.LongAreaLocation)
		if length == 0 {
			return fmt.Errorf("Time zone is specified as area/location, but the AreaLocation field is empty")
		}
		if length > 127 {
			return fmt.Errorf("Area/location time zones cannot be over 127 bytes long")
		}
	case TimezoneTypeLatitudeLongitude:
		if this.LongitudeHundredths < longitudeMin || this.LongitudeHundredths > longitudeMax {
			return fmt.Errorf("%v: Invalid longitude (must be %v to %v)", this.LongitudeHundredths, longitudeMin, longitudeMax)
		}
		if this.LatitudeHundredths < latitudeMin || this.LatitudeHundredths > latitudeMax {
			return fmt.Errorf("%v: Invalid latitude (must be %v to %v)", this.LatitudeHundredths, latitudeMin, latitudeMax)
		}
	case TimezoneTypeUTCOffset:
		if this.MinutesOffsetFromUTC < minutesFromUTCMin || this.MinutesOffsetFromUTC > minutesFromUTCMax {
			return fmt.Errorf("%v: Invalid UTC offset", this.MinutesOffsetFromUTC)
		}
	}
	return nil
}

func (this *Timezone) IsEquivalentTo(that *Timezone) bool {
	if this.Type != that.Type {
		return false
	}
	switch this.Type {
	case TimezoneTypeAreaLocation:
		return this.ShortAreaLocation == that.ShortAreaLocation && this.LongAreaLocation == that.LongAreaLocation
	case TimezoneTypeLatitudeLongitude:
		return this.LatitudeHundredths == that.LatitudeHundredths && this.LongitudeHundredths == that.LongitudeHundredths
	case TimezoneTypeUTCOffset:
		return this.MinutesOffsetFromUTC == that.MinutesOffsetFromUTC
	}
	return true
}

func (this *Timezone) String() string {
	switch this.Type {
	case TimezoneTypeUTC:
		return ""
	case TimezoneTypeAreaLocation, TimezoneTypeLocal:
		return fmt.Sprintf("/%s", this.LongAreaLocation)
	case TimezoneTypeLatitudeLongitude:
		return fmt.Sprintf("/%.2f/%.2f", float64(this.LatitudeHundredths)/100, float64(this.LongitudeHundredths)/100)
	case TimezoneTypeUTCOffset:
		sign := '+'
		minute := int(this.MinutesOffsetFromUTC)
		if minute < 0 {
			sign = '-'
			minute = -minute
		}
		hour := minute / 60
		minute %= 60
		return fmt.Sprintf("%c%02d%02d", sign, hour, minute)
	default:
		return fmt.Sprintf("Error: %v: Unknown time zone type", this.Type)
	}
}

type Time struct {
	Timezone   Timezone
	Year       int
	Nanosecond uint32
	Second     uint8
	Minute     uint8
	Hour       uint8
	Day        uint8
	Month      uint8
	Type       TimeType
}

// Create a "zero" date, which will encode to all zeroes.
func ZeroDate() Time {
	return Time{Type: TimeTypeDate}
}

// Create a "zero" time, which will encode to all zeroes.
func ZeroTime() Time {
	return Time{Type: TimeTypeTime}
}

// Create a "zero" timestamp, which will encode to all zeroes.
func ZeroTimestamp() Time {
	return Time{Type: TimeTypeTimestamp}
}

func NewDate(year, month, day int) Time {
	var this Time
	this.InitDate(year, month, day)
	return this
}

func (this *Time) InitDate(year, month, day int) {
	this.Type = TimeTypeDate
	this.Year = year
	this.Month = uint8(month)
	this.Day = uint8(day)
	this.Timezone.Type = TimezoneTypeLocal
}

func NewTime(hour, minute, second, nanosecond int, timezone Timezone) Time {
	var this Time
	this.InitTime(hour, minute, second, nanosecond, timezone)
	return this
}

func (this *Time) InitTime(hour, minute, second, nanosecond int, timezone Timezone) {
	this.Type = TimeTypeTime
	this.Hour = uint8(hour)
	this.Minute = uint8(minute)
	this.Second = uint8(second)
	this.Nanosecond = uint32(nanosecond)
	this.Timezone = timezone
}

func NewTimestamp(year, month, day, hour, minute, second, nanosecond int, timezone Timezone) Time {
	var this Time
	this.InitTimestamp(year, month, day, hour, minute, second, nanosecond, timezone)
	return this
}

func (this *Time) InitTimestamp(year, month, day, hour, minute, second, nanosecond int, tz Timezone) {
	this.Year = year
	this.Month = uint8(month)
	this.Day = uint8(day)
	this.Hour = uint8(hour)
	this.Minute = uint8(minute)
	this.Second = uint8(second)
	this.Nanosecond = uint32(nanosecond)
	this.Timezone = tz
	this.Type = TimeTypeTimestamp
}

func (this *Time) IsZeroValue() bool {
	return this.Timezone.Type == TimezoneTypeUnset
}

// Check if two times are equivalent. This handles cases where the time zones
// are technically equivalent (Z == UTC == Etc/UTC == Etc/GMT, etc)
func (this *Time) IsEquivalentTo(that Time) bool {
	if this.Timezone.Type == TimezoneTypeUTC && that.Timezone.Type == TimezoneTypeUTC {
		return this.Year == that.Year &&
			this.Month == that.Month &&
			this.Day == that.Day &&
			this.Hour == that.Hour &&
			this.Minute == that.Minute &&
			this.Second == that.Second &&
			this.Nanosecond == that.Nanosecond
	}
	return *this == that
}

// Convert a golang time value to compact time
func AsCompactTime(src gotime.Time) Time {
	locationStr := src.Location().String()
	if src.Location() == gotime.Local {
		locationStr = "Local"
	}
	return NewTimestamp(src.Year(), int(src.Month()), src.Day(), src.Hour(),
		src.Minute(), src.Second(), src.Nanosecond(), TZAtAreaLocation(locationStr))
}

// Convert compact time into golang time.
// Note: Go time doesn't support latitude/longitude time zones. Attempting to
//       convert this type of time zone will result in an error.
// Note: Converting to go time will validate area/location time zone (if any)
func (this *Time) AsGoTime() (result gotime.Time, err error) {
	location := gotime.UTC
	switch this.Timezone.Type {
	case TimezoneTypeUTC:
		location = gotime.UTC
	case TimezoneTypeLocal:
		location = gotime.Local
	case TimezoneTypeLatitudeLongitude:
		err = fmt.Errorf("Latitude/Longitude time zones are not supported by time.Time")
		return
	case TimezoneTypeAreaLocation:
		location, err = gotime.LoadLocation(this.Timezone.LongAreaLocation)
		if err != nil {
			return
		}
	case TimezoneTypeUTCOffset:
		location = gotime.FixedZone("", int(this.Timezone.MinutesOffsetFromUTC)*60)
	default:
		err = fmt.Errorf("%v: Unknown time zone type", this.Timezone.Type)
		return
	}
	result = gotime.Date(this.Year,
		gotime.Month(this.Month),
		int(this.Day),
		int(this.Hour),
		int(this.Minute),
		int(this.Second),
		int(this.Nanosecond),
		location)
	return
}

func (this Time) String() string {
	// Workaround for go's broken Stringer type handling
	return this.pString()
}

func (this *Time) pString() string {
	if this.IsZeroValue() {
		return "<zero time value>"
	}
	switch this.Type {
	case TimeTypeDate:
		return this.formatDate()
	case TimeTypeTime:
		return this.formatTime()
	case TimeTypeTimestamp:
		return this.formatTimestamp()
	default:
		return fmt.Sprintf("Error: %v: Unknown time type", this.Type)
	}
}

func (this *Time) Validate() error {
	if this.Type == TimeTypeDate || this.Type == TimeTypeTimestamp {
		if this.Year == 0 {
			return fmt.Errorf("Year cannot be 0")
		}
		if this.Month < monthMin || this.Month > monthMax {
			return fmt.Errorf("%v: Invalid month (must be %v to %v)", this.Month, monthMin, monthMax)
		}
		if this.Day < dayMin || this.Day > dayMax[this.Month] {
			return fmt.Errorf("%v: Invalid day (must be %v to %v)", this.Day, dayMin, dayMax[this.Month])
		}
	}

	if this.Type == TimeTypeTime || this.Type == TimeTypeTimestamp {
		if this.Hour < hourMin || this.Hour > hourMax {
			return fmt.Errorf("%v: Invalid hour (must be %v to %v)", this.Hour, hourMin, hourMax)
		}
		if this.Minute < minuteMin || this.Minute > minuteMax {
			return fmt.Errorf("%v: Invalid minute (must be %v to %v)", this.Minute, minuteMin, minuteMax)
		}
		if this.Second < secondMin || this.Second > secondMax {
			return fmt.Errorf("%v: Invalid second (must be %v to %v)", this.Second, secondMin, secondMax)
		}
		if this.Nanosecond < nanosecondMin || this.Nanosecond > nanosecondMax {
			return fmt.Errorf("%v: Invalid nanosecond (must be %v to %v)", this.Nanosecond, nanosecondMin, nanosecondMax)
		}
		return this.Timezone.Validate()
	}

	return nil
}

// =============================================================================

func splitAreaLocation(areaLocation string) (shortAreaLocation, longAreaLocation string) {
	longAreaLocation = areaLocation
	tzPair := strings.SplitN(areaLocation, "/", 2)
	if len(tzPair) > 1 {
		area := tzPair[0]
		location := tzPair[1]
		if len(area) == 1 {
			shortAreaLocation = areaLocation
			if longArea := shortAreaToArea[area]; longArea != "" {
				longAreaLocation = longArea + "/" + location
			} else {
				longAreaLocation = areaLocation
			}
		} else {
			if shortArea := areaToShortArea[area]; shortArea != "" {
				shortAreaLocation = shortArea + "/" + location
			} else {
				shortAreaLocation = areaLocation
			}
		}
	} else {
		shortAreaLocation = areaLocation
	}
	return
}

func (this *Time) formatDate() string {
	return fmt.Sprintf("%d-%02d-%02d", this.Year, this.Month, this.Day)
}

func (this *Time) formatTime() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%02d:%02d:%02d", this.Hour, this.Minute, this.Second))
	if this.Nanosecond != 0 {
		str := []byte(fmt.Sprintf("%09d", this.Nanosecond))
		for str[len(str)-1] == '0' {
			str = str[:len(str)-1]
		}
		builder.WriteByte('.')
		builder.WriteString(string(str))
	}
	builder.WriteString(this.Timezone.String())
	return builder.String()
}

func (this *Time) formatTimestamp() string {
	var builder strings.Builder
	builder.WriteString(this.formatDate())
	builder.WriteByte('/')
	builder.WriteString(this.formatTime())
	return builder.String()
}

var shortAreaToArea = map[string]string{
	"F": "Africa",
	"M": "America",
	"N": "Antarctica",
	"R": "Arctic",
	"S": "Asia",
	"T": "Atlantic",
	"U": "Australia",
	"C": "Etc",
	"E": "Europe",
	"I": "Indian",
	"P": "Pacific",
	"L": "Local",
	"Z": "Zero",
}

var areaToShortArea = map[string]string{
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
	"Local":      "L",
	"Zero":       "Z",
}

const (
	monthMin          = 1
	monthMax          = 12
	dayMin            = 1
	hourMin           = 0
	hourMax           = 23
	minuteMin         = 0
	minuteMax         = 59
	secondMin         = 0
	secondMax         = 60
	nanosecondMin     = 0
	nanosecondMax     = 999999999
	latitudeMin       = -9000
	latitudeMax       = 9000
	longitudeMin      = -18000
	longitudeMax      = 18000
	minutesFromUTCMin = -1439
	minutesFromUTCMax = 1439
)

var dayMax = [...]uint8{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

type internalTZType int

const (
	internalTZAreaLocation = iota
	internalTZUTC
	internalTZUTCPreserve
	internalTZLocal
)

var areaLocationToTimezoneType = map[string]internalTZType{
	"":              internalTZUTC,
	"Etc/UTC":       internalTZUTC,
	"Z":             internalTZUTC,
	"Zero":          internalTZUTC,
	"Etc/GMT":       internalTZUTCPreserve,
	"Etc/GMT+0":     internalTZUTCPreserve,
	"Etc/GMT-0":     internalTZUTCPreserve,
	"Etc/GMT0":      internalTZUTCPreserve,
	"Etc/Greenwich": internalTZUTCPreserve,
	"Etc/UCT":       internalTZUTCPreserve,
	"Etc/Universal": internalTZUTCPreserve,
	"Etc/Zulu":      internalTZUTCPreserve,
	"Factory":       internalTZUTCPreserve,
	"GMT":           internalTZUTCPreserve,
	"GMT+0":         internalTZUTCPreserve,
	"GMT-0":         internalTZUTCPreserve,
	"GMT0":          internalTZUTCPreserve,
	"Greenwich":     internalTZUTCPreserve,
	"UCT":           internalTZUTCPreserve,
	"Universal":     internalTZUTCPreserve,
	"UTC":           internalTZUTCPreserve,
	"Zulu":          internalTZUTCPreserve,
	"L":             internalTZLocal,
	"Local":         internalTZLocal,
}

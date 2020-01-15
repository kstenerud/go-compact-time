package compact_time

import (
	"fmt"
	gotime "time"
)

const (
	monthMin      = 1
	monthMax      = 12
	dayMin        = int8(1)
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

var dayMax = [...]int8{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

type TimeType int8

const (
	TypeDate = TimeType(iota)
	TypeTime
	TypeTimestamp
)

type TimezoneType int8

const (
	TypeUTC = TimezoneType(iota)
	TypeAreaLocation
	TypeLatitudeLongitude
)

type Time struct {
	Year                int
	Nanosecond          int32
	LatitudeHundredths  int16
	LongitudeHundredths int16
	Month               int8
	Day                 int8
	Hour                int8
	Minute              int8
	Second              int8
	TimeIs              TimeType
	TimezoneIs          TimezoneType
	AreaLocation        string
}

func (this *Time) initTimeCommon(hour, minute, second, nanosecond int) {
	this.Hour = int8(hour)
	this.Minute = int8(minute)
	this.Second = int8(second)
	this.Nanosecond = int32(nanosecond)
	this.TimeIs = TypeTime
}

func NewDate(year, month, day int) *Time {
	this := new(Time)
	this.InitDate(year, month, day)
	return this
}

func (this *Time) InitDate(year, month, day int) {
	this.Year = year
	this.Month = int8(month)
	this.Day = int8(day)
	this.TimeIs = TypeDate
}

// Create a new time. If areaLocation is empty, UTC is assumed.
func NewTime(hour, minute, second, nanosecond int, areaLocation string) *Time {
	this := new(Time)
	this.InitTime(hour, minute, second, nanosecond, areaLocation)
	return this
}

// Init a time. If areaLocation is empty, UTC is assumed.
func (this *Time) InitTime(hour, minute, second, nanosecond int, areaLocation string) {
	this.initTimeCommon(hour, minute, second, nanosecond)
	this.AreaLocation = areaLocation
	if len(areaLocation) == 0 {
		this.TimezoneIs = TypeUTC
	} else {
		this.TimezoneIs = TypeAreaLocation
	}
}

func NewTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) *Time {
	this := new(Time)
	this.InitTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	return this
}

func (this *Time) InitTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) {
	this.initTimeCommon(hour, minute, second, nanosecond)
	this.LatitudeHundredths = int16(latitudeHundredths)
	this.LongitudeHundredths = int16(longitudeHundredths)
	this.TimezoneIs = TypeLatitudeLongitude
}

// Create a new timestamp. If areaLocation is empty, UTC is assumed.
func NewTimestamp(year, month, day, hour, minute, second, nanosecond int, areaLocation string) *Time {
	this := new(Time)
	this.InitTimestamp(year, month, day, hour, minute, second, nanosecond, areaLocation)
	return this
}

// Init a timestamp. If areaLocation is empty, UTC is assumed.
func (this *Time) InitTimestamp(year, month, day, hour, minute, second, nanosecond int, areaLocation string) {
	this.InitDate(year, month, day)
	this.InitTime(hour, minute, second, nanosecond, areaLocation)
	this.TimeIs = TypeTimestamp
}

func NewTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) *Time {
	this := new(Time)
	this.InitTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	return this
}

func (this *Time) InitTimestampLatLong(year, month, day, hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths int) {
	this.InitDate(year, month, day)
	this.InitTimeLatLong(hour, minute, second, nanosecond, latitudeHundredths, longitudeHundredths)
	this.TimeIs = TypeTimestamp
}

func AsCompactTime(src *gotime.Time) (result *Time) {
	result = NewTimestamp(src.Year(), int(src.Month()), src.Day(), src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location().String())
	if src.Location() == gotime.Local {
		result.AreaLocation = "Local"
	}
	return result
}

// Convert this time into a standard go time.
// Note: Go time doesn't support latitude/longitude time zones. Attempting to
//       convert this type of time zone will result in an error.
func (this *Time) AsGoTime() (result *gotime.Time, err error) {
	location := gotime.UTC
	switch this.TimezoneIs {
	case TypeUTC:
		location = gotime.UTC
	case TypeLatitudeLongitude:
		err = fmt.Errorf("Latitude/Longitude time zones are not supported by time.Time")
		return
	case TypeAreaLocation:
		if this.AreaLocation == "Local" {
			location = gotime.Local
		}
		location, err = gotime.LoadLocation(this.AreaLocation)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("%v: Unknown time zone type", this.TimezoneIs)
		return
	}
	time := gotime.Date(this.Year,
		gotime.Month(this.Month),
		int(this.Day),
		int(this.Hour),
		int(this.Minute),
		int(this.Second),
		int(this.Nanosecond),
		location)
	result = &time
	return
}

func (this *Time) validateDate() error {
	if this.Month < monthMin || this.Month > monthMax {
		return fmt.Errorf("%v: Invalid month (must be %v to %v)", this.Month, monthMin, monthMax)
	}
	if this.Day < dayMin || this.Day > dayMax[this.Month] {
		return fmt.Errorf("%v: Invalid day (must be %v to %v)", this.Day, dayMin, dayMax)
	}
	return nil
}

func (this *Time) validateTime() error {
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
	return nil
}

func (this *Time) validateTimezone() error {
	switch this.TimezoneIs {
	case TypeUTC:
		return nil
	case TypeAreaLocation:
		if len(this.AreaLocation) == 0 {
			return fmt.Errorf("Time zone is specified as area/location, but the AreaLocation field is empty")
		}
		return nil
	case TypeLatitudeLongitude:
		if this.LongitudeHundredths < longitudeMin || this.LongitudeHundredths > longitudeMax {
			return fmt.Errorf("%v: Invalid longitude (must be %v to %v)", this.LongitudeHundredths, longitudeMin, longitudeMax)
		}
		if this.LatitudeHundredths < latitudeMin || this.LatitudeHundredths > latitudeMax {
			return fmt.Errorf("%v: Invalid latitude (must be %v to %v)", this.LatitudeHundredths, latitudeMin, latitudeMax)
		}
		return nil
	default:
		return fmt.Errorf("%v: Unknown time zone type", this.TimezoneIs)
	}
}

// Validate this time.
// Only basic validation is performed, enough to ensure that it isn't blatantly
// wrong (such as invalid area/location values, latitude 500, december 54th, etc).
// It does not do more nuanced checks such as on which years February 29th is valid,
// or when leap seconds are allowed. It also doesn't check for impossible timestamp
// values such as 2011-03-13/02:10:00/Los_Angeles.
//
// Note: The field AreaLocation is not validated. You can validate this field
// using time.LoadLocation() or time.LoadLocationFromTZData().
func (this *Time) Validate() (err error) {
	switch this.TimeIs {
	case TypeDate:
		return this.validateDate()
	case TypeTime:
		if err = this.validateTime(); err != nil {
			return
		}
		return this.validateTimezone()
	case TypeTimestamp:
		if err = this.validateDate(); err != nil {
			return
		}
		if err = this.validateTime(); err != nil {
			return
		}
		return this.validateTimezone()
	default:
		return fmt.Errorf("%v: Unknown time type", this.TimeIs)
	}
}

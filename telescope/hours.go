package telescope

import (
	"math"
)

/*
 * Hours, Minutes, Seconds: RA and Hour Angles
 *
 * We store both H:M:S and floating point in our struct so that
 * no matter how we create the struct, the original source is never
 * modified so retrival doesn't cause a conversion and introduce any error
 */

type HMS struct {
	Hours   int
	Minutes int
	Seconds float64
	Float   float64 // alternate representation
}

// New HMS via hours, minutes & seconds
func NewHMS(hours int, minutes int, seconds float64) HMS {
	ret := HMS{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
		Float:   0.0,
	}
	ret.Float = ret.toFloat()
	return ret
}

// New HMS via hours.frac_hours
func NewHMSHours(hours float64) HMS {
	var hrs, min, sec float64
	if hours > 0 {
		hrs = math.Floor(hours)
		fracHours := hours - hrs
		min = math.Floor(fracHours * 60.0)

		sec = fracHours - (float64(min) / 60.0)
		sec *= 3600.0
	} else {
		hrs = math.Ceil(hours)
		fracHours := hours - hrs
		min = math.Abs(math.Ceil(fracHours * 60.0))
		sec = fracHours + (float64(min) / 60.0)
		sec *= -3600.0
	}

	return HMS{
		Hours:   int(hrs),
		Minutes: int(min),
		Seconds: sec,
		Float:   hours,
	}
}

// Sometime we express things in hours min.frac_min
func NewHMSShort(hours int, minutes float64) HMS {
	min := math.Floor(minutes)
	sec := (minutes - min) * 60.0

	hms := HMS{
		Hours:   hours,
		Minutes: int(min),
		Seconds: sec,
		Float:   0.0,
	}
	hms.Float = hms.toFloat()
	return hms
}

// Convert HMS to DMS
func (dms *HMS) ToDMS() DMS {
	degrees := dms.ToDegrees()
	return NewDMSDegrees(degrees)
}

// converts H:M:S to a hours.frac_hours
func (hms *HMS) toFloat() float64 {
	ret := math.Abs(float64(hms.Hours))
	ret += float64(hms.Minutes) / 60.0
	ret += hms.Seconds / 3600.0
	if hms.Hours < 0 {
		ret *= -1.0
	}
	return ret
}

// Convert to 0->360
func (hms *HMS) ToDegrees() float64 {
	return hms.toFloat() * 15.0
}

// Return to hours & minutes.frac_min
func (hms *HMS) HourMinute() (int, float64) {
	min := float64(hms.Minutes) + hms.Seconds/60.0
	return hms.Hours, min
}

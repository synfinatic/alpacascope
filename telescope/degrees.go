package telescope

import (
	"math"
)

/*
 * Degrees, Minutes, Seconds: Lat, Long, Az, Alt and Dec
 *
 * We store both D:M:S and floating point in our struct so that
 * no matter how we create the struct, the original source is never
 * modified so retrival doesn't cause a convertion and introduce any error
 */

type DMS struct {
	Degrees       int
	Minutes       int
	Seconds       float64
	Float         float64 // will be +/- 180
	FloatPositive float64 // will be 0->360
}

func NewDMS(degrees int, minutes int, seconds float64) DMS {
	dms := DMS{
		Degrees:       degrees,
		Minutes:       minutes,
		Seconds:       seconds,
		Float:         0.0,
		FloatPositive: 0.0,
	}
	dms.Float = dms.toFloat()
	if degrees < 0 {
		dms.FloatPositive = dms.Float + 360
	} else {
		dms.FloatPositive = dms.Float
	}
	return dms
}

func NewDMSDegrees(degrees float64) DMS {
	var deg, min, sec float64
	if degrees > 0 {
		deg = math.Floor(degrees)
		min = math.Floor((degrees - deg) * 60.0)
		sec = (degrees - deg - (min / 60.0)) * 3600.0
	} else {
		deg = math.Ceil(degrees)
		min = math.Abs(math.Ceil((degrees - deg) * 60.0))
		sec = math.Abs((degrees - deg + (min / 60.0)) * 3600.0)
	}
	dms := DMS{
		Degrees:       int(deg),
		Minutes:       int(min),
		Seconds:       sec,
		Float:         degrees,
		FloatPositive: degrees,
	}
	if degrees < 0 {
		dms.FloatPositive += 360
	}
	return dms
}

func NewDMSShort(degrees int, minutes float64) DMS {
	min := math.Floor(minutes)
	sec := (minutes - min) * 60.0

	dms := DMS{
		Degrees:       degrees,
		Minutes:       int(min),
		Seconds:       sec,
		Float:         0.0,
		FloatPositive: 0.0,
	}
	dms.Float = dms.toFloat()
	if degrees < 0 {
		dms.FloatPositive = dms.Float + 360
	} else {
		dms.FloatPositive = dms.Float
	}
	return dms
}

func (dms *DMS) HMS() HMS {
	hours := dms.Hours()
	return NewHMSHours(hours)
}

// Returns +/- degrees.  If you need 0->360, add 360.0 to negative results!
func (dms *DMS) toFloat() float64 {
	var ret float64 = math.Abs(float64(dms.Degrees))
	ret += float64(dms.Minutes) / 60.0
	ret += dms.Seconds / 3600.0
	if dms.Degrees < 0 {
		ret *= -1
	}
	return ret
}

// Convert Degrees to hours. +-/12
func (dms *DMS) Hours() float64 {
	return dms.Float / 15.0
}

// Convert Degrees to hours (positive). 0->24
func (dms *DMS) HoursPositive() float64 {
	return dms.FloatPositive / 15.0
}

// Sometimes we want degrees & minutes with seconds as frac_min
func (dms *DMS) DegreeMinute() (int, float64) {
	frac_min := dms.Seconds / 60.0
	return dms.Degrees, float64(dms.Minutes) + frac_min
}

/*
 * GoLang's math library works in radians, so we need functions to convert to degrees
 */

func Rads2degs(rads float64) float64 {
	return rads * 180.0 / math.Pi
}

func Degs2rads(degs float64) float64 {
	return degs * math.Pi / 180.0
}

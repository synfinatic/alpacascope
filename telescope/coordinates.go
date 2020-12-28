package telescope

/*
 * Structs & Functions to manage coordinate conversions
 * Some math stolen from: http://www.stargazing.net/kepler/altaz.html
 */

import (
	"math"
)

/*
 * GoLang's math library works in radians, so we need functions to convert to degrees
 */

func Rads2degs(rads float64) float64 {
	return rads * 180.0 / math.Pi
}

func Degs2rads(degs float64) float64 {
	return degs * math.Pi / 180.0
}

/*
 *  Right Ascension & Declination
 */
type Coordinates struct {
	RA  float64 // Hours.frac_hours
	Dec float64 // Degrees
}

// Returns RA as HMS struct
func (c *Coordinates) RAToHMS() HMS {
	h := int(c.RA)
	remainder := c.RA - float64(h)
	m := int(remainder * 60.0)
	// remove the minutes from the hours and leave fractions of minute
	frac_minute := c.RA - float64(h) - float64(m)/60.0
	s := 60.0 * frac_minute
	return HMS{
		Hours:   h,
		Minutes: m,
		Seconds: s,
	}
}

func (c *Coordinates) DecToDegrees() DMS {
	degrees := int(c.Dec)
	remainder := c.Dec - float64(degrees)
	minutes := int(remainder * 60.0)
	// remove the minutes from the degrees and leave fractions of minute
	frac_minute := c.Dec - float64(degrees) - float64(minutes)/60.0
	seconds := 60.0 * frac_minute
	return DMS{
		Degrees: degrees,
		Minutes: minutes,
		Seconds: seconds,
	}

}

/*
 * Hours, Minutes, Seconds: RA and Hour Angles
 */

type HMS struct {
	Hours   int
	Minutes int
	Seconds float64
}

// Converts hours.frac_hours to HMS
func NewHMS(hours float64) HMS {
	hrs := math.Floor(hours)
	frac_hours := hours - float64(hrs)
	min := math.Floor(frac_hours * 60.0)

	sec := frac_hours - (float64(min) / 60.0)
	sec *= 3600.0
	return HMS{
		Hours:   int(hrs),
		Minutes: int(min),
		Seconds: sec,
	}
}

func (dms *HMS) DMS() DMS {
	degrees := dms.ToDegrees()
	return NewDMS(degrees)
}

// converts H:M:S to a hours.frac_hours
func (hms *HMS) ToFloat() float64 {
	var ret float64 = math.Abs(float64(hms.Hours))
	ret += float64(hms.Minutes) / 60.0
	ret += hms.Seconds / 3600.0
	if hms.Hours < 0 {
		ret *= -1.0
	}
	return ret
}

func (hms *HMS) ToDegrees() float64 {
	return hms.ToFloat() * 15.0
}

// Sometime we express things in hours min.frac_min
func HMToFloat(hours int, minutes float64) float64 {
	min := math.Floor(minutes)
	sec := (minutes - min) * 60.0

	hms := HMS{
		Hours:   hours,
		Minutes: int(min),
		Seconds: sec,
	}
	return hms.ToFloat()
}

func HMSToFloat(hours int, minutes int, seconds float64) float64 {
	hms := HMS{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
	}
	return hms.ToFloat()
}

/*
 * Degrees, Minutes, Seconds: Lat, Long, Az, Alt and Dec
 */
type DMS struct {
	Degrees int
	Minutes int
	Seconds float64
}

func NewDMS(degrees float64) DMS {
	deg := math.Floor(degrees)
	min := math.Floor((degrees - deg) * 60.0)
	sec := (degrees - deg - (min / 60.0)) * 3600.0
	return DMS{
		Degrees: int(deg),
		Minutes: int(min),
		Seconds: sec,
	}
}

func (dms *DMS) HMS() HMS {
	hours := dms.ToHours()
	return NewHMS(hours)
}

// Returns +/- degrees.  If you need 0->360, add 360.0 to negative results!
func (dms *DMS) ToFloat() float64 {
	var ret float64 = math.Abs(float64(dms.Degrees))
	ret += float64(dms.Minutes) / 60.0
	ret += dms.Seconds / 3600.0
	if dms.Degrees < 0 {
		ret *= -1
	}
	return ret
}

// Convert Degrees to hours
func (dms *DMS) ToHours() float64 {
	return dms.ToFloat() / 15.0
}

// Sometimes we want degrees & minutes with seconds as frac_min
func (dms *DMS) ToDM() (int, float64) {
	frac_min := dms.Seconds / 60.0
	return dms.Degrees, float64(dms.Minutes) + frac_min
}

/*
 * Altitude & Azimuth
 */
type AltAz struct {
	Alt float64
	Az  float64
}

// returns degrees of Alt.  + North, - South
func GetAlt(hourangle float64, dec float64, latitude float64) float64 {
	rha := Degs2rads(hourangle)
	rdec := Degs2rads(dec)
	rlat := Degs2rads(latitude)

	alt := math.Sin(rdec) * math.Sin(rlat)
	alt += math.Cos(rdec) * math.Cos(rlat) * math.Cos(rha)
	return Rads2degs(math.Asin(alt))
}

// returns degrees Az.  + East, - West
func GetAz(hourangle float64, dec float64, latitude float64) float64 {
	ralt := Degs2rads(GetAlt(hourangle, dec, latitude))
	rdec := Degs2rads(dec)
	rlat := Degs2rads(latitude)
	rha := Degs2rads(hourangle)

	az := math.Sin(rdec) - (math.Sin(ralt) * math.Sin(rlat))
	az /= math.Cos(ralt) * math.Cos(rlat)
	if rha < 0 {
		return Rads2degs(math.Acos(az))
	} else {
		return 360.0 - Rads2degs(math.Acos(az))
	}
}

// Convert RA + LocalSiderialTime to Hour Angle (RA/LST should be in degrees)
func RAToHourAngle(ra HMS, lst float64) float64 {
	var ha = lst - ra.ToDegrees()
	if ha < 0 {
		ha += 360.0
	}
	return ha
}

/*
// Returns the Alt/Az (in degrees)
// for an object given the RA/Dec & observer time, lat (deg), long (hrs)
func GetAltAz(ra HMS, dec HMS, latitude DMS, longitude DMS, local_time time.Time) AltAz {
	gmst := GreenwichMeanStandardTime(local_time)
	lst := GMSTToLST(gmst, longitude.ToHours())
	hourangle := RAToHourAngle(ra, lst)
	ha := NewDMS(hourangle)
	fmt.Printf("hourangle = %v\n", ha.ToHours())
	altaz := AltAz{
		Alt: GetAlt(hourangle, dec.ToFloat(), latitude.ToFloat()),
		Az:  GetAz(hourangle, dec.ToFloat(), latitude.ToFloat()),
	}
	return altaz
}
*/

package telescope

/*
 * Structs & Functions to manage coordinate conversions
 * Some math stolen from: http://www.stargazing.net/kepler/altaz.html
 */

import (
	"math"
)

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

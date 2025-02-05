package telescope

/*
 * Note that all of these functions are NOT used by AlpacaScope!
 * Formulas from https://thecynster.home.blog/2019/11/04/calculating-sidereal-time/
 */
import (
	"math"
	"time"
)

// Returns days.frac_day since J2000 for UTC time
func TimeToJ2000(t time.Time) float64 {
	return TimeToJYear(2000, t)
}

// Returns days.frac_day since given JXXXX day for UTC time
func TimeToJYear(jyear int, t time.Time) float64 {
	var utc time.Time = t.UTC()
	year, _, _ := utc.Date()
	hour, minute, second := utc.Clock()

	fracDay := float64(hour)
	fracDay += float64(minute) / 60.0
	fracDay += float64(second) / 3600.0
	fracDay /= 24.0

	yearDays := -1.5

	utctz, _ := time.LoadLocation("UTC")
	if year < jyear {
		for i := jyear - 1; i > year; i-- {
			y := time.Date(i, 12, 31, 12, 0, 0, 0, utctz)
			yearDays -= float64(y.YearDay())
		}

		y := time.Date(year, 12, 31, 12, 0, 0, 0, utctz)
		yearDays -= float64(y.YearDay()) - float64(utc.YearDay())
	} else if year > jyear {
		for i := jyear; i < year; i++ {
			y := time.Date(i, 12, 31, 12, 0, 0, 0, utctz)
			yearDays += float64(y.YearDay())
		}
		yearDays += float64(utc.YearDay()) - 1
	} else {
		// year is the jyear
		yearDays += float64(utc.YearDay())
	}

	return fracDay + yearDays
}

// Convert Greenwich Mean Siderial Time to Local Siderial Time (hours)
func GMSTToLST(gmst float64, longitudeHrs float64) float64 {
	lst := gmst + longitudeHrs // must be in hrs, not degrees!
	return math.Mod(lst, 24.0)
}

// Returns hour.hour_frac
func GreenwichMeanSiderealTime(t time.Time) float64 {
	UTChrs := TimeToUTCHours(t)

	// Julian Date
	JDutc := CalcJulianDate(t, UTChrs)

	// Corrected UTC Time
	TIMEtt := UTChrs + (LeapSeconds(t)+32.184)/3600.0

	// Corrected Juilan Date
	JDtt := CalcJulianDate(t, TIMEtt)

	// Du = days since epoch J2000
	Du := JDutc - 2451545.0

	// Calc Earth Rotational Angle for Du
	ThetaDu := 2.0 * math.Pi * (0.7790572732640 + 1.00273781191135448*Du)
	ThetaDu = math.Mod(ThetaDu, 2.0*math.Pi)
	ThetaDu = ThetaDu * 180.0 / math.Pi / 15.0 // convert radians to hours

	// number of centries since epoch J2000
	T := (JDtt - 2451545.0) / 36525.0

	// Greenwich Mean Standard Time polynominal part for earth's obliquity (total hours)
	GMSTpT := 0.014506 + 4612.156534*T + 1.3915817*math.Pow(T, 2)
	GMSTpT -= 0.00000044*math.Pow(T, 3) + 0.000029956*math.Pow(T, 4) + 0.0000000368*math.Pow(T, 5)

	GMSTpT = math.Mod(GMSTpT/3600.0, 360.0) / 15.0 // convert arc-seconds to hours
	if GMSTpT < 0.0 {
		GMSTpT += 360.0
	}

	return math.Mod(ThetaDu+GMSTpT, 24.0) // return in hours
}

// Convert to time in any TZ to the current hour in UTC
func TimeToUTCHours(t time.Time) float64 {
	// can't just t.UTC(), because hours can be > 24.0 if in "next date UTC" (ie USA late at night)
	_, offsetSec := t.Zone()

	hours := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0
	hours -= float64(offsetSec) / 3600.0
	return hours
}

// returns number of leap seconds to date UTC per https://en.m.wikipedia.org/wiki/Leap_second
func LeapSeconds(t time.Time) float64 {
	leapSeconds := 0.0
	utc := t.UTC()
	leapSecTimes := [][]int{
		// year, month, day.  Only Jun 30 & Dec 31st are valid
		{1972, 6, 30},
		{1972, 12, 31},
		{1973, 12, 31},
		{1974, 12, 31},
		{1975, 12, 31},
		{1976, 12, 31},
		{1977, 12, 31},
		{1978, 12, 31},
		{1979, 12, 31},
		{1981, 6, 30},
		{1982, 6, 30},
		{1983, 6, 30},
		{1985, 6, 30},
		{1987, 12, 31},
		{1989, 12, 31},
		{1990, 12, 31},
		{1992, 6, 30},
		{1993, 6, 30},
		{1994, 6, 30},
		{1995, 12, 31},
		{1997, 6, 30},
		{1998, 12, 31},
		{2005, 12, 31},
		{2008, 12, 31},
		{2012, 6, 30},
		{2015, 6, 30},
		{2016, 12, 31},
		// nothing else as of 2020-12-27.  New leap seconds are announced ~6mo in advance
	}
	utctz, _ := time.LoadLocation("UTC")
	for _, ls := range leapSecTimes {
		// hours, min, sec are always the end of the day
		leapSec := time.Date(ls[0], time.Month(ls[1]), ls[2], 23, 59, 59, 999999999, utctz)
		if utc.After(leapSec) {
			leapSeconds += 1.0
		}
	}
	return leapSeconds
}

// Works for UTC and Corrected UTC times
func CalcJulianDate(t time.Time, utcHours float64) float64 {
	year := float64(t.Year())
	month := float64(t.Month())
	day := float64(t.Day())

	JD := 367.0*year - math.Floor(7.0*(year+math.Floor((month+9.0)/12.0))/4.0)
	JD -= math.Floor(3.0 * (math.Floor((year+(month-9.0)/7.0)/100.0) + 1.0) / 4.0)
	JD += math.Floor(275.0 * month / 9.0)
	JD += day + 1721028.5 + utcHours/24.0
	return JD
}

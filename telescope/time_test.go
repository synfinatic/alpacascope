package telescope

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeToJ2000(t *testing.T) {
	utctz, _ := time.LoadLocation("UTC")
	tests := map[float64]time.Time{
		-508.53472222222223: time.Date(1998, 8, 10, 23, 10, 0, 0, utctz),
		-0.5:                time.Date(2000, 1, 1, 0, 0, 0, 0, utctz),
		364.5:               time.Date(2001, 1, 1, 0, 0, 0, 0, utctz),
		3651.5:              time.Date(2010, 1, 1, 0, 0, 0, 0, utctz),
		3015.1458333333335:  time.Date(2008, 4, 4, 15, 30, 0, 0, utctz),
	}
	for j2000, test := range tests {
		assert.Equal(t, j2000, TimeToJ2000(test))
	}
}

func TestTimeToUTCHours(t *testing.T) {
	est, _ := time.LoadLocation("EST")
	pst, _ := time.LoadLocation("America/Los_Angeles")
	utc, _ := time.LoadLocation("UTC")
	tests := map[float64]time.Time{
		25.083333333333333: time.Date(1969, 1, 5, 20, 5, 0, 0, est),
		27.0:               time.Date(2020, 12, 26, 19, 0, 0, 0, pst),
		23.166666666666668: time.Date(1998, 8, 10, 23, 10, 0, 0, utc),
	}

	for check, test := range tests {
		assert.Equal(t, check, TimeToUTCHours(test))
	}
}

func TestSideralTimeCalcs(t *testing.T) {
	// Using Marlin Manson example from:
	// https://thecynster.home.blog/2019/11/04/calculating-sidereal-time/
	tz, _ := time.LoadLocation("EST")
	mm := time.Date(1969, 1, 5, 20, 5, 0, 0, tz)
	hours := TimeToUTCHours(mm)
	assert.Equal(t, 2440227.54513888889, CalcJulianDate(mm, hours))
	tt_hours := hours + (LeapSeconds(mm)+32.184)/3600.0
	assert.Equal(t, 2.440227545511389e+06, CalcJulianDate(mm, tt_hours))
	gmst := GreenwichMeanSiderealTime(mm)
	assert.Equal(t, 8.112740425894629, gmst)
	dms := DMS{
		-81,
		23,
		0,
	}
	long_hrs := dms.ToFloat() / 15.0
	assert.Equal(t, -5.425555555555556, long_hrs)
	assert.Equal(t, 2.687184870339072, GMSTToLST(gmst, long_hrs))

	/*
	 * second test from SkySafari
	 */
	pacific, _ := time.LoadLocation("America/Los_Angeles") // GMT-8
	// Sat, Dec 26 2020 @ 7pm
	viewing := time.Date(2020, 12, 26, 19, 0, 0, 0, pacific)

	// Julian Date per SkySafari
	assert.Equal(t, 2459210.625000, CalcJulianDate(viewing, TimeToUTCHours(viewing)))

	// San Jose, CA
	gmst = GreenwichMeanSiderealTime(viewing)
	dms = DMS{
		-121,
		53,
		41.9,
	}
	long_hrs = dms.ToHours()

	// LST per SkySafari is 1h 16m 41sec but doesn't have enough resolution
	assert.Equal(t, 1.2779162228025633, GMSTToLST(gmst, long_hrs))

	/*
	 * Third test using old values from http://www.stargazing.net/kepler/altaz.html
	 */
	utctz, _ := time.LoadLocation("UTC")
	m13_time := time.Date(1998, 8, 10, 22, 10, 0, 0, utctz)

	// Julian date per SkySafari
	assert.Equal(t, 2.451036423611111e+06, CalcJulianDate(m13_time, TimeToUTCHours(m13_time)))

	gmst = GreenwichMeanSiderealTime(m13_time)
	dms = DMS{
		-1,
		53,
		59.4,
	}
	long_hrs = dms.ToHours()

	// LST per SkySafari is 19h 19m 9s but doesn't have enough resolution
	assert.Equal(t, 19.318920795558416, GMSTToLST(gmst, long_hrs))

}

package telescope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMS(t *testing.T) {
	tests := map[float64]HMS{
		0.0:    HMS{Hours: 0, Minutes: 0, Seconds: 0.0},
		8.2525: HMS{Hours: 8, Minutes: 15, Seconds: 8.99999999999821},    // float error
		12.505: HMS{Hours: 12, Minutes: 30, Seconds: 18.000000000002814}, // float error
	}
	for test, check := range tests {
		assert.Equal(t, check, NewHMS(test))
	}
}

func TestHMSToFloat(t *testing.T) {
	tests := map[float64]HMS{
		-1.9166666666666665: HMS{-1, 55, 0},
		22.841666666666665:  HMS{22, 50, 30},
	}
	for check, test := range tests {
		assert.Equal(t, check, test.ToFloat())
	}
}

func TestHM(t *testing.T) {
	tests := map[float64]float64{
		0.0:                HMToFloat(0, 0.0),
		8.508333333333333:  HMToFloat(8, 30.5),   // 8:30:30
		12.254166666666666: HMToFloat(12, 15.25), // 12:15:15
	}
	for check, test := range tests {
		assert.Equal(t, check, test)
	}
}

func TestRAToHourAngle(t *testing.T) {
	tests := map[float64]map[string]float64{
		54.382619999999974: map[string]float64{
			"ra":  250.425,
			"lst": 304.80762,
		},
	}
	for check, test := range tests {
		ra := NewDMS(test["ra"])
		assert.Equal(t, check, RAToHourAngle(ra.HMS(), test["lst"]))
	}
}

func TestGetAlt(t *testing.T) {
	tests := map[float64]map[string]float64{
		49.169127488469556: map[string]float64{
			"ha":  54.382617,
			"dec": 36.466667,
			"lat": 52.5,
		},
	}
	for check, test := range tests {
		assert.Equal(t, check, GetAlt(test["ha"], test["dec"], test["lat"]))
	}
}

func TestGetAz(t *testing.T) {
	tests := map[float64]map[string]float64{
		269.1463277297406: map[string]float64{
			"ha":  54.382617,
			"dec": 36.466667,
			"lat": 52.5,
		},
	}
	for check, test := range tests {
		assert.Equal(t, check, GetAz(test["ha"], test["dec"], test["lat"]))
	}
}

/*
// Full test using SkySafari as a check
func TestGetAltAz(t *testing.T) {
	// Betelgeuse J2000
	ra := HMS{Hours: 5, Minutes: 55, Seconds: 10.31}
	dec := HMS{Hours: 7, Minutes: 24, Seconds: 25.4}

	// San Jose, CA
	lat := DMS{Degrees: 37, Minutes: 20, Seconds: 21.8}
	long := DMS{Degrees: -121, Minutes: 53, Seconds: 41.9}

	tz, _ := time.LoadLocation("America/Los_Angeles")
	local_time := time.Date(2020, 12, 26, 19, 0, 0, 0, tz)
	altaz := GetAltAz(ra, dec, lat, long, local_time)
	//	Alt/Az per SkySafari Pro 6.7.2
	alt := DMS{20, 28, 27.5}
	az := DMS{96, 22, 39.2}
	assert.Equal(t, alt.ToFloat(), altaz.Alt)
	assert.Equal(t, az.ToFloat(), altaz.Az)
}
*/

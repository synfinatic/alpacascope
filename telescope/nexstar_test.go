package telescope

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexstarRA32(t *testing.T) {
	one_step := 5.587935447692871e-09
	one_hour := 0.9999999962747097    // rounding error
	two_hours := 1.9999999981373549   // rouding error
	eight_hours := 7.999999998137355  // rounding error
	thirty_min := 0.49999999813735485 // rounding error
	thirty_min_steps := uint32(math.Pow(2, 32) / 24.0 / 2.0)
	tests := map[uint32]float64{
		0:                                0.0,
		1:                                one_step,
		thirty_min_steps:                 thirty_min,
		uint32(math.Pow(2, 32)):          0.0,
		uint32(math.Pow(2, 32) / 2.0):    12.0,
		uint32(math.Pow(2, 32) / 3.0):    eight_hours,
		uint32(math.Pow(2, 32)/2.0) + 1:  12.0 + one_step,
		uint32(math.Pow(2, 32)/2.0) - 1:  12.0 - one_step,
		uint32(math.Pow(2, 32) / 12.0):   two_hours,
		uint32(math.Pow(2, 32) / 24.0):   one_hour,
		uint32(math.Pow(2, 32)/24.0) + 1: one_hour + one_step,
		uint32(math.Pow(2, 32)/24.0) - 1: one_hour - one_step,
		uint32(math.Pow(2, 32)/24.0) + thirty_min_steps: one_hour + thirty_min,
		uint32(math.Pow(2, 32)/24.0) - thirty_min_steps: one_hour - thirty_min,
	}

	for k, v := range tests {
		ra := uint32StepsToRA(k)
		assert.Equal(t, v, ra)
		steps := raTo32bitSteps(ra)
		assert.Equal(t, k, steps)
	}
}

func TestNexstarRA16(t *testing.T) {
	one_step := 0.0003662109375
	one_hour := 0.999755859375     // rounding error
	two_hours := 1.9998779296875   // rouding error
	eight_hours := 7.9998779296875 // rounding error
	thirty_min := 0.4998779296875  // rounding error
	thirty_min_steps := uint16(math.Pow(2, 16) / 24.0 / 2.0)
	tests := map[uint16]float64{
		0:                                0.0,
		1:                                one_step,
		thirty_min_steps:                 thirty_min,
		uint16(math.Pow(2, 16)):          0.0,
		uint16(math.Pow(2, 16) / 2.0):    12.0,
		uint16(math.Pow(2, 16) / 3.0):    eight_hours,
		uint16(math.Pow(2, 16)/2.0) + 1:  12.0 + one_step,
		uint16(math.Pow(2, 16)/2.0) - 1:  12.0 - one_step,
		uint16(math.Pow(2, 16) / 12.0):   two_hours,
		uint16(math.Pow(2, 16) / 24.0):   one_hour,
		uint16(math.Pow(2, 16)/24.0) + 1: one_hour + one_step,
		uint16(math.Pow(2, 16)/24.0) - 1: one_hour - one_step,
		uint16(math.Pow(2, 16)/24.0) + thirty_min_steps: one_hour + thirty_min,
		uint16(math.Pow(2, 16)/24.0) - thirty_min_steps: one_hour - thirty_min,
	}
	for k, _ := range tests {
		ra := uint16StepsToRA(k)
		steps := raTo16bitSteps(ra)
		assert.Equal(t, k, steps)
	}

}

func TestNexstarDec32(t *testing.T) {
	tests := map[float64]float64{
		0.0:   0.0,
		12.5:  12.499999925494194,
		45.0:  45.0,
		57.5:  57.499999925494194,
		90.0:  90.0,
		-90.0: -90.0,
		-57.5: -57.499999925494194,
		-45.0: -45.0,
		-12.5: -12.499999925494194,
	}
	for k, v := range tests {
		bytes := uint32(math.Pow(2, 32) / 360.0 * k)
		result := uint32StepsToDec(bytes)
		assert.Equal(t, v, result)
		result2 := decTo32bitSteps(result)
		assert.Equal(t, bytes, result2)
	}
}

func TestNexstarDec16(t *testing.T) {
	tests := map[float64]float64{
		0.0:   0.0,
		12.5:  12.4969482421875,
		45.0:  45.0,
		57.5:  57.4969482421875,
		90.0:  90.0,
		-90.0: -90.0,
		-57.5: -57.4969482421875,
		-45.0: -45.0,
		-12.5: -12.4969482421875,
	}
	for k, v := range tests {
		bytes := uint16(math.Pow(2, 16) / 360.0 * k)
		result := uint16StepsToDec(bytes)
		assert.Equal(t, v, result)
		result2 := decTo16bitSteps(result)
		assert.Equal(t, bytes, result2)
	}
}

func TestNestarToHMS(t *testing.T) {
	tests := map[uint32]HMS{
		0:                                HMS{0, 0, 0.0},
		1:                                HMS{0, 0, 3.3527612686157227e-07},
		uint32(math.Pow(2, 32) / 2.0):    HMS{12.0, 0, 0.0},
		uint32(math.Pow(2, 32)/2.0 + 1):  HMS{12.0, 0, 3.3527612686157227e-07},
		uint32(math.Pow(2, 32)/2.0 + 2):  HMS{12.0, 0, 6.705522537231445e-07},
		uint32(math.Pow(2, 32) / 24.0):   HMS{0.0, 59, 0.9999997764825852},
		uint32(math.Pow(2, 32)/24.0 + 1): HMS{1.0, 0, 1.1175870895385742e-07},
	}
	for input, check := range tests {
		ra_value := uint32StepsToRA(input)
		c := Coordinates{
			RA:  ra_value,
			Dec: 0.0,
		}
		hms := c.RAToHMS()
		assert.Equal(t, check.Hours, hms.Hours)
		assert.Equal(t, check.Minutes, hms.Minutes)
		assert.Equal(t, check.Seconds, hms.Seconds)
	}
}

func TestNexstarRABytes(t *testing.T) {
	tests := map[float64][]byte{
		0.0:                []byte{'0', '0', '0', '0', '0', '0', '0', '0'},
		6.0:                []byte{'4', '0', '0', '0', '0', '0', '0', '0'},
		6.24999999627471:   []byte{'4', '2', 'A', 'A', 'A', 'A', 'A', 'A'}, // rounding error
		11.999999994412065: []byte{'7', 'F', 'F', 'F', 'F', 'F', 'F', 'F'}, // rounding error
		23.999999994412065: []byte{'F', 'F', 'F', 'F', 'F', 'F', 'F', 'F'}, // rounding error
	}
	for ra, bytes := range tests {
		c := NewCoordinateNexstar(bytes, tests[0.0], true)
		assert.Equal(t, ra, c.RA)
	}
}

func TestNexstarDecBytes(t *testing.T) {
	tests := map[float64][]byte{
		0.0:                     []byte{'0', '0', '0', '0', '0', '0', '0', '0'},
		45.0:                    []byte{'2', '0', '0', '0', '0', '0', '0', '0'},
		90.0:                    []byte{'4', '0', '0', '0', '0', '0', '0', '0'},
		93.74999994412065:       []byte{'4', '2', 'A', 'A', 'A', 'A', 'A', 'A'}, // rounding error
		179.99999991618097:      []byte{'7', 'F', 'F', 'F', 'F', 'F', 'F', 'F'}, // rounding error
		180.0:                   []byte{'8', '0', '0', '0', '0', '0', '0', '0'},
		-8.381903171539307e-08:  []byte{'F', 'F', 'F', 'F', 'F', 'F', 'F', 'F'}, // rounding error
		-1.6763806343078613e-07: []byte{'F', 'F', 'F', 'F', 'F', 'F', 'F', 'E'}, // rounding error
		-45.0:                   []byte{'E', '0', '0', '0', '0', '0', '0', '0'},
		-179.99999991618097:     []byte{'8', '0', '0', '0', '0', '0', '0', '1'}, // rounding error
	}
	for dec, bytes := range tests {
		c := NewCoordinateNexstar(tests[0.0], bytes, true)
		assert.Equal(t, dec, c.Dec)
	}
}

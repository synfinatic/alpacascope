package telescope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDMS(t *testing.T) {
	tests := map[float64]DMS{
		0.0:                 NewDMS(0, 0, 0.0),
		8.2525:              NewDMS(8, 15, 8.99999999999821),    // float error
		12.505:              NewDMS(12, 30, 18.000000000002814), // float error
		-1.9166666666666665: NewDMS(-1, 55, 0),
		22.841666666666665:  NewDMS(22, 50, 30),
	}
	for check, test := range tests {
		assert.Equal(t, check, test.Float)
		assert.Equal(t, check/15.0, test.Hours())
	}
}

func TestNewDMSShort(t *testing.T) {
	tests := map[float64]DMS{
		0.0:                 NewDMSShort(0, 0.0),
		8.508333333333333:   NewDMSShort(8, 30.5),   // 8:30:30
		12.254166666666666:  NewDMSShort(12, 15.25), // 12:15:15
		-1.9166666666666665: NewDMSShort(-1, 55),
	}
	for check, test := range tests {
		assert.Equal(t, check, test.Float)
	}
}

func TestNewDMSHours(t *testing.T) {
	tests := map[float64]DMS{
		0.0:                 NewDMS(0, 0, 0.0),
		8.2525:              NewDMS(8, 15, 8.99999999999821),    // float error
		12.505:              NewDMS(12, 30, 18.000000000002814), // float error
		-1.9166666666666665: NewDMS(-1, 54, 59.99999999999939),  // float error
		22.841666666666665:  NewDMS(22, 50, 29.999999999993896),
	}
	for test, check := range tests {
		assert.Equal(t, check, NewDMSDegrees(test))
	}
}

func TestNewDMSHourMinute(t *testing.T) {
	hms := NewDMSShort(0, 0.0)
	hr, min := hms.DegreeMinute()
	assert.Equal(t, 0, hr)
	assert.Equal(t, 0.0, min)

	hms = NewDMSShort(8, 30.5)
	hr, min = hms.DegreeMinute()
	assert.Equal(t, 8, hr)
	assert.Equal(t, 30.5, min)

	hms = NewDMSShort(12, 15.25)
	hr, min = hms.DegreeMinute()
	assert.Equal(t, 12, hr)
	assert.Equal(t, 15.25, min)

	hms = NewDMSShort(-1, 55)
	hr, min = hms.DegreeMinute()
	assert.Equal(t, -1, hr)
	assert.Equal(t, 55.0, min)
}

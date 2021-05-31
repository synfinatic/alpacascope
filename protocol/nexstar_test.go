package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type LatLong struct {
	Bytes []byte
	Lat   float64
	Long  float64
}

func TestLatLong(t *testing.T) {
	tests := []LatLong{
		LatLong{
			Bytes: []byte{118, 20, 17, 0, 33, 50, 41, 1},
			Lat:   118.33805555555556,
			Long:  -33.844722222222224,
		},
		LatLong{
			Bytes: []byte{118, 20, 17, 1, 33, 50, 41, 0},
			Lat:   -118.33805555555556,
			Long:  33.844722222222224,
		},
	}

	for _, test := range tests {
		lat, long := NexstarToLatLong(test.Bytes)
		assert.Equal(t, test.Lat, lat)
		assert.Equal(t, test.Long, long)
		cbytes := LatLongToNexstar(test.Lat, test.Long)
		assert.Equal(t, test.Bytes, cbytes)
	}
}

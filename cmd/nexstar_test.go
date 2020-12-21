package main

import (
	// 	"encoding/binary"
	"math"
	"testing"
)

// Test 32bit -> RA -> 32bit via uint32StepsToRA & raTo32bitSteps
func Test32BitRALoop(t *testing.T) {
	one_step := 5.587935447692871e-09
	one_hour := 0.9999999962747097    // rounding error
	two_hours := 1.9999999981373549   // rouding error
	eight_hours := 7.999999998137355  // rounding error
	thirty_min := 0.49999999813735485 // rounding error
	thirty_min_steps := uint32(math.Pow(2, 32) / 24.0 / 2.0)
	my_tests := map[uint32]float64{
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
	for k, v := range my_tests {
		raLoop32bit(t, k, v)
	}
}

// Test 16bit -> RA -> 16bit via uint16StepsToRA & raTo16bitSteps
func Test16BitRALoop(t *testing.T) {
	one_step := 0.0003662109375
	one_hour := 0.999755859375     // rounding error
	two_hours := 1.9998779296875   // rouding error
	eight_hours := 7.9998779296875 // rounding error
	thirty_min := 0.4998779296875  // rounding error
	thirty_min_steps := uint16(math.Pow(2, 16) / 24.0 / 2.0)
	my_tests := map[uint16]float64{
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
	for k, v := range my_tests {
		raLoop16bit(t, k, v)
	}

}

// Test 32bit -> Dec -> 32bit via uint32StepsToDec & decTo32bitSteps
func Test32BitDecLoop(t *testing.T) {
	my_tests := map[float64]float64{
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
	for bytes, result := range my_tests {
		val := uint32(math.Pow(2, 32) / 360.0 * bytes)
		decLoop32bit(t, val, result)
	}
}

// Test 16bit -> Dec -> 16bit via uint16StepsToDec & decTo16bitSteps
func Test16BitDecLoop(t *testing.T) {
	my_tests := map[float64]float64{
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
	for _, test := range my_tests {
		val := uint16(math.Pow(2, 16) / 360.0 * test)
		decLoop16bit(t, val, test)
	}

}

// Tests rev32ToHMS()
func Test32BitHMS(t *testing.T) {
	my_tests := map[uint32]HMS{
		0:                                HMS{0, 0, 0.0},
		1:                                HMS{0, 0, 3.3527612686157227e-07},
		uint32(math.Pow(2, 32) / 2.0):    HMS{12.0, 0, 0.0},
		uint32(math.Pow(2, 32)/2.0 + 1):  HMS{12.0, 0, 3.3527612686157227e-07},
		uint32(math.Pow(2, 32)/2.0 + 2):  HMS{12.0, 0, 6.705522537231445e-07},
		uint32(math.Pow(2, 32) / 24.0):   HMS{0.0, 59, 0.9999997764825852},
		uint32(math.Pow(2, 32)/24.0 + 1): HMS{1.0, 0, 1.1175870895385742e-07},
	}
	for input, check := range my_tests {
		hms := rev32ToHMS(input)
		if check.Hours != hms.Hours {
			t.Errorf("rev32ToHMS: %v failed, expected hours %v, got %v", input, check.Hours, hms.Hours)
		}
		if check.Minutes != hms.Minutes {
			t.Errorf("rev32ToHMS: %v failed, expected minutes %v, got %v", input, check.Minutes, hms.Minutes)
		}
		if check.Seconds != hms.Seconds {
			t.Errorf("rev32ToHMS: %v failed, expected seconds %v, got %v", input, check.Seconds, hms.Seconds)
		}
	}
}

// Tests rev16ToHMS()
func Test16BitHMS(t *testing.T) {
	my_tests := map[uint16]HMS{
		0:                                HMS{0, 0, 0.0},
		1:                                HMS{0, 0, 0.02197265625},
		uint16(math.Pow(2, 16) / 2.0):    HMS{12.0, 0, 0.0},
		uint16(math.Pow(2, 16)/2.0 + 1):  HMS{12.0, 0, 0.02197265625},
		uint16(math.Pow(2, 16)/2.0 + 2):  HMS{12.0, 0, 0.0439453125},
		uint16(math.Pow(2, 16) / 24.0):   HMS{0.0, 59, 0.9853515625000031},
		uint16(math.Pow(2, 16)/24.0 + 1): HMS{1.0, 0, 0.00732421875},
	}
	for input, check := range my_tests {
		hms := rev16ToHMS(input)
		if check.Hours != hms.Hours {
			t.Errorf("rev16ToHMS: %v failed, expected hours %v, got %v", input, check.Hours, hms.Hours)
		}
		if check.Minutes != hms.Minutes {
			t.Errorf("rev16ToHMS: %v failed, expected minutes %v, got %v", input, check.Minutes, hms.Minutes)
		}
		if check.Seconds != hms.Seconds {
			t.Errorf("rev16ToHMS: %v failed, expected seconds %v, got %v", input, check.Seconds, hms.Seconds)
		}
	}

}

// Tests bytes_to_latlong()
func TestBytesToLatLong(t *testing.T) {

}

/*
 * Helper functions for the tests defined above
 */

func raLoop32bit(t *testing.T, bytes uint32, ra float64) {
	result := uint32StepsToRA(bytes)
	if result != ra {
		t.Errorf("raLoop32bit bytes %v -> float failed, expected %v, got %v", bytes, ra, result)
	}
	result2 := raTo32bitSteps(result)
	if bytes != result2 {
		t.Errorf("raLoop32bit float %v -> bytes failed, expected %v, got %v", result, bytes, result2)
	}
}

func raLoop16bit(t *testing.T, bytes uint16, ra float64) {
	result := uint16StepsToRA(bytes)
	if result != ra {
		t.Errorf("raLoop16bit bytes %v -> float failed, expected %v, got %v", bytes, ra, result)
	}
	result2 := raTo16bitSteps(result)
	if bytes != result2 {
		t.Errorf("raLoop16bit float %v -> bytes failed, expected %v, got %v", result, bytes, result2)
	}
}

func decLoop32bit(t *testing.T, bytes uint32, dec float64) {
	result := uint32StepsToDec(bytes)
	if result != dec {
		t.Errorf("decLoop32bit bytes %v -> float failed, expected %v, got %v", bytes, dec, result)
	}
	result2 := decTo32bitSteps(result)
	if bytes != result2 {
		t.Errorf("decLoop32bit float %v -> bytes failed, expected %v, got %v", result, bytes, result2)
	}
}

func decLoop16bit(t *testing.T, bytes uint16, dec float64) {
	result := uint16StepsToDec(bytes)
	if result != dec {
		t.Errorf("decLoop16bit bytes %v -> float failed, expected %v, got %v", bytes, dec, result)
	}
	result2 := decTo16bitSteps(result)
	if bytes != result2 {
		t.Errorf("decLoop16bit float %v -> bytes failed, expected %v, got %v", result, bytes, result2)
	}
}

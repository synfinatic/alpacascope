package main

import (
	// 	"encoding/binary"
	"math"
	"testing"
)

func Test16BitToRA(t *testing.T) {
	test := uint16(math.Pow(2, 16) / 2.0)
	result := uint16StepsToRA(test)
	expected := 12.0
	if result != expected {
		t.Errorf("Test16BitToresult(%v) failed, expected %v, got %v", test, expected, result)
	}

	one_hour := uint16(math.Pow(2, 16) / 24.0)
	test = uint16(math.Pow(2, 16)/2.0) + one_hour
	result = uint16StepsToRA(test)
	expected = 12.999755859375 // rouding errors :(
	if result != expected {
		t.Errorf("Test16BitToresult(%v) failed, expected %v, got %v", test, expected, result)
	}

	thirty_min := uint16(math.Pow(2, 16) / 24.0 / 2.0)
	test = uint16(math.Pow(2, 16)/2.0) + thirty_min
	result = uint16StepsToRA(test)
	expected = 12.4998779296875 // rouding errors :(
	if result != expected {
		t.Errorf("Test16BitToresult(%v) failed, expected %v, got %v", test, expected, result)
	}
}

func Test32BitToRA(t *testing.T) {
	test := uint32(math.Pow(2, 32) / 2.0)
	result := uint32StepsToRA(test)
	expected := 12.0
	if result != expected {
		t.Errorf("Test32BitToRA(%v) failed, expected %v, got %v", test, expected, result)
	}

	one_hour := uint32(math.Pow(2, 32) / 24.0)
	test = uint32(math.Pow(2, 32)/2.0) + one_hour
	result = uint32StepsToRA(test)
	expected = 12.99999999627471 // rouding errors :(
	if result != expected {
		t.Errorf("Test32BitToresult(%v) failed, expected %v, got %v", test, expected, result)
	}

	thirty_min := uint32(math.Pow(2, 32) / 24.0 / 2.0)
	test = uint32(math.Pow(2, 32)/2.0) + thirty_min
	result = uint32StepsToRA(test)
	expected = 12.499999998137355 // rouding errors :(
	if result != expected {
		t.Errorf("Test32BitToresult(%v) failed, expected %v, got %v", test, expected, result)
	}
}

func Test16BitToDec(t *testing.T) {
	test := uint16(math.Pow(2, 16) / 360.0 * 45.0)
	result := uint16StepsToDec(test)
	expected := 45.0
	if result != expected {
		t.Errorf("Test16BitToDec(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = uint16(math.Pow(2, 16) / 360.0 * (270.0 + 45.0))
	result = uint16StepsToDec(test)
	expected = -45.0
	if result != expected {
		t.Errorf("Test16BitToDec(%v) failed, expected %v, got %v", test, expected, result)
	}
}

func Test32BitToDec(t *testing.T) {
	test := uint32(math.Pow(2, 32) / 360.0 * 45.0)
	result := uint32StepsToDec(test)
	expected := 45.0
	if result != expected {
		t.Errorf("Test32BitToDec(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = uint32(math.Pow(2, 32) / 360.0 * (270.0 + 45.0))
	result = uint32StepsToDec(test)
	expected = -45.0
	if result != expected {
		t.Errorf("Test32BitToDec(%v) failed, expected %v, got %v", test, expected, result)
	}

}

func TestRATo16Bit(t *testing.T) {
	test := 12.0
	result := raTo16bitSteps(test)
	expected := uint16(math.Pow(2, 16) / 2.0)
	if result != expected {
		t.Errorf("TestRATo16Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = 12.5
	result = raTo16bitSteps(test)
	thirty_min := uint16(math.Pow(2, 16) / 24.0 / 2.0)
	expected = uint16(math.Pow(2, 16)/2.0) + thirty_min
	if result != expected {
		t.Errorf("TestRATo16Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = 13.0
	result = raTo16bitSteps(test)
	expected = uint16(math.Pow(2, 16)/2.0) + thirty_min + thirty_min
	if result != expected {
		t.Errorf("TestRATo16Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

}

func TestRATo32Bit(t *testing.T) {
	test := 12.0
	result := raTo32bitSteps(test)
	expected := uint32(math.Pow(2, 32) / 2.0)
	if result != expected {
		t.Errorf("TestRATo32Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = 12.5
	result = raTo32bitSteps(test)
	thirty_min := uint32(math.Pow(2, 32) / 24.0 / 2.0)
	expected = uint32(math.Pow(2, 32)/2.0) + thirty_min
	if result != expected {
		t.Errorf("TestRATo32Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = 13.0
	result = raTo32bitSteps(test)
	expected = uint32(math.Pow(2, 32)/2.0) + thirty_min + thirty_min
	if result != expected {
		t.Errorf("TestRATo32Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

}

func TestDecTo16Bit(t *testing.T) {
	test := 45.0
	result := decTo16bitSteps(test)
	expected := uint16(math.Pow(2, 16) / 360.0 * 45.0)
	if result != expected {
		t.Errorf("TestdecTo16Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = -45.0
	result = decTo16bitSteps(test)
	expected = uint16(math.Pow(2, 16) / 360.0 * (360.0 - 45.0))
	if result != expected {
		t.Errorf("TestdecTo16Bit(%v) failed, expected %v, got %v", test, expected, result)
	}
}

func TestDecTo32Bit(t *testing.T) {
	test := 45.0
	result := decTo32bitSteps(test)
	expected := uint32(math.Pow(2, 32) / 360.0 * 45.0)
	if result != expected {
		t.Errorf("TestdecTo32Bit(%v) failed, expected %v, got %v", test, expected, result)
	}

	test = -45.0
	result = decTo32bitSteps(test)
	expected = uint32(math.Pow(2, 32) / 360.0 * (360.0 - 45.0))
	if result != expected {
		t.Errorf("TestdecTo32Bit(%v) failed, expected %v, got %v", test, expected, result)
	}
}

func TestRev16ToRA(t *testing.T) {
	// _ := binary.BigEndian.Uint16([]byte{128, 0})
}

func TestRev16ToDec(t *testing.T) {

}

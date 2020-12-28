package telescope

/*
 * Functions for Nexstar telescopes
 */

import (
	"fmt"
	"math"
)

/*
 * Nexstar supports querying in RA/Dec as well as Alt/Azm, but goto only
 * works in RA/Dec mode, so we don't support Alt/Azm
 */
func NewCoordinateNexstar(ra_bytes []byte, dec_bytes []byte, highp bool) Coordinates {
	var ra, dec float64
	if highp {
		rs := StepsToUint32(ra_bytes)
		ds := StepsToUint32(dec_bytes)
		ra = uint32StepsToRA(rs)
		dec = uint32StepsToDec(ds)
	} else {
		rs := StepsToUint16(ra_bytes)
		ds := StepsToUint16(dec_bytes)
		ra = uint16StepsToRA(rs)
		dec = uint16StepsToDec(ds)
	}
	return Coordinates{
		RA:  ra,
		Dec: dec,
	}
}

// Converts our RA/Dec to an ASCII string format for Nexstar
func (c *Coordinates) Nexstar(highp bool) string {
	var ra, dec uint32
	var fmtstr string
	if !highp {
		ra = uint32(raTo16bitSteps(c.RA))
		dec = uint32(decTo16bitSteps(c.Dec))
		fmtstr = "%04X,%04X"
	} else {
		ra = raTo32bitSteps(c.RA)
		dec = decTo32bitSteps(c.Dec)
		fmtstr = "%08X,%08X"
	}
	return fmt.Sprintf(fmtstr, ra, dec)
}

/*
 * Functions to convert the Nexstar ASCII RA/DEC & ALT/AZM
 * steps to uint32/16
 */
func StepsToUint32(steps []byte) uint32 {
	var a, b, c, d byte
	fmt.Sscanf(string(steps), "%02x%02x%02x%02x", &a, &b, &c, &d)
	return uint32(a)<<24 + uint32(b)<<16 + uint32(c)<<8 + uint32(d)
}

func StepsToUint16(steps []byte) uint16 {
	var a, b byte
	fmt.Sscanf(string(steps), "%02x%02x", &a, &b)
	return uint16(a)<<8 + uint16(b)
}

func uint32StepsToDec(steps uint32) float64 {
	s := int64(steps)
	// check if negative value
	if float64(s) > math.Pow(2, 32)/2.0 {
		s = (int64(math.Pow(2, 32)) - s) * -1
	}

	resolution := math.Pow(2, 32) / 360.0
	return float64(s) / resolution
}

func uint32StepsToRA(steps uint32) float64 {
	hrs := float64(steps) / math.Pow(2, 32) * 24.0
	return hrs
}

func uint16StepsToDec(steps uint16) float64 {
	s := int64(steps)
	// check if negative value
	if float64(s) > math.Pow(2, 16)/2.0 {
		s = (int64(math.Pow(2, 16)) - s) * -1
	}

	resolution := math.Pow(2, 16) / 360.0
	return float64(s) / resolution
}

func uint16StepsToRA(steps uint16) float64 {
	hrs := float64(steps) / math.Pow(2, 16) * 24.0
	return hrs
}

func raTo32bitSteps(ra float64) uint32 {
	return uint32(math.Pow(2, 32) * ra / 24.0)
}

func raTo16bitSteps(ra float64) uint16 {
	return uint16(math.Pow(2, 16) * ra / 24.0)
}

func decTo32bitSteps(dec float64) uint32 {
	if dec < 0 {
		return uint32(math.Pow(2, 32) / 360.0 * (360.0 + dec))
	} else {
		return uint32(dec / 360.0 * math.Pow(2, 32))
	}
}

func decTo16bitSteps(dec float64) uint16 {
	if dec < 0 {
		return uint16(math.Pow(2, 16) / 360.0 * (360.0 + dec))
	} else {
		return uint16(dec / 360.0 * math.Pow(2, 16))
	}
}

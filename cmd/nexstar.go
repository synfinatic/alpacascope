package main

import (
	"fmt"
	"math"
	"net"
	"time"

	alpaca "github.com/synfinatic/alpaca-gateway/alpaca"

	log "github.com/sirupsen/logrus"
)

func handleNexstar(conn net.Conn, t *alpaca.Telescope) {
	buf := make([]byte, 1024)
	rlen, err := conn.Read(buf)
	for {
		if err != nil {
			break
		}
		reply := process_command(t, rlen, buf)
		if len(reply) > 0 {
			_, err = conn.Write([]byte(reply))
			if err != nil {
				log.Errorf("writing reply to NexStar client: %s", err.Error())
			}
		} else {
			log.Errorf("command '%s' returned a zero length reply", string(buf))
		}
		rlen, err = conn.Read(buf)
	}
	if err.Error() != "EOF" {
		log.Errorf("conn.Read() returned error: %s", err.Error())
	}
}

func process_command(t *alpaca.Telescope, len int, buf []byte) string {
	var ret string = ""
	var err error
	if log.IsLevelEnabled(log.DebugLevel) {
		var strbuf string
		for i := 1; i < len; i++ {
			strbuf = fmt.Sprintf("%s %d", strbuf, buf[i])
		}
		if buf[0] != 'e' {
			log.Debugf("Received %d bytes [%s]: %c %s", len, string(buf), buf[0], strbuf)
		}
	}

	// single byte commands
	switch buf[0] {
	case 'E':
		ra, dec, err := t.GetRaDec()
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
			return ret
		}

		ra_int := raTo16bitSteps(ra)
		dec_int := decTo16bitSteps(dec)
		ret = fmt.Sprintf("%04X,%04X#", ra_int, dec_int)

	case 'e':
		// Get Precise RA/DEC
		ra, dec, err := t.GetRaDec()
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
			return ret
		}

		ra_int := raTo32bitSteps(ra)
		dec_int := decTo32bitSteps(dec)
		ret = fmt.Sprintf("%08X,%08X#", ra_int, dec_int)

	case 'Z':
		// Get AZM/ALT.  Note that AZM is 0->360, while Alt is -90->90
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
			return ret
		}

		azm_int := uint32(azm / 360.0 * math.Pow(2, 16))
		alt_int := uint32(alt / 360.0 * math.Pow(2, 16))
		ret = fmt.Sprintf("%04X,%04X#", azm_int, alt_int)

	case 'z':
		// Get Precise AZM/ALT
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
			return ret
		}

		azm_int := uint32(azm / 360.0 * 4294967296.0)
		alt_int := uint32(alt / 360.0 * 4294967296.0)
		ret = fmt.Sprintf("%08X,%08X#", azm_int, alt_int)

	case 't':
		// Get tracking mode
		// 0 = off, 1 alt/az, 2 EQ North, 3 EQ south
		ret = "1#" // hard code to alt/az for now

	case 'T':
		// Set tracking mode
		ret = "#"

	case 'V':
		// Get Version
		ret = "50#"

	case 'P':
		// Slew
		err = executeSlew(t, buf)
		if err != nil {
			log.Errorf("Unable to slew: %s", err.Error())
			return ret
		}
		ret = "#"

	case 's':
		// Precise sync aka: Align on object.  Uses the same math as 'e'
		ra_bytes := StepsToUint32(buf[1:9])
		dec_bytes := StepsToUint32(buf[10:18])
		ra := uint32StepsToRA(ra_bytes)
		dec := uint32StepsToDec(dec_bytes)
		err = t.PutSyncToCoordinates(ra, dec)
		ret = "#"

	case 'S':
		// sync aka: Align on object.  Uses same math as 'E'
		ra_bytes := StepsToUint16(buf[1:5])
		dec_bytes := StepsToUint16(buf[6:10])
		ra := uint16StepsToRA(ra_bytes)
		dec := uint16StepsToDec(dec_bytes)
		err = t.PutSyncToCoordinates(ra, dec)
		ret = "#"

	case 'r':
		// precise goto Ra/Dec values.  RA is in hours, Dec in deg
		ra_bytes := StepsToUint32(buf[1:9])
		dec_bytes := StepsToUint32(buf[10:18])
		log.Debugf("RAW RA: %d\t\tDec: %d", ra_bytes, dec_bytes)
		ra := uint32StepsToRA(ra_bytes)
		dec := uint32StepsToDec(dec_bytes)
		if log.IsLevelEnabled(log.DebugLevel) {
			ra_hms := uint32StepsToHMS(ra_bytes)
			dec_hms := uint32StepsToHMS(dec_bytes)
			log.Debugf("Goto RA: %dh %dm%gs\t\tDec: %dh %dm %gs", ra_hms.Hours,
				ra_hms.Minutes, ra_hms.Seconds, dec_hms.Hours, dec_hms.Minutes, dec_hms.Seconds)
		}

		log.Debugf("Goto RA: %v\t\tDec: %v", ra, dec)
		err = t.PutSlewToCoordinatestAsync(ra, dec)
		ret = "#"

	case 'R':
		// goto Ra/Dec values
		ra_bytes := StepsToUint16(buf[1:5])
		dec_bytes := StepsToUint16(buf[6:10])
		ra := uint16StepsToRA(ra_bytes)
		dec := uint16StepsToDec(dec_bytes)
		err = t.PutSlewToCoordinatestAsync(ra, dec)
		ret = "#"

	case 'W':
		// set location
		lat, long := bytes_to_latlong(buf[1:8])
		err = t.PutSiteLatitude(lat)
		if err != nil {
			log.Errorf("Error setting site latitude: %s", err.Error())
		}
		err = t.PutSiteLongitude(long)
		if err != nil {
			log.Errorf("Error setting site longitude: %s", err.Error())
		}
		ret = "#"

	case 'H':
		// set date/time
		tz_val := int(buf[7])
		// UTC-X values are stored as 256-X so need to be converted back to a negative if tz_val > 128 {
		if tz_val > 128 {
			tz_val = (256 - tz_val) * -1
		}
		tz := time.FixedZone("Telescope Time", tz_val*60*60)
		date := time.Date(
			int(buf[6])+2000,   // year V
			time.Month(buf[4]), // month T
			int(buf[5]),        // day U
			int(buf[1]),        // hour Q
			int(buf[2]),        // min R
			int(buf[3]),        // sec S
			0,                  // nanosec
			tz)
		err = t.PutUTCDate(date)
		ret = "#"

	case 'M':
		// cancel GOTO
		err = t.PutAbortSlew()
		ret = "#"

	default:
		log.Errorf("Unsupported command: %c", buf[0])
		ret = "#"
	}

	if err != nil {
		log.Errorf("Error talking to scope: %s", err.Error())
	}
	return ret
}

/*
 * So ASCOM and NexStar impliment slewing a little differently.
 *
 * NexStar uses Azm/Alt, direction, speed (0, 2, 5, 7, 9).
 * ASCOM uses Azm/Alt + direction w/ speed as a function of -3 to +3
 *
 * In both cases, it will slew until a new slew command with a speed of 0 is
 * issued to stop the slew so we can handle this without any state.
 *
 * Note that NexStar supports both fixed & variable slew rates, however SkySafari
 * only uses the fixed type and ASCOM has no concept of variable rates so we will
 * treat variable as fixed.
 */
func executeSlew(t *alpaca.Telescope, buf []byte) error {
	var axis alpaca.AxisType = alpaca.AxisAzmRa
	var positive_direction bool = false
	var rate int = 0 // SkySafari uses direction with speeds of 0,2,5,7,9 but ASCOM uses axis with speeds -3 to 3

	switch int(buf[2]) {
	case 16:
		// Azm/RA
		axis = alpaca.AxisAzmRa
	case 17:
		// Alt/Dec
		axis = alpaca.AxisAltDec
	default:
		log.Errorf("Unknown axis: %d", int(buf[2]))
	}

	switch int(buf[3]) {
	case 6, 36:
		positive_direction = true
	case 7, 37:
		positive_direction = false
	default:
		log.Errorf("Unknown direction: %d", int(buf[3]))
	}

	rate = rate_to_ascom(positive_direction, int(buf[4]))

	// buf[1] is variable vs. fixed rate, but we always use fixed
	// buf[5] is the "slow" variable rate which we always ignore
	// Last two bytes (6, 7) are always 0

	err := t.PutMoveAxis(axis, rate)
	return err
}

// Converts the direction & rate to an ASCOM rate
func rate_to_ascom(direction bool, rate int) int {
	switch rate {
	case 0:
		rate = 0
	case 1, 2, 3:
		rate = 1
	case 4, 5, 6:
		rate = 2
	case 7, 8, 9:
		rate = 3
	}
	// flip our direction?
	if !direction {
		rate *= -1
	}
	return rate
}

// convert ABCDEFGH bytes to lat/long
func bytes_to_latlong(b []byte) (float64, float64) {
	var lat float64 = float64(b[0]) + float64(b[1])/60.0 + float64(b[2])/3600.0
	if b[3] > 0 {
		lat *= -1
	}
	var long float64 = float64(b[4]) + float64(b[5])/60.0 + float64(b[6])/3600.0
	if b[7] > 0 {
		long *= -1
	}
	return lat, long
}

type HMS struct {
	Hours   int
	Minutes int
	Seconds float64
}

/*
 * 16bit percent revolution to H:M:S.s
 */
func uint16StepsToHMS(rev uint16) HMS {
	hours := uint16StepsToRA(rev)
	h := int(hours)
	remainder := hours - float64(h)
	m := int(remainder * 60.0)
	// remove the minutes from the hours and leave fractions of minute
	frac_minute := hours - float64(h) - float64(m)/60.0
	s := 60.0 * frac_minute
	return HMS{
		Hours:   h,
		Minutes: m,
		Seconds: s,
	}
}

/*
 * 32bit percent revolution to H:M:S.s
 */
func uint32StepsToHMS(rev uint32) HMS {
	hours := uint32StepsToRA(rev)
	h := int(hours)
	remainder := hours - float64(h)
	m := int(remainder * 60.0)
	// remove the minutes from the hours and leave fractions of minute
	frac_minute := hours - float64(h) - float64(m)/60.0
	s := 60.0 * frac_minute
	return HMS{
		Hours:   h,
		Minutes: m,
		Seconds: s,
	}
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

/*
 * Convert steps to RA (hours) or Dec (deg)
 * Dec is +90 -> -90 deg
 * RA is in hours.frac_hour
 */

func uint16StepsToDec(steps uint16) float64 {
	s := int32(steps)
	// check if negative value
	if float64(s) > math.Pow(2, 16)/2.0 {
		s = (int32(math.Pow(2, 16)) - s) * -1
	}

	resolution := math.Pow(2, 16) / 360.0

	return float64(s) / resolution
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

func uint16StepsToRA(steps uint16) float64 {
	hrs := float64(steps) / math.Pow(2, 16) * 24.0
	return hrs
}

func uint32StepsToRA(steps uint32) float64 {
	hrs := float64(steps) / math.Pow(2, 32) * 24.0
	return hrs
}

/*
 * Convert RA (hours) or Dec (deg) to Nexstar steps
 */

func decTo16bitSteps(ra float64) uint16 {
	if ra < 0 {
		return uint16(math.Pow(2, 16) / 360.0 * (360.0 + ra))
	}
	return uint16(ra / 360.0 * math.Pow(2, 16))
}

func decTo32bitSteps(ra float64) uint32 {
	if ra < 0 {
		// need to convert to positive value
		return uint32(math.Pow(2, 32) / 360.0 * (360.0 + ra))
	}
	return uint32(ra / 360.0 * math.Pow(2, 32))
}

func raTo16bitSteps(hours float64) uint16 {
	return uint16(math.Pow(2, 16) * hours / 24.0)
}

func raTo32bitSteps(hours float64) uint32 {
	return uint32(math.Pow(2, 32) * hours / 24.0)
}

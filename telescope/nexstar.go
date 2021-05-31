package telescope

import (
	"fmt"
	"math"
	"net"
	"time"

	alpaca "github.com/synfinatic/alpacascope/alpaca"

	log "github.com/sirupsen/logrus"
)

type NexStar struct {
}

func NewNexStar() *NexStar {
	return &NexStar{}
}

func (n *NexStar) HandleConnection(conn net.Conn, t *alpaca.Telescope) {
	buf := make([]byte, 1024)

	defer conn.Close() // make sure we close connection before we leave
	rlen, err := conn.Read(buf)
	for {
		if err != nil {
			break
		}
		reply := nexstar_command(t, rlen, buf)
		if len(reply) > 0 {
			_, err = conn.Write(reply)
			if err != nil {
				log.Errorf("writing reply to NexStar client: %s", err.Error())
			}
		} else {
			log.Errorf("command '%s' returned a zero length reply", string(buf))
		}
		rlen, err = conn.Read(buf)
	}
	// Will get this any time the client sends a Fin, so don't log that
	if err.Error() != "EOF" {
		log.Errorf("conn.Read() returned error: %s", err.Error())
	}
}

func nexstar_command(t *alpaca.Telescope, len int, buf []byte) []byte {
	var ret_val []byte
	ret := ""
	var err error
	if log.IsLevelEnabled(log.DebugLevel) {
		var strbuf string
		for i := 1; i < len; i++ {
			strbuf = fmt.Sprintf("%s %d", strbuf, buf[i])
		}
		if buf[0] != 'e' {
			log.Debugf("Received %d bytes [%s]: %c %s", len, string(buf[:len]), buf[0], strbuf)
		}
	}

	// single byte commands
	switch buf[0] {
	case 'K':
		// echo next byte
		ret = fmt.Sprintf("%c#", buf[1])

	case 'e', 'E':
		ra, dec, err := t.GetRaDec()
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
		} else {
			radec := Coordinates{
				RA:  ra,
				Dec: dec,
			}

			high_precision := true
			if buf[0] == 'E' {
				high_precision = false
			}
			ret = fmt.Sprintf("%s#", radec.Nexstar(high_precision))
		}

	case 'Z':
		// Get AZM/ALT.  Note that AZM is 0->360, while Alt is -90->90
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
		} else {
			azm_int := uint32(azm / 360.0 * math.Pow(2, 16))
			alt_int := uint32(alt / 360.0 * math.Pow(2, 16))
			ret = fmt.Sprintf("%04X,%04X#", azm_int, alt_int)
		}

	case 'z':
		// Get Precise AZM/ALT
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
		} else {
			azm_int := uint32(azm / 360.0 * 4294967296.0)
			alt_int := uint32(alt / 360.0 * 4294967296.0)
			ret = fmt.Sprintf("%08X,%08X#", azm_int, alt_int)
		}

	case 't':
		// get tracking mode
		mode, err := t.GetTracking()
		if err != nil {
			log.Errorf("Unable to get tracking mode: %s", err.Error())
		} else {
			ret = fmt.Sprintf("%d#", mode)
		}

	case 'T':
		// set tracking mode
		var tracking_mode alpaca.TrackingMode
		fmt.Sscanf(string(buf[1]), "%d", &tracking_mode)
		err = t.PutTracking(tracking_mode)
		ret = "#"

	case 'V':
		// Get Version
		ret = "50#"

	case 'P':
		// Slew
		err = executeSlew(t, buf)
		ret = "#"

	case 's':
		// Precise sync aka: Align on object.  Uses the same math as 'e'
		radec := NewCoordinateNexstar(buf[1:9], buf[10:18], true)
		err = t.PutSyncToCoordinates(radec.RA, radec.Dec)
		ret = "#"

	case 'S':
		// sync aka: Align on object.  Uses same math as 'E'
		radec := NewCoordinateNexstar(buf[1:5], buf[6:10], false)
		err = t.PutSyncToCoordinates(radec.RA, radec.Dec)
		ret = "#"

	case 'r':
		// precise goto Ra/Dec values.  RA is in hours, Dec in deg
		radec := NewCoordinateNexstar(buf[1:9], buf[10:18], true)
		err = t.PutSlewToCoordinatestAsync(radec.RA, radec.Dec)
		ret = "#"

	case 'R':
		// goto Ra/Dec values
		radec := NewCoordinateNexstar(buf[1:5], buf[6:10], false)
		err = t.PutSlewToCoordinatestAsync(radec.RA, radec.Dec)
		ret = "#"

	case 'w':
		// get location
		failed := false
		lat, err := t.GetSiteLatitude()
		if err != nil {
			log.Errorf("Error talking to scope: %s", err.Error())
			failed = true
		}

		long, err := t.GetSiteLatitude()
		if err != nil {
			// logged at the end
			failed = true
		}

		if !failed {
			ret_val = LatLongToNexstar(lat, long)
			ret_val = append(ret_val, '#')
		}

	case 'W':
		// set location
		lat, long := NexstarToLatLong(buf[1:9])
		err = t.PutSiteLatitude(lat)
		if err != nil {
			log.Errorf("Error talking to scope: %s", err.Error())
		}
		err = t.PutSiteLongitude(long)
		// logged at the end
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

	case 'L':
		// Goto in progress??
		var slewing bool
		slewing, err = t.GetSlewing()
		if slewing {
			ret = "1#"
		} else {
			ret = "0#"
		}

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

	// convert our return string to the ret_val
	if ret != "" {
		ret_val = []byte(ret)
	}
	return ret_val
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

	rate = nexstar_rate_to_ascom(positive_direction, int(buf[4]))

	// buf[1] is variable vs. fixed rate, but we always use fixed
	// buf[5] is the "slow" variable rate which we always ignore
	// Last two bytes (6, 7) are always 0

	err := t.PutMoveAxis(axis, rate)
	return err
}

// Converts the direction & rate to an ASCOM rate
func nexstar_rate_to_ascom(direction bool, rate int) int {
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
func NexstarToLatLong(b []byte) (float64, float64) {
	var lat float64 = float64(b[0]) + float64(b[1])/60.0 + float64(b[2])/3600.0
	if b[3] == 1 {
		lat *= -1.0
	}
	var long float64 = float64(b[4]) + float64(b[5])/60.0 + float64(b[6])/3600.0
	if b[7] == 1 {
		long *= -1.0
	}
	return lat, long
}

// Convert Lat/Long to ABCDEFGH bytes
func LatLongToNexstar(lat float64, long float64) []byte {
	var a, b, c, d, e, f, g, h byte
	// West & South are negative
	if lat < 0 {
		d = 1
		lat = math.Abs(lat)
	}
	if long < 0 {
		h = 1
		long = math.Abs(long)
	}

	// latitude
	a = byte(int(lat))
	remain := lat - math.Floor(lat)
	b = byte(int(remain * 60.0))
	frac_minute := remain - float64(b)/60.0
	c = byte(int(frac_minute * 60.0 * 60.0))

	// longitude
	e = byte(int(long))
	remain = long - math.Floor(long)
	f = byte(int(remain * 60.0))
	frac_minute = remain - float64(f)/60.0
	g = byte(int(frac_minute * 60.0 * 60.0))

	return []byte{a, b, c, d, e, f, g, h}
}

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

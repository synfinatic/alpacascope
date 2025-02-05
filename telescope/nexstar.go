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
	AutoTrack bool // ensure tracking is enabled for goto
}

func NewNexStar(autoTrack bool) *NexStar {
	return &NexStar{
		AutoTrack: autoTrack,
	}
}

func (n *NexStar) HandleConnection(conn net.Conn, t *alpaca.Telescope) {
	buf := make([]byte, 1024)

	defer conn.Close() // make sure we close connection before we leave
	rlen, err := conn.Read(buf)
	for {
		if err != nil {
			break
		}
		reply := n.nexstarCommand(t, rlen, buf)
		wlen := len(reply)
		log.Debugf("our reply %d bytes: %v", wlen, reply)

		if wlen > 0 {
			x, err := conn.Write(reply)
			if err != nil {
				log.Errorf("writing reply to NexStar client: %s", err.Error())
			} else if x != wlen {
				log.Errorf("only wrote %d of %d bytes", x, wlen)
			}
		} else {
			log.Errorf("command '%s' returned a zero length reply", string(buf))
		}
		rlen, err = conn.Read(buf) // blocks for next command
	}

	// Will get this any time the client sends a Fin, so don't log that
	if err.Error() != "EOF" {
		log.Errorf("conn.Read() returned error: %s", err.Error())
	}
}

func (n *NexStar) nexstarCommand(t *alpaca.Telescope, len int, buf []byte) []byte {
	var retVal []byte
	ret := ""
	var err error
	if log.IsLevelEnabled(log.DebugLevel) {
		var strbuf string
		for i := 1; i < len; i++ {
			strbuf = fmt.Sprintf("%s %d", strbuf, buf[i])
		}
		log.Debugf("Received %d bytes [%s]: %c %s", len, string(buf[:len]), buf[0], strbuf)
	}

	// single byte commands
	switch buf[0] {
	case 'K':
		// echo next byte
		retVal = []byte{buf[1], '#'}

	case 'e', 'E':
		ra, dec, err := t.GetRaDec()
		if err != nil {
			log.Errorf("unable to get RA/DEC: %s", err.Error())
		} else {
			radec := Coordinates{
				RA:  ra,
				Dec: dec,
			}

			highPrecision := true
			if buf[0] == 'E' {
				highPrecision = false
			}
			ret = fmt.Sprintf("%s#", radec.Nexstar(highPrecision))
		}

	case 'Z':
		// Get AZM/ALT.  Note that AZM is 0->360, while Alt is -90->90
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("unable to get AZM/ALT: %s", err.Error())
		} else {
			asmInt := uint32(azm / 360.0 * math.Pow(2, 16))
			altInt := uint32(alt / 360.0 * math.Pow(2, 16))
			ret = fmt.Sprintf("%04X,%04X#", asmInt, altInt)
		}

	case 'z':
		// Get Precise AZM/ALT
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("unable to get AZM/ALT: %s", err.Error())
		} else {
			azmInt := uint32(azm / 360.0 * 4294967296.0)
			altInt := uint32(alt / 360.0 * 4294967296.0)
			ret = fmt.Sprintf("%08X,%08X#", azmInt, altInt)
		}

	case 't':
		// get tracking mode
		mode, err := t.GetTracking()
		if err != nil {
			log.Errorf("unable to get tracking mode: %s", err.Error())
		} else {
			ret = fmt.Sprintf("%c#", mode)
		}

	case 'T':
		// set tracking mode
		var trackingMode alpaca.TrackingMode
		_, _ = fmt.Sscanf(string(buf[1]), "%d", &trackingMode)
		err = t.PutTracking(trackingMode)
		ret = "#"

	case 'V':
		// Get Version
		ret = "50#"

	case 'P':
		// Pass through commands for Slew, GPS, RTC, etc
		if int(buf[3]) == 254 {
			// Get passthrough device version
			switch int(buf[2]) {
			case 16, 17:
				// mounts
				retVal = []byte{5, 0, '#'}
			case 176, 178:
				// GPS & RTC
				retVal = []byte{1, 6, '#'}
			default:
				log.Errorf("invalid device version device type: %d", int(buf[3]))
			}
		} else {
			switch int(buf[2]) {
			case 176:
				// GPS
				retVal, err = getGPS(t, buf)
			case 178:
				// RTC
				retVal, err = getRTC(t, buf)
			case 16, 17:
				err = executeSlew(t, buf)
				ret = "#"
			default:
				log.Errorf("unsupported P command: %c%c%c%c%c%c%c",
					buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6])
				ret = "#"
			}
		}

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
		if n.AutoTrack {
			// auto-enable tracking?
			mode, err := t.GetTracking()
			if err != nil {
				log.Errorf("unable to get tracking mode: %s", err.Error())
			} else {
				if mode == alpaca.NotTracking {
					err = t.PutTracking(alpaca.AltAz) // need any non-NotTracking value for true
					if err != nil {
						log.Errorf("unable to auto-enable tracking: %s", err.Error())
					}
				}
			}
		}
		radec := NewCoordinateNexstar(buf[1:9], buf[10:18], true)
		err = t.PutSlewToCoordinatestAsync(radec.RA, radec.Dec)
		ret = "#"

	case 'R':
		// goto Ra/Dec values
		if n.AutoTrack {
			// auto-enable tracking?
			mode, err := t.GetTracking()
			if err != nil {
				log.Errorf("unable to get tracking mode: %s", err.Error())
			} else {
				if mode == alpaca.NotTracking {
					err = t.PutTracking(alpaca.AltAz) // need any non-NotTracking value for true
					if err != nil {
						log.Errorf("unable to auto-enable tracking: %s", err.Error())
					}
				}
			}
		}
		radec := NewCoordinateNexstar(buf[1:5], buf[6:10], false)
		err = t.PutSlewToCoordinatestAsync(radec.RA, radec.Dec)
		ret = "#"

	case 'w':
		// get location
		failed := false
		lat, err := t.GetSiteLatitude()
		if err != nil {
			log.Errorf("talking to scope: %s", err.Error())
			failed = true
		}

		long, err := t.GetSiteLongitude()
		if err != nil {
			// logged at the end
			failed = true
		}

		if !failed {
			retVal = LatLongToNexstar(lat, long)
			retVal = append(retVal, '#')
		}

	case 'W':
		// set location
		lat, long := NexstarToLatLong(buf[1:9])
		err = t.PutSiteLatitude(lat)
		if err != nil {
			log.Errorf("talking to scope: %s", err.Error())
		}
		err = t.PutSiteLongitude(long)
		// logged at the end
		ret = "#"

	case 'h':
		// get date/time
		utcDate, err := t.GetUTCDate()
		if err != nil {
			log.Errorf("computer returned no UTC date: %s", err.Error())
			/*
				ret_val = []byte{
					0, 0, 0,
					0, 0, 0,
					0, 0, '#',
				}
			*/
		} else {
			h, m, s := utcDate.Clock()
			var isDST byte = 0
			if utcDate.IsDST() {
				isDST = 1
			}

			y, M, d := utcDate.Date()
			y -= 2000 // need to output a single byte for the year
			retVal = []byte{
				byte(h), byte(m), byte(s), // H:M:S
				byte(M), byte(d), byte(y), // M:D:Y
				0, // always UTC
				isDST, '#',
			}
		}

	case 'H':
		// set date/time
		tzVal := int(buf[7])
		// UTC-X values are stored as 256-X so need to be converted back to a negative if tz_val > 128 {
		if tzVal > 128 {
			tzVal = (256 - tzVal) * -1
		}
		tz := time.FixedZone("Telescope Time", tzVal*60*60)
		date := time.Date(
			int(buf[6])+2000,   // year V
			time.Month(buf[4]), // month T
			int(buf[5]),        // day U
			int(buf[1]),        // hour Q
			int(buf[2]),        // min R
			int(buf[3]),        // sec S
			0,                  // nanosec
			tz)
		log.Errorf("client set date to: %s", date.String())
		err = t.PutUTCDate(date)
		ret = "#"

	case 'J':
		// is alignment complete?
		// since Alpaca has no similar command, aways return true
		retVal = []byte{1, '#'}

	case 'L':
		// Goto in progress??
		var slewing bool
		slewing, err = t.GetSlewing()
		if slewing {
			ret = "1#"
		} else {
			ret = "0#"
		}

	case 'm':
		// Model- hard code to 6/8 SE
		retVal = []byte{12, '#'}

	case 'M':
		// cancel GOTO
		err = t.PutAbortSlew()
		ret = "#"

	default:
		log.Errorf("unsupported command: %c", buf[0])
		ret = "#"
	}

	if err != nil {
		log.Errorf("error talking to scope: %s", err.Error())
	}

	// convert our return string to the ret_val
	if ret != "" {
		retVal = []byte(ret)
	}
	return retVal
}

/*
 * So ASCOM and NexStar implement slewing a little differently.
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
	var positiveDirection bool = false
	var rate int // SkySafari uses direction with speeds of 0,2,5,7,9 but ASCOM uses axis with speeds -3 to 3

	switch int(buf[2]) {
	case 16:
		// Azm/RA
		axis = alpaca.AxisAzmRa
	case 17:
		// Alt/Dec
		axis = alpaca.AxisAltDec
	default:
		log.Errorf("unknown axis: %d", int(buf[2]))
	}

	switch int(buf[3]) {
	case 6, 36:
		positiveDirection = true
	case 7, 37:
		positiveDirection = false
	default:
		log.Errorf("unknown direction: %d", int(buf[3]))
	}

	rate = nextstartRateToASCOM(positiveDirection, int(buf[4]))

	// buf[1] is variable vs. fixed rate, but we always use fixed
	// buf[5] is the "slow" variable rate which we always ignore
	// Last two bytes (6, 7) are always 0

	err := t.PutMoveAxis(axis, rate)
	return err
}

/*
 * getGPS is another 'P' command which returns date, location, etc from the GPS
 * unit.  Note that the time & location values from these commands are supposed
 * to come from the GPS and not RTC or hand controller, but ASCOM doesn't treat
 * them differently since it is up to the driver.
 *
 * However, GPS was v1.6+ while hand controller was v2.3+ so we can expect
 * some software to prefer/only support the GPS and not hand controller
 * (Stellarium?)
 */
func getGPS(t *alpaca.Telescope, buf []byte) ([]byte, error) {
	retVal := []byte{}
	switch int(buf[2]) {
	case 55:
		_, err := t.GetSiteLatitude()
		if err != nil {
			// GPS is not linked
			return []byte{0, '#'}, nil
		}
		// GPS is linked
		return []byte{1, '#'}, nil
	case 1:
		// Latitude
		lat, err := t.GetSiteLatitude()
		if err != nil {
			log.Errorf("unable to GetSiteLatitude(): %s", err.Error())
			return retVal, err
		}
		retVal = LatLongToGPS(lat)
	case 2:
		// Longitude
		long, err := t.GetSiteLongitude()
		if err != nil {
			log.Errorf("unable to GetSiteLongitude(): %s", err.Error())
			return retVal, err
		}
		retVal = LatLongToGPS(long)
	case 3:
		// Date: m, d
		utcDate, err := t.GetUTCDate()
		if err != nil {
			log.Errorf("GPS returned no UTC date: %s", err.Error())
			return retVal, err
		}
		_, m, d := utcDate.Date()
		retVal = []byte{byte(m), byte(d), '#'}
	case 4:
		// Year: (x * 256) + y = year
		utcDate, err := t.GetUTCDate()
		if err != nil {
			log.Errorf("GPS returned no UTC date: %s", err.Error())
			return retVal, err
		}
		year, _, _ := utcDate.Date()
		var x int = year / 256
		var y int = year % 256
		retVal = []byte{byte(x), byte(y), '#'}
	case 51:
		// Time: h, m, s
		utcDate, err := t.GetUTCDate()
		if err != nil {
			log.Errorf("GPS returned no UTC date: %s", err.Error())
			return retVal, err
		}
		h, m, s := utcDate.Date()
		retVal = []byte{byte(h), byte(m), byte(s), '#'}
	}
	return retVal, nil
}

/*
 * getRTC is another 'P' command which returns date from the real time clock
 * in the mount.  Note that the time values from these commands are supposed
 * to come from the RTC and not GPS or hand controller, but ASCOM doesn't treat
 * them differently since it is up to the driver.
 *
 * RTC is v1.6+ for get and v3.01+ for set.
 */
func getRTC(t *alpaca.Telescope, buf []byte) ([]byte, error) {
	switch int(buf[2]) {
	case 3, 4, 51:
		// These commands to get date, time and year are the same
		// as the GPS commands, so reuse that code
		return getGPS(t, buf)
	default:
		log.Errorf("unsupported RTC P command: %c%c%c%c%c%c%c",
			buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6])
	}
	return []byte{}, nil
}

// Converts the direction & rate to an ASCOM rate
func nextstartRateToASCOM(direction bool, rate int) int {
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

// Convert Lat/Long to ABCDEFGH bytes for hand controller
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
	fracMinute := remain - float64(b)/60.0
	c = byte(int(fracMinute * 60.0 * 60.0))

	// longitude
	e = byte(int(long))
	remain = long - math.Floor(long)
	f = byte(int(remain * 60.0))
	fracMinute = remain - float64(f)/60.0
	g = byte(int(fracMinute * 60.0 * 60.0))

	return []byte{a, b, c, d, e, f, g, h}
}

/*
 * Convert a Lat OR Long to GPS fraction of a rotation "XYZ#"
 *
 * This is really a 24bit integer representing degrees as:
 * (x*65536)+(y*256)+z / 2^24 * 360
 */
func LatLongToGPS(latlong float64) []byte {
	var pos = make([]byte, 4)

	// West & South are negative, so need to convert to positive degrees
	if latlong < 0 {
		latlong += 360.0
	}

	position := uint32(latlong * float64(2^24) / 360.0)
	pos[0] = byte(position & 0x00ff0000 >> 16)
	pos[1] = byte(position & 0x0000ff00 >> 8)
	pos[2] = byte(position & 0x000000ff)
	pos[3] = '#'
	return pos
}

/*
 * Nexstar supports querying in RA/Dec as well as Alt/Azm, but goto only
 * works in RA/Dec mode, so we don't support Alt/Azm
 */
func NewCoordinateNexstar(raBytes []byte, decBytes []byte, highp bool) Coordinates {
	var ra, dec float64
	if highp {
		rs := StepsToUint32(raBytes)
		ds := StepsToUint32(decBytes)
		ra = uint32StepsToRA(rs)
		dec = uint32StepsToDec(ds)
	} else {
		rs := StepsToUint16(raBytes)
		ds := StepsToUint16(decBytes)
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
	_, _ = fmt.Sscanf(string(steps), "%02x%02x%02x%02x", &a, &b, &c, &d)
	return uint32(a)<<24 + uint32(b)<<16 + uint32(c)<<8 + uint32(d)
}

func StepsToUint16(steps []byte) uint16 {
	var a, b byte
	_, _ = fmt.Sscanf(string(steps), "%02x%02x", &a, &b)
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

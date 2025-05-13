package telescope

import (
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/alpacascope/alpaca"
)

type LX200 struct {
	AutoTrack      bool // ensure tracking is enabled for goto
	HighPrecision  bool
	TwentyFourHour bool // :H#
	MaxSlew        float64
	MinSlew        float64
	SlewRate       int
	UTCOffset      float64
	haveTime       bool
	haveDate       bool
	hour           int
	minute         int
	second         int
	day            int
	month          int
	year           int
}

func NewLX200(autoTrack, highPrecision, twentyfourhr bool, rates map[string]float64, utcoffset float64) *LX200 {
	state := LX200{
		AutoTrack:      autoTrack,
		HighPrecision:  highPrecision,
		TwentyFourHour: twentyfourhr,
		MaxSlew:        rates["Maximum"],
		MinSlew:        rates["Minimum"],
		SlewRate:       int(rates["Maximum"]),
		UTCOffset:      utcoffset,
	}
	return &state
}

func (state *LX200) HandleConnection(conn net.Conn, t *alpaca.Telescope) {
	buf := make([]byte, 1024)

	defer conn.Close()
	rlen, err := conn.Read(buf)
	for err != nil {
		/*
		 * LX200 has a single byte command <0x06> and variable length commands
		 * which start with a ':' and end with a '#'.  Also, clients may send
		 * multiple commands at once :-/
		 */
		for rlen > 0 {
			reply, consumed := state.lx200Command(t, rlen, buf)
			if len(reply) > 0 {
				_, err = conn.Write(reply)
				if err != nil {
					log.Errorf("writing reply to LX200 client: %s", err.Error())
				}
			} else {
				// many LX200 commands don't generate a reply
				log.Debugf("command '%s' returned a zero length reply", string(buf[0:consumed]))
			}
			rlen -= consumed
			buf = buf[consumed:]
			if rlen > 0 {
				log.Debugf("processing remaining command(s) in buffer: '%s'", string(buf[0:rlen]))
			}
		}
		rlen, err = conn.Read(buf)
	}
	// Will get this any time the client sends a Fin, so don't log that
	if err.Error() != "EOF" {
		log.Errorf("conn.Read() returned error: %s", err.Error())
	}
}

func (state *LX200) lx200Command(t *alpaca.Telescope, cmdlen int, buf []byte) ([]byte, int) {
	var consumed int
	var retVal []byte
	ret := ""
	var err error

	if log.IsLevelEnabled(log.DebugLevel) {
		var strbuf string
		for i := 1; i < cmdlen; i++ {
			strbuf = fmt.Sprintf("%s %d", strbuf, buf[i])
		}
		if buf[0] != 'e' {
			log.Debugf("Received %d bytes [%s]: %c %s", cmdlen, string(buf[0:cmdlen]), buf[0], strbuf)
		}
	}

	// LX200 protocol is a mix of binary and ASCII.
	if buf[0] == 0x06 {
		consumed = 1
		mode, err := t.GetAlignmentMode()
		if err != nil {
			log.Errorf("Unable to determine alignmentmode: %s", err.Error())
		}
		switch mode {
		case alpaca.AlignmentAltAz:
			ret = "A"
		case alpaca.AlignmentPolar, alpaca.AlignmentGermanPolar:
			ret = "P"
		}
	} else if cmdlen < 3 {
		log.Errorf("Unexpected/Invalid command: %s", string(buf[0:cmdlen]))
		consumed = cmdlen
	} else {
		/*
		 * variable length string commands, all which start with a ':' and end with a '#'.
		 * Some commands are toggle (fixed) while others include some kind of variable data so
		 * we'll need to check variable length prefixes.
		 *
		 * To make matters more fun, SkySafari will send multiple commands at a time and so
		 * we have to split them up and process one at a time
		 */

		commands := string(buf)
		endOfCommand := strings.Index(commands, "#")
		consumed = endOfCommand + 1
		cmd := commands[0:consumed]
		log.Debugf("Consumed %d of %d bytes in buffer", consumed, cmdlen)

		// Variable len commands, but we can alway match on the first 3 bytes
		switch cmd[0:3] {
		/*
		 * Unsupported commands
		 * :A - Setting AlignmentMode
		 * :$B - Backlash compensation
		 * :B - Reticule accessory control
		 * :CL# - Sync with object by selenographic coordinates
		 * :D - Distance bars
		 * :f - Fan control
		 * :F - Focuser control works very differently... unclear how to support
		 * :g - GPS
		 * :G0, :G1, :G2 - no idea what this is :(
		 * :Gb - Browse brigher magnitude limit
		 * :GF - field diameter
		 * :GF - faint magnitude limit
		 * :Gh - get high limit
		 * :Gl - Larger size limit
		 * :GM, :GN, :GO, :GP - get site (1, 2, 3, 4) name
		 * :Go - Get lower limit
		 * :Gq - Get minimum quality for find operation
		 * :GVD - Get firmware date
		 * :GVN - Get firmware version
		 * :GVP - Get product name
		 * :GVT - Get firmware time
		 * :Gy# - Get deepsky object string
		 *
		 * :h - home position commands
		 * :I - initialize scope
		 * :L - object library commands
		 * :$Q - PEC control
		 *
		 * :r - field derotator
		 */
		case ":CM":
			// Sync with current target
			err = t.PutSyncToTarget()
			if err != nil {
				log.Errorf("Unable to sync on target: %s", err.Error())
			} else {
				ret = "M31 EX GAL MAG 3.5 SZ178.0'#" // static string like Autostars/LX200GPS
			}

		case ":GA":
			// telescope altitude based on precision config
			alt, err := t.GetAltitude()
			if err != nil {
				log.Errorf("Unable to get telescope altitude (:GA#): %s", err.Error())
				alt = 0.0
			}
			ret = DegreesToStr(alt, state.HighPrecision) + "#"

		case ":Ga":
			// get local time in 12hr format: HH:MM:SS#

		case ":GL":
			// local time in 24hr format: HH:MM:SS#

		case ":GC":
			// Get current date: MM/DD/YY#
			t, err := t.GetUTCDate()
			if err != nil {
				log.Errorf("Unable to get telescope time (:GC#): %s", err.Error())
				t = time.Unix(0, 0)
			}
			y, m, d := t.Date()
			if y > 2000 {
				y -= 2000
			} else if y > 1900 {
				y -= 1900
			}
			ret = fmt.Sprintf("%02d/%02d/%02d#", m, d, y)

		case ":Gc":
			// Get calendar format: 12# or 24#

		case ":GD":
			// telescope declination based on precision config
			alt, err := t.GetDeclination()
			if err != nil {
				log.Errorf("Unable to get telescope declination (:GD#): %s", err.Error())
				alt = 0.0
			}
			ret = DegreesToStr(alt, state.HighPrecision) + "#"

		case ":GZ":
			// telescope azimuth baesd on precision config
			az, err := t.GetAzimuth()
			if err != nil {
				log.Errorf("Unable to get telescope azimuth (:GZ#): %s", err.Error())
				az = 0.0
			}
			ret = DegreesToStr(az, state.HighPrecision) + "#"

		case ":GR":
			// telescope RA based on precision config
			ra, err := t.GetRightAscension()
			if err != nil {
				log.Errorf("Unable to get telescope right ascension (:GR#): %s", err.Error())
				ra = 0.0
			}
			ret = DegreesToStr(ra, state.HighPrecision) + "#"

		case ":Gd":
			// get target declination
			dec, err := t.GetTargetDeclination()
			if err != nil {
				log.Errorf("Unable to get target declination (:Gd#): %s", err.Error())
				dec = 0.0
			}
			ret = DegreesToStr(dec, state.HighPrecision) + "#"

		case ":GG":
			// get UTC offset time

		case ":Gg":
			// Get site longitude
			long, err := t.GetSiteLongitude()
			if err != nil {
				log.Errorf("Unable to get site longitude (:Gg#): %s", err.Error())
				long = 0.0
			}
			ret = DegreesToLong(long) + "#"

		case ":Gt":
			// Get site latitude
			lat, err := t.GetSiteLatitude()
			if err != nil {
				log.Errorf("Unable to get site latitude (:Gt#): %s", err.Error())
				lat = 0.0
			}
			ret = DegreesToLat(lat) + "#"

		case ":H#":
			// switch between 12/24hr clock mode
			state.TwentyFourHour = !state.TwentyFourHour
			// returns nothing

		case ":MA":
			// slew to target alt/az not supported
			ret = "1" // fault

		case ":P#":
			// toggle high precision
			state.HighPrecision = !state.HighPrecision
			if state.HighPrecision {
				ret = "HIGH PRECISION"
			} else {
				ret = "LOW PRECISION"
			}

		/* Move Telescope Manually */
		case ":Me":
			// slew east (+ long)
			axis := alpaca.AxisAzmRa
			rate := state.rateToASCOM(false)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing

		case ":Mw":
			// slew west (- long)
			axis := alpaca.AxisAzmRa
			rate := state.rateToASCOM(true)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing
			//
		case ":Mn":
			// slew north (+ long)
			axis := alpaca.AxisAltDec
			rate := state.rateToASCOM(true)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing

		case ":Ms":
			// slew south (-lat)
			axis := alpaca.AxisAltDec
			rate := state.rateToASCOM(false)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing

		case ":MS":
			// slew to target
			if state.AutoTrack {
				// auto-enable tracking?
				mode, err := t.GetTracking()
				if err != nil {
					log.Errorf("Unable to get tracking mode: %s", err.Error())
				} else {
					if mode == alpaca.NotTracking {
						err = t.PutTracking(alpaca.AltAz) // need any non-NotTracking value for true
						if err != nil {
							log.Errorf("Unable to auto-enable tracking: %s", err.Error())
						}
					}
				}
			}
			err = t.PutSlewToTargetAsync()
			// we don't get any good/bad answer from Alpaca, so always say success
			ret = "0"

		case ":Q#":
			// halt slewing
			_ = t.PutAbortSlew()
			// returns nothing

		case ":Qe", ":Qw":
			// halt slew in E/W
			axis := alpaca.AxisAzmRa
			err = t.PutMoveAxis(axis, 0)
			// returns nothing

		case ":Qn", ":Qs":
			// halt slew in N/S
			axis := alpaca.AxisAltDec
			err = t.PutMoveAxis(axis, 0)
			// returns nothing

		case ":RG":
			// slew to slowest
			state.SlewRate = 1
			// returns nothing

		case ":RC":
			// slew to 2nd slowest
			state.SlewRate = 2
			// returns nothing

		case ":RM":
			// slew to 2nd fastest
			state.SlewRate = int(state.MaxSlew) - 1
			// returns nothing

		case ":RS":
			// slew at max rate
			state.SlewRate = int(state.MaxSlew)
			// returns nothing

		case ":Sd":
			// Set target Declination
			var degrees, min, sec int
			var sign byte
			var dms DMS
			switch strings.Count(cmd, ":") {
			case 2:
				// sDD:MM:SS
				_, err = fmt.Sscanf(cmd, ":Sd%c%02d*%02d:%02d#", &sign, &degrees, &min, &sec)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
					ret = "0"
				}
				if sign == 0x45 { // look for '-'
					degrees *= -1
				}
				dms = NewDMS(degrees, min, float64(sec))
			case 1:
				// sDD:MM
				_, err = fmt.Sscanf(cmd, ":Sd%c%02d*%02d#", &sign, &degrees, &min)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
					ret = "0"
				}
				if sign == 0x45 { // look for '-'
					degrees *= -1
				}
				dms = NewDMSShort(degrees, float64(min))
			default:
				log.Errorf("Unable to parse %s", cmd)
				ret = "0"
			}
			if ret == "" {
				err = t.PutTargetDeclination(dms.Float)
				if err != nil {
					ret = "0"
				} else {
					ret = "1"
				}
			}

		case ":Sg":
			// Set site longitude: :SgDDD*MM#
			var deg, min int
			_, err = fmt.Sscanf(cmd, ":Sg%03d*%02d#", &deg, &min)
			if err != nil {
				log.Errorf("Error parsing '%s': %s", cmd, err.Error())
				ret = "0"
			} else {
				dms := NewDMS(deg, min, 0)
				err = t.PutSiteLongitude(dms.FloatPositive)
				if err != nil {
					ret = "0"
				} else {
					ret = "1"
				}
			}

		/*
		 * LX200 uses 3 different commands to set date and time where Alpaca uses
		 * only one.  This means we have to assume that SkySafari/etc will send
		 * all three.  Pretty sure :SC must be last, but who knows?
		 *
		 * If SkySafari uses "LX200 Classic" then setting date/time will cause
		 * SkySafari to hang for about 30sec until it times out.  LX200 GPS
		 * does not seem to have this issue.
		 */
		case ":SG":
			// set TZ offset hours: :SGsHH.H
			var sign byte
			var hrsFloat float64
			var hrsInt int

			ret = "1"

			// this is what the docs say
			_, err = fmt.Sscanf(cmd, ":SG%c%2.1f#", &sign, &hrsFloat)
			if err != nil {
				// and this is what SkySafari actually sends :(
				_, err = fmt.Sscanf(cmd, ":SG%c%02d#", &sign, &hrsInt)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
					ret = "0"
				} else {
					hrsFloat = float64(hrsInt)
				}
			}
			if err == nil {
				if sign == '-' {
					hrsFloat *= -1
				}
				state.UTCOffset = hrsFloat
				err = state.SendDateTime(t)
			}

		case ":SC":
			// set local date: :SCMM/DD/YY#
			ret = "1Updating Planetary Data#" // LOL, WTF is this???
			_, err = fmt.Sscanf(cmd, ":SC%02d/%02d/%02d#", &state.month, &state.day, &state.year)
			if err != nil {
				err = fmt.Errorf("unable to parse time '%s': %s", cmd, err.Error())
				ret = "0"
			} else {
				state.haveDate = true
				state.year += 2000
				err = state.SendDateTime(t)
			}

		case ":SL":
			// set local time: :SLHH:MM:SS in 24hr format
			ret = "1"
			_, err = fmt.Sscanf(cmd, ":SL%02d:%02d:%02d#", &state.hour, &state.minute, &state.second)
			if err != nil {
				err = fmt.Errorf("unable to parse time '%s': %s", cmd, err.Error())
				ret = "0"
			} else {
				state.haveTime = true
				err = state.SendDateTime(t)
			}

		case ":Sr":
			// Set target RA
			var hour, min, sec int
			var hms HMS
			switch strings.Count(cmd, ":") {
			case 3:
				// HH:MM:SS
				_, err = fmt.Sscanf(cmd, ":Sr%02d:%02d:%02d#", &hour, &min, &sec)
				if err != nil {
					log.Errorf("error parsing '%s': %s", cmd, err.Error())
					ret = "0"
				}
				hms = NewHMS(hour, min, float64(sec))
			case 2:
				// HH:MM.T  not sure what T is.  Assuming is tenth of sec?
				_, err = fmt.Sscanf(cmd, ":Sr%02d:%02d.%d#", &hour, &min, &sec)
				if err != nil {
					log.Errorf("error parsing '%s': %s", cmd, err.Error())
					ret = "0"
				}
				minFloat := float64(min) + (float64(sec) / 10.0)
				hms = NewHMSShort(hour, minFloat)
			default:
				log.Errorf("unable to parse %s", cmd)
				ret = "0"
			}

			if ret == "" {
				err = t.PutTargetRightAscension(hms.Float)
				if err != nil {
					ret = "0"
				} else {
					ret = "1"
				}
			}

		case ":St":
			// Set site latitude: :StsDD*MM#
			var sign byte
			var deg, min int
			_, err = fmt.Sscanf(cmd, ":St%c%02d*%02d#", &sign, &deg, &min)
			if err != nil {
				log.Errorf("error parsing '%s': %s", cmd, err.Error())
				ret = "0"
			} else {
				if sign == '-' {
					deg *= -1
				}
				dms := NewDMS(deg, min, 0)
				err = t.PutSiteLatitude(dms.Float)
				if err != nil {
					ret = "0"
				} else {
					ret = "1"
				}
			}

		default:
			log.Errorf("unsupported command: '%s'", cmd)
		}
	}

	if err != nil {
		log.Errorf("error talking to scope: %s", err.Error())
	}

	// convert our return string to the ret_val
	if ret != "" {
		retVal = []byte(ret)
	}
	log.Debugf("sending ret_val = %v, %d bytes consumed", retVal, consumed)
	return retVal, consumed
}

func (state *LX200) rateToASCOM(movePostion bool) int {
	ret := state.SlewRate
	if !movePostion {
		ret *= -1
	}
	return ret
}

// Converts float to sDD*MM or sDD*MM'SS
func DegreesToStr(deg float64, highp bool) string {
	sign := '+'
	if deg < 0.0 {
		sign = '-'
	}
	dd := int(deg)
	remain := deg - math.Floor(deg)
	mm := int(remain * 60.0)
	fracMinute := remain - float64(mm)/60.0
	ss := int(fracMinute * 60.0 * 60.0)
	if highp {
		return fmt.Sprintf("%c%02d*%02d'%02d", sign, dd, mm, ss)
	} else {
		return fmt.Sprintf("%c%02d*%02d", sign, dd, mm)
	}
}

// Converts float to sDDD*MM for Long.
func DegreesToLong(deg float64) string {
	sign := '+'
	if deg < 0.0 {
		sign = '-'
	}
	dd := int(deg)
	remain := deg - math.Floor(deg)
	mm := int(remain * 60.0)
	return fmt.Sprintf("%c%03d*%02d", sign, dd, mm)
}

// Converts float to sDD*MM for Latitude.
func DegreesToLat(deg float64) string {
	sign := '+'
	if deg < 0.0 {
		sign = '-'
	}
	dd := int(deg)
	remain := deg - math.Floor(deg)
	mm := int(remain * 60.0)
	return fmt.Sprintf("%c%02d*%02d", sign, dd, mm)
}

/*
 * Function called by :SC, :SL and SG to see if we can
 * send the current time to Alpaca
 */
func (state *LX200) SendDateTime(t *alpaca.Telescope) error {
	if state.UTCOffset > 24.0 || !state.haveTime || !state.haveDate {
		log.Debugf("Skipping SendDateTime()")
		return nil // nothing to do
	}

	location, _ := time.LoadLocation("UTC")
	date := time.Date(state.year, time.Month(state.month), state.day,
		state.hour, state.minute, state.second, 0, location)
	date = date.Add(time.Hour * time.Duration(state.UTCOffset))
	log.Debugf("calling PutUTCDate: %v", date)
	return t.PutUTCDate(date)
}

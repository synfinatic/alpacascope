package main

import (
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/synfinatic/alpacascope/alpaca"
	"github.com/synfinatic/alpacascope/telescope"
)

type LX200State struct {
	HighPrecision  bool
	TwentyFourHour bool // :H#
	MaxSlew        float64
	MinSlew        float64
	SlewRate       int
}

func handleLX200Conn(conn net.Conn, t *alpaca.Telescope, state *LX200State) {
	buf := make([]byte, 1024)

	defer conn.Close()
	rlen, err := conn.Read(buf)
	for {
		if err != nil {
			break
		}
		/*
		 * LX200 has a single byte command <0x06> and variable length commands
		 * which start with a ':' and end with a '#'.  Also, clients may send
		 * multiple commands at once :-/
		 */
		for rlen > 0 {
			reply, consumed := lx200_command(t, rlen, buf, state)
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

func lx200_command(t *alpaca.Telescope, cmdlen int, buf []byte, state *LX200State) ([]byte, int) {
	var consumed int = 0
	var ret_val []byte
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
		end_of_command := strings.Index(commands, "#")
		consumed = end_of_command + 1
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
		 * :Gb - Browse brigher magintude limit
		 * :GF - field diameter
		 * :GF - faint magintude limit
		 * :Gh - get high limit
		 * :Gl - Larger size limit
		 * :GM, :GN, :GO, :GP - get site (1, 2, 3, 4) name
		 * :Go - Get lower limit
		 * :Gq - Get minimum quality for find operation
		 * :GVD - get firmware date
		 * :GVN - Get firmware version
		 * :GVP - Get product name
		 * :GVT - Get firmware time
		 * :Gy# - Get deepsky object string
		 *
		 * :h - home position commands
		 * :I - initialize scope
		 * :L - object library commands
		 * :MS - slew to target // ??
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
				log.Errorf("Unable to tget telescope azimuth (:GZ#): %s", err.Error())
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
			rate := lx200_rate_to_ascom(state, false)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing

		case ":Mw":
			// slew west (- long)
			axis := alpaca.AxisAzmRa
			rate := lx200_rate_to_ascom(state, true)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing
			//
		case ":Mn":
			// slew north (+ long)
			axis := alpaca.AxisAltDec
			rate := lx200_rate_to_ascom(state, true)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing

		case ":Ms":
			// slew south (-lat)
			axis := alpaca.AxisAltDec
			rate := lx200_rate_to_ascom(state, false)
			err = t.PutMoveAxis(axis, rate)
			// returns nothing

		case ":MS":
			// slew to target
			err = t.PutSlewToTargetAsync()
			// we don't get any good/bad answer from Alpaca, so always say success
			ret = "0"

		case ":Q#":
			// halt slewing
			t.PutAbortSlew()
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
			var err error
			var dms telescope.DMS
			if state.HighPrecision {
				// sDD:MM:SS
				_, err = fmt.Sscanf(cmd, ":Sd%c%02d*%02d:%02d#", &sign, &degrees, &min, &sec)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
				}
				if sign == 0x45 { // look for '-'
					degrees *= -1
				}
				dms = telescope.NewDMS(degrees, min, float64(sec))
			} else {
				// sDD:MM
				_, err = fmt.Sscanf(cmd, ":Sd%c%02d*%02d#", &sign, &degrees, &min)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
				}
				if sign == 0x45 { // look for '-'
					degrees *= -1
				}
				dms = telescope.NewDMSShort(degrees, float64(min))
			}
			err = t.PutTargetDeclination(dms.Float)
			if err != nil {
				ret = "0"
			} else {
				ret = "1"
			}

		case ":Sr":
			// Set target RA
			var hour, min, sec int
			var err error
			var hms telescope.HMS
			if state.HighPrecision {
				// HH:MM:SS
				_, err = fmt.Sscanf(cmd, ":Sr%02d:%02d:%02d#", &hour, &min, &sec)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
				}
				hms = telescope.NewHMS(hour, min, float64(sec))
			} else {
				// HH:MM.T  not sure what T is.  Assuming is tenth of sec?
				_, err = fmt.Sscanf(cmd, ":Sr%02d:%02d.%d#", &hour, &min, &sec)
				if err != nil {
					log.Errorf("Error parsing '%s': %s", cmd, err.Error())
				}
				min_float := float64(min) + (float64(sec) / 10.0)
				hms = telescope.NewHMSShort(hour, min_float)
			}
			err = t.PutTargetRightAscension(hms.Float)
			if err != nil {
				ret = "0"
			} else {
				ret = "1"
			}

		default:
			log.Errorf("Unsupported command: '%s'", cmd)
		}
	}

	if err != nil {
		log.Errorf("Error talking to scope: %s", err.Error())
	}

	// convert our return string to the ret_val
	if ret != "" {
		ret_val = []byte(ret)
	}
	return ret_val, consumed
}

func lx200_rate_to_ascom(state *LX200State, move_positive bool) int {
	ret := state.SlewRate
	if !move_positive {
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
	frac_minute := remain - float64(mm)/60.0
	ss := int(frac_minute * 60.0 * 60.0)
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

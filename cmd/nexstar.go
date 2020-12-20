package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	alpaca "github.com/synfinatic/alpaca-gateway/alpaca"

	log "github.com/sirupsen/logrus"
)

func handleNexstar(conn net.Conn, t *alpaca.Telescope) {
	var we error = nil
	buf := make([]byte, 1024)
	len, err := conn.Read(buf)
	if err != nil {
		log.Errorf("reading from NexStar client: %s", err.Error())
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		var strbuf string
		for i := 1; i < len; i++ {
			strbuf = fmt.Sprintf("%s %d", strbuf, buf[i])
		}
		log.Debugf("Received %d bytes [%s]: %c %s", len, string(buf), buf[0], strbuf)
	}

	// single byte commands
	switch buf[0] {
	case 'E':
		ra, dec, err := t.GetRaDec()
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
			return
		}

		var ra_int uint32 = uint32(ra / 360.0 * 65536.0)
		var dec_int uint32 = uint32(dec / 360.0 * 65536.0)
		var ret string = fmt.Sprintf("%04X,%04X#", ra_int, dec_int)
		_, we = conn.Write([]byte(ret))

	case 'e':
		// Get Precise RA/DEC
		ra, dec, err := t.GetRaDec()
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
			return
		}

		var ra_int uint32 = uint32(ra / 360.0 * 4294967296.0)
		var dec_int uint32 = uint32(dec / 360.0 * 4294967296.0)
		var ret string = fmt.Sprintf("%08X,%08X#", ra_int, dec_int)
		_, we = conn.Write([]byte(ret))

	case 'Z':
		// Get AZM/ALT.  Note that AZM is 0->360, while Alt is -90->90
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
			return
		}

		var azm_int uint32 = uint32(azm / 360.0 * 65536.0)
		var alt_int uint32 = uint32(alt / 360.0 * 65536.0)
		var ret string = fmt.Sprintf("%04X,%04X#", azm_int, alt_int)
		_, we = conn.Write([]byte(ret))

	case 'z':
		// Get Precise AZM/ALT
		azm, alt, err := t.GetAzmAlt()
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
			return
		}

		var azm_int uint32 = uint32(azm / 360.0 * 4294967296.0)
		var alt_int uint32 = uint32(alt / 360.0 * 4294967296.0)
		var ret string = fmt.Sprintf("%08X,%08X#", azm_int, alt_int)
		_, we = conn.Write([]byte(ret))

	case 't':
		// Get tracking mode
		// 0 = off, 1 alt/az, 2 EQ North, 3 EQ south
		_, we = conn.Write([]byte("1#")) // hard code to alt/az for now

	case 'T':
		// Set tracking mode
		_, we = conn.Write([]byte("#")) // just say ok

	case 'V':
		// Get Version
		_, we = conn.Write([]byte("50#"))

	case 'P':
		// Slew
		err = executeSlew(t, buf)
		if err != nil {
			log.Errorf("Unable to slew: %s", err.Error())
			return
		}
		_, we = conn.Write([]byte("#"))

	case 's':
		// Precise sync aka: Align on object.  Uses the same math as 'e'
		ra_bytes := buf[1:8]
		dec_bytes := buf[10:18]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 4294967296.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 4294967296.0 * 360.0
		err = executeSync(t, ra, dec)

		_, we = conn.Write([]byte("#"))

	case 'S':
		// sync aka: Align on object.  Uses same math as 'E'
		ra_bytes := buf[1:4]
		dec_bytes := buf[6:10]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 65536.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 65536.0 * 360.0
		err = executeSync(t, ra, dec)
		_, we = conn.Write([]byte("#"))

	case 'r':
		// precise goto Ra/Dec values
		ra_bytes := buf[1:8]
		dec_bytes := buf[10:18]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 4294967296.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 4294967296.0 * 360.0
		err = executeSlewToCoordinatesAsync(t, ra, dec)
		_, we = conn.Write([]byte("#"))

	case 'R':
		// precise goto Ra/Dec values
		ra_bytes := buf[1:4]
		dec_bytes := buf[6:10]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 65536.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 65536.0 * 360.0
		err = executeSlewToCoordinatesAsync(t, ra, dec)
		_, we = conn.Write([]byte("#"))

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
		_, we = conn.Write([]byte("#"))

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
		_, we = conn.Write([]byte("#"))
	default:
		log.Errorf("Unsupported command: %c", buf[0])
	}

	if err != nil {
		log.Errorf("Error talking to scope: %s", err.Error())
	}

	if we != nil {
		log.Errorf("Error writing to client: %s", we.Error())
	}
}

/*
 * Aligns/Sync a scope to the Ra/Dec
 */
func executeSync(t *alpaca.Telescope, ra float64, dec float64) error {
	err := t.PutSyncToCoordinates(ra, dec)
	return err
}

/*
 * Tells the scope to GoTo a Ra/Dec
 */
func executeSlewToCoordinatesAsync(t *alpaca.Telescope, ra float64, dec float64) error {
	err := t.PutSlewToCoordinatestAsync(ra, dec)
	return err
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

	switch buf[2] {
	case 16:
		// Azm/RA
		axis = alpaca.AxisAzmRa
	case 17:
		// Alt/Dec
		axis = alpaca.AxisAltDec
	}

	switch buf[3] {
	case 6, 36:
		positive_direction = true
	case 7, 37:
		positive_direction = false
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

package main

import (
	"encoding/binary"
	"fmt"
	"net"

	alpaca "github.com/synfinatic/alpaca-gateway/alpaca"

	log "github.com/sirupsen/logrus"
)

// Get RA/DEC as degrees (double)
func ra_dec(t *alpaca.Telescope) (float64, float64, error) {
	ra, err := t.GetRightAscension()
	if err != nil {
		return 0.0, 0.0, err
	} else {
		log.Debugf("RA: %g", ra)
	}
	dec, err := t.GetDeclination()
	if err != nil {
		return 0.0, 0.0, err
	} else {
		log.Debugf("Dec: %g", dec)
	}

	return ra, dec, nil
}

// Get Azmiuth / Altitude as degrees (double)
func azm_alt(t *alpaca.Telescope) (float64, float64, error) {
	azm, err := t.GetAzimuth()
	if err != nil {
		return 0.0, 0.0, err
	} else {
		log.Debugf("Azm: %g", azm)
	}
	alt, err := t.GetAltitude()
	if err != nil {
		return 0.0, 0.0, err
	} else {
		log.Debugf("Alt: %g", alt)
	}

	return azm, alt, nil

}

func handleNexStar(conn net.Conn, t *alpaca.Telescope) {
	buf := make([]byte, 1024)
	len, err := conn.Read(buf)
	if err != nil {
		log.Errorf("reading from NexStar client: %s", err.Error())
	}

	// single byte commands
	switch buf[0] {
	case 'E':
		ra, dec, err := ra_dec(t)
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
			return
		}

		var ra_int uint32 = uint32(ra / 360.0 * 65536.0)
		var dec_int uint32 = uint32(dec / 360.0 * 65536.0)
		var ret string = fmt.Sprintf("%04X,%04X#", ra_int, dec_int)
		_, err = conn.Write([]byte(ret))
		if err != nil {
			log.Fatalf("Error replying to client: %s", err.Error())
		}

	case 'e':
		// Get Precise RA/DEC
		ra, dec, err := ra_dec(t)
		if err != nil {
			log.Errorf("Unable to get RA/DEC: %s", err.Error())
			return
		}

		var ra_int uint32 = uint32(ra / 360.0 * 4294967296.0)
		var dec_int uint32 = uint32(dec / 360.0 * 4294967296.0)
		var ret string = fmt.Sprintf("%08X,%08X#", ra_int, dec_int)
		_, err = conn.Write([]byte(ret))
		if err != nil {
			log.Fatalf("Error replying to client: %s", err.Error())
		}

	case 'Z':
		// Get AZM/ALT.  Note that AZM is 0->360, while Alt is -90->90
		azm, alt, err := azm_alt(t)
		if err != nil {
			log.Errorf("Unable to get AZM/ALT: %s", err.Error())
			return
		}

		var azm_int uint32 = uint32(azm / 360.0 * 65536.0)
		var alt_int uint32 = uint32(alt / 360.0 * 65536.0)
		var ret string = fmt.Sprintf("%04X,%04X#", azm_int, alt_int)
		_, err = conn.Write([]byte(ret))
		if err != nil {
			log.Fatalf("Error replying to client: %s", err.Error())
		}

	case 'z':
		// Get Precise AZM/ALT
		log.Debugf("Requested 'z'")
		return

	case 't':
		// Get tracking mode
		// 0 = off, 1 alt/az, 2 EQ North, 3 EQ south
		_, err = conn.Write([]byte("1#")) // hard code to alt/az for now
		return

	case 'T':
		// Set tracking mode
		_, err = conn.Write([]byte("#")) // just say ok
		return

	case 'V':
		// Get Version
		_, err = conn.Write([]byte("50#"))
		return

	case 'P':
		// Slew
		err = executeSlew(t, buf)
		if err != nil {
			log.Errorf("Unable to slew: %s", err.Error())
			return
		}
		_, err = conn.Write([]byte("#"))

	case 's':
		// Precise sync aka: Align on object.  Uses the same math as 'e'
		ra_bytes := buf[1:8]
		dec_bytes := buf[10:18]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 4294967296.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 4294967296.0 * 360.0
		err = executeSync(t, ra, dec)

		_, err = conn.Write([]byte("#"))

	case 'S':
		// sync aka: Align on object.  Uses same math as 'E'
		ra_bytes := buf[1:4]
		dec_bytes := buf[6:10]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 65536.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 65536.0 * 360.0
		err = executeSync(t, ra, dec)

		_, err = conn.Write([]byte("#"))

	case 'r':
		// precise goto Ra/Dec values
		ra_bytes := buf[1:8]
		dec_bytes := buf[10:18]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 4294967296.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 4294967296.0 * 360.0
		err = executeSlewToCoordinatesAsync(t, ra, dec)

		_, err = conn.Write([]byte("#"))

	case 'R':
		// precise goto Ra/Dec values
		ra_bytes := buf[1:4]
		dec_bytes := buf[6:10]
		ra := float64(binary.BigEndian.Uint32(ra_bytes)) / 65536.0 * 360.0
		dec := float64(binary.BigEndian.Uint32(dec_bytes)) / 65536.0 * 360.0
		err = executeSlewToCoordinatesAsync(t, ra, dec)

		_, err = conn.Write([]byte("#"))
	}
	var strbuf string
	for i := 1; i < len; i++ {
		strbuf = fmt.Sprintf("%s %d", strbuf, buf[i])
	}
	log.Debugf("Received %d bytes [%s]: %c %s", len, string(buf), buf[0], strbuf)
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

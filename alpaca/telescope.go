package alpaca

/*
 * Impliments the Alpaca Telescope client API
 */

import (
	"fmt"
	"time"
)

type AlignmentMode int

const (
	algAltAz       int = 0
	algPolar       int = 1
	algGermanPolar int = 2
)

type AxisType int

const (
	AxisAzmRa    = 0
	AxisAltDec   = 1
	AxisTertiary = 2
)

type Telescope struct {
	alpaca *Alpaca
	Id     uint32
}

func NewTelescope(id uint32, alpaca *Alpaca) *Telescope {
	t := Telescope{
		alpaca: alpaca,
		Id:     id,
	}
	return &t
}

func (t *Telescope) GetName() (string, error) {
	return t.alpaca.GetName("telescope", t.Id)
}

func (t *Telescope) GetDescription() (string, error) {
	return t.alpaca.GetDescription("telescope", t.Id)
}

func (t *Telescope) GetConnected() (bool, error) {
	return t.alpaca.GetConnected("telescope", t.Id)
}

func (t *Telescope) GetSupportedActions() ([]string, error) {
	return t.alpaca.GetSupportedActions("telescope", t.Id)
}

func (t *Telescope) GetAlignmentMode() (int32, error) {
	return t.alpaca.GetInt32("telescope", t.Id, "alignmentmode")
}

func (t *Telescope) GetAltitude() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "altitude")
}

func (t *Telescope) GetAzimuth() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "azimuth")
}

func (t *Telescope) GetDeclination() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "declination")
}

func (t *Telescope) GetRightAscension() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "rightascension")
}

func (t *Telescope) GetCanPark() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "canpark")
}

func (t *Telescope) GetCanFindHome() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "canfindhome")
}

func (t *Telescope) GetCanSlew() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "canslew")
}

func (t *Telescope) GetCanSlewAltAz() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "canslewaltaz")
}

func (t *Telescope) GetCanSlewAsync() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "canslewasync")
}

func (t *Telescope) GetCanSlewAltAzAsync() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "canslewaltazasync")
}

// Returns the min & max rate (deg/sec) that the given axis can move
func (t *Telescope) GetAxisRates(axis AxisType) ([]uint32, error) {
	return t.alpaca.GetListUint32("telescope", t.Id, "axisrates")
}

type putMoveAxis struct {
	Axis                AxisType `json:"Axis"`
	Rate                int      `json:"Rate"`
	ClientID            uint32   `json:"ClientID"`
	ClientTransactionID uint32   `json:"ClientTransactionID"`
}

func (t *Telescope) PutMoveAxis(axis AxisType, rate int) error {
	var form map[string]string = map[string]string{
		"Axis":                fmt.Sprintf("%d", axis),
		"Rate":                fmt.Sprintf("%d", rate),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "moveaxis", form)
	return err
}

func (t *Telescope) PutSyncToCoordinates(ra float64, dec float64) error {
	var form map[string]string = map[string]string{
		"RightAscension":      fmt.Sprintf("%g", ra),
		"Declination":         fmt.Sprintf("%g", dec),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "synctocoordinates", form)
	return err
}

func (t *Telescope) PutSlewToCoordinatestAsync(ra float64, dec float64) error {
	var form map[string]string = map[string]string{
		"RightAscension":      fmt.Sprintf("%g", ra),
		"Declination":         fmt.Sprintf("%g", dec),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "slewtocoordinatesasync", form)
	return err
}

func (t *Telescope) PutSlewToCoordinates(ra float64, dec float64) error {
	var form map[string]string = map[string]string{
		"RightAscension":      fmt.Sprintf("%g", ra),
		"Declination":         fmt.Sprintf("%g", dec),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "slewtocoordinates", form)
	return err
}

func (t *Telescope) PutSiteLatitude(lat float64) error {
	var form map[string]string = map[string]string{
		"SiteLatitude":        fmt.Sprintf("%g", lat),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "sitelatitude", form)
	return err
}

func (t *Telescope) PutSiteLongitude(long float64) error {
	var form map[string]string = map[string]string{
		"SiteLongitude":       fmt.Sprintf("%g", long),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "sitelongitude", form)
	return err
}

func (t *Telescope) PutUTCDate(date time.Time) error {
	var form map[string]string = map[string]string{
		"UTCDate":             fmt.Sprintf("%s", date.Format(time.RFC3339)),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "utcdate", form)
	return err
}

func (t *Telescope) PutAbortSlew() error {
	var form map[string]string = map[string]string{
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "abortslew", form)
	return err
}

/*
 * Helper functions
 */

// Get RA/DEC as hours/degrees (double)
func (t *Telescope) GetRaDec() (float64, float64, error) {
	ra, err := t.GetRightAscension()
	if err != nil {
		return 0.0, 0.0, err
	}
	dec, err := t.GetDeclination()
	if err != nil {
		return 0.0, 0.0, err
	}

	return ra, dec, nil
}

// Get Azmiuth / Altitude as degrees (double)
func (t *Telescope) GetAzmAlt() (float64, float64, error) {
	azm, err := t.GetAzimuth()
	if err != nil {
		return 0.0, 0.0, err
	}
	alt, err := t.GetAltitude()
	if err != nil {
		return 0.0, 0.0, err
	}

	return azm, alt, nil
}

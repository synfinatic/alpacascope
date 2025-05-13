package alpaca

/*
 * Implements the Alpaca Telescope client API
 */

import (
	"fmt"
	"time"

	"github.com/relvacode/iso8601"
	log "github.com/sirupsen/logrus"
)

type TrackingMode int

const (
	NotTracking TrackingMode = iota
	AltAz
	EQNorth
	EQSouth
)

type AlignmentMode int32

const (
	AlignmentAltAz AlignmentMode = iota
	AlignmentPolar
	AlignmentGermanPolar
)

type AxisType int

const (
	AxisAzmRa AxisType = iota
	AxisAltDec
	AxisTertiary
)

type Telescope struct {
	alpaca   *Alpaca
	Id       uint32
	Tracking TrackingMode
}

func NewTelescope(id uint32, tm TrackingMode, alpaca *Alpaca) *Telescope {
	t := Telescope{
		alpaca:   alpaca,
		Id:       id,
		Tracking: tm,
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

func (t *Telescope) GetAlignmentMode() (AlignmentMode, error) {
	mode, err := t.alpaca.GetInt32("telescope", t.Id, "alignmentmode")
	return AlignmentMode(mode), err
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

func (t *Telescope) GetSlewing() (bool, error) {
	return t.alpaca.GetBool("telescope", t.Id, "slewing")
}

func (t *Telescope) GetSiteLatitude() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "sitelatitude")
}

func (t *Telescope) GetSiteLongitude() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "sitelongitude")
}

func (t *Telescope) GetTargetDeclination() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "targetdeclination")
}

func (t *Telescope) GetTargetAltitude() (float64, error) {
	return t.alpaca.GetFloat64("telescope", t.Id, "targetrightascension")
}

func (t *Telescope) GetTracking() (TrackingMode, error) {
	tracking, err := t.alpaca.GetBool("telescope", t.Id, "tracking")
	if err != nil {
		return NotTracking, err
	}
	if !tracking {
		return NotTracking, nil
	}
	return t.Tracking, nil
}

// Parse ISO8601 w/ fractional seconds
func (t *Telescope) GetUTCDate() (time.Time, error) {
	isoTime, err := t.alpaca.GetString("telescope", t.Id, "utcdate")
	if err != nil {
		return time.Unix(0, 0), err
	} else if isoTime == "" {
		// sometimes we get no error, but we get an empty string?
		return time.Unix(0, 0), fmt.Errorf("got an empty UTCDate string")
	}
	return iso8601.ParseString(isoTime)
}

type mapAxisRates struct {
	Value               []map[string]float64 `json:"Value"`
	ClientTransactionID int32                `json:"ClientTransactionID"`
	ServerTransactionID int32                `json:"ServerTransactionID"`
	ErrorNumber         int32                `json:"ErrorNumber"`
	ErrorMessage        string               `json:"ErrorMessage"`
}

// Returns the `Maximum` & `Minimum` rate (deg/sec) that the given axis can move
func (t *Telescope) GetAxisRates(axis AxisType) (map[string]float64, error) {
	url := t.alpaca.url("telescope", t.Id, "axisrates")
	querystr := fmt.Sprintf("Axis=%d&%s", axis, t.alpaca.getQueryString())
	resp, err := t.alpaca.client.R().
		SetResult(&mapAxisRates{}).
		SetQueryString(querystr).
		Get(url)
	if err != nil {
		return map[string]float64{}, err
	}
	result := (resp.Result().(*mapAxisRates))
	if len(result.Value) == 0 {
		log.Errorf("telescope driver returned an empty list for axisrates")
		// sometime it's an empty list?
		return map[string]float64{
			"Minimum": 0.0,
			"Maximum": 0.0,
		}, nil
	}
	return result.Value[0], nil
}

func (t *Telescope) PutConnected(connected bool) error {
	c := "true"
	if !connected {
		c = "false"
	}
	form := map[string]string{
		"Connected":           c,
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "connected", form)
	return err
}

func (t *Telescope) PutMoveAxis(axis AxisType, rate int) error {
	form := map[string]string{
		"Axis":                fmt.Sprintf("%d", axis),
		"Rate":                fmt.Sprintf("%d", rate),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "moveaxis", form)
	return err
}

func (t *Telescope) PutSyncToCoordinates(ra float64, dec float64) error {
	form := map[string]string{
		"RightAscension":      fmt.Sprintf("%g", ra),
		"Declination":         fmt.Sprintf("%g", dec),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "synctocoordinates", form)
	return err
}

func (t *Telescope) PutSlewToCoordinatestAsync(ra float64, dec float64) error {
	form := map[string]string{
		"RightAscension":      fmt.Sprintf("%g", ra),
		"Declination":         fmt.Sprintf("%g", dec),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "slewtocoordinatesasync", form)
	return err
}

func (t *Telescope) PutSlewToCoordinates(ra float64, dec float64) error {
	form := map[string]string{
		"RightAscension":      fmt.Sprintf("%g", ra),
		"Declination":         fmt.Sprintf("%g", dec),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "slewtocoordinates", form)
	return err
}

func (t *Telescope) PutSiteLatitude(lat float64) error {
	form := map[string]string{
		"SiteLatitude":        fmt.Sprintf("%g", lat),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "sitelatitude", form)
	return err
}

func (t *Telescope) PutSiteLongitude(long float64) error {
	form := map[string]string{
		"SiteLongitude":       fmt.Sprintf("%g", long),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "sitelongitude", form)
	return err
}

func (t *Telescope) PutTargetRightAscension(long float64) error {
	form := map[string]string{
		"TargetRightAscension": fmt.Sprintf("%g", long),
		"ClientID":             fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID":  fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "targetrightascension", form)
	return err
}

func (t *Telescope) PutTargetDeclination(long float64) error {
	form := map[string]string{
		"TargetDeclination":   fmt.Sprintf("%g", long),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "targetdeclination", form)
	return err
}

func (t *Telescope) PutUTCDate(date time.Time) error {
	form := map[string]string{
		"UTCDate":             date.Format(time.RFC3339),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "utcdate", form)
	return err
}

func (t *Telescope) PutAbortSlew() error {
	form := map[string]string{
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "abortslew", form)
	return err
}

func (t *Telescope) PutSlewToTargetAsync() error {
	form := map[string]string{
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "slewtotargetasync", form)
	return err
}

func (t *Telescope) PutSyncToTarget() error {
	form := map[string]string{
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "synctotarget", form)
	return err
}

func (t *Telescope) PutTracking(tracking TrackingMode) error {
	enableTracking := tracking != NotTracking
	form := map[string]string{
		"Tracking":            fmt.Sprintf("%v", enableTracking),
		"ClientID":            fmt.Sprintf("%d", t.alpaca.ClientId),
		"ClientTransactionID": fmt.Sprintf("%d", t.alpaca.GetNextTransactionId()),
	}
	err := t.alpaca.Put("telescope", t.Id, "tracking", form)
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

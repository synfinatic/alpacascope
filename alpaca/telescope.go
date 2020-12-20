package alpaca

/*
 * Impliments the Alpaca Telescope client API
 */

import (
	"fmt"
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

type AlignmentMode int

const (
	algAltAz       int = 0
	algPolar       int = 1
	algGermanPolar int = 2
)

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

type AxisType int

const (
	AxisAzmRa  = 0
	AxisAltDec = 1
)

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
	/*
		article := putMoveAxis{
			Axis:                axis,
			Rate:                rate,
			ClientID:            t.alpaca.ClientId,
			ClientTransactionID: t.alpaca.GetNextTransactionId(),
		}
		log.Debugf(article)
	*/
	t.alpaca.Put("telescope", t.Id, "moveaxis", form)
	return nil
}

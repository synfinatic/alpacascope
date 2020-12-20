package alpaca

/*
 * Main functions implimenting an Alpaca REST client.
 * Covers the generic API calls that all ASCOM devices should support
 * as well as the fundamental API call types
 */

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/go-resty/resty/v2"
)

type Alpaca struct {
	client        *resty.Client
	url_base      string
	ClientId      uint32
	transactionId uint32
}

func NewAlpaca(clientid uint32, ip string, port int32) *Alpaca {
	a := Alpaca{
		client:        resty.New(),
		url_base:      fmt.Sprintf("http://%s:%d", ip, port),
		ClientId:      clientid,
		transactionId: 0,
	}
	return &a
}

// Each Alpaca call should have a monotonically incrementing transactionId
func (a *Alpaca) GetNextTransactionId() uint32 {
	a.transactionId += 1
	return a.transactionId
}

// Generate our QueryString with the default parameters
func (a *Alpaca) getQueryString() string {
	return fmt.Sprintf("ClientID=%d&ClientTransactionID=%d", a.ClientId, a.GetNextTransactionId())
}

func (a *Alpaca) url(device string, id uint32, api string) string {
	return fmt.Sprintf("%s/api/v1/%s/%d/%s", a.url_base, device, id, api)
}

type stringResponse struct {
	Value               string `json:"Value"`
	ClientTransactionID uint32 `json:"ClientTransactionID"`
	ServerTransactionID uint32 `json:"ServerTransactionID"`
	ErrorNumber         int32  `json:"ErrorNumber"`
	ErrorMessage        string `json:"ErrorMessage"`
}

func (a *Alpaca) GetString(device string, id uint32, api string) (string, error) {
	url := a.url(device, id, api)
	resp, err := a.client.R().
		SetResult(&stringResponse{}).
		SetQueryString(a.getQueryString()).
		Get(url)
	if err != nil {
		return "", err
	}
	result := (resp.Result().(*stringResponse))
	return result.Value, nil
}

type stringlistResponse struct {
	Value               []string `json:"Value"`
	ClientTransactionID uint32   `json:"ClientTransactionID"`
	ServerTransactionID uint32   `json:"ServerTransactionID"`
	ErrorNumber         int32    `json:"ErrorNumber"`
	ErrorMessage        string   `json:"ErrorMessage"`
}

func (a *Alpaca) GetStringList(device string, id uint32, api string) ([]string, error) {
	url := a.url(device, id, api)
	resp, err := a.client.R().
		SetResult(&stringlistResponse{}).
		SetQueryString(a.getQueryString()).
		Get(url)
	if err != nil {
		return []string{""}, err
	}
	result := (resp.Result().(*stringlistResponse))
	return result.Value, nil
}

type boolResponse struct {
	Value               bool   `json:"Value"`
	ClientTransactionID uint32 `json:"ClientTransactionID"`
	ServerTransactionID uint32 `json:"ServerTransactionID"`
	ErrorNumber         int32  `json:"ErrorNumber"`
	ErrorMessage        string `json:"ErrorMessage"`
}

func (a *Alpaca) GetBool(device string, id uint32, api string) (bool, error) {
	url := a.url(device, id, api)
	resp, err := a.client.R().
		SetResult(&boolResponse{}).
		SetQueryString(a.getQueryString()).
		Get(url)
	if err != nil {
		return false, err
	}
	result := (resp.Result().(*boolResponse))
	return result.Value, nil
}

type int32Response struct {
	Value               int32  `json:"Value"`
	ClientTransactionID uint32 `json:"ClientTransactionID"`
	ServerTransactionID uint32 `json:"ServerTransactionID"`
	ErrorNumber         int32  `json:"ErrorNumber"`
	ErrorMessage        string `json:"ErrorMessage"`
}

func (a *Alpaca) GetInt32(device string, id uint32, api string) (int32, error) {
	url := a.url(device, id, api)
	resp, err := a.client.R().
		SetResult(&int32Response{}).
		SetQueryString(a.getQueryString()).
		Get(url)
	if err != nil {
		return 0, err
	}
	result := (resp.Result().(*int32Response))
	return result.Value, nil
}

type float64Response struct {
	Value               float64 `json:"Value"`
	ClientTransactionID int32   `json:"ClientTransactionID"`
	ServerTransactionID int32   `json:"ServerTransactionID"`
	ErrorNumber         int32   `json:"ErrorNumber"`
	ErrorMessage        string  `json:"ErrorMessage"`
}

func (a *Alpaca) GetFloat64(device string, id uint32, api string) (float64, error) {
	url := a.url(device, id, api)
	resp, err := a.client.R().
		SetResult(&float64Response{}).
		SetQueryString(a.getQueryString()).
		Get(url)
	if err != nil {
		return 0, err
	}
	result := (resp.Result().(*float64Response))
	return result.Value, nil
}

/*
 * https://ascom-standards.org/api/#/ASCOM%20Methods%20Common%20To%20All%20Devices/get__device_type___device_number__name
 */
func (a *Alpaca) GetName(device string, id uint32) (string, error) {
	return a.GetString(device, id, "name")
}

/*
 * https://ascom-standards.org/api/#/ASCOM%20Methods%20Common%20To%20All%20Devices/get__device_type___device_number__description
 */
func (a *Alpaca) GetDescription(device string, id uint32) (string, error) {
	return a.GetString(device, id, "description")
}

/*
 * https://ascom-standards.org/api/#/ASCOM%20Methods%20Common%20To%20All%20Devices/get__device_type___device_number__connected
 */
func (a *Alpaca) GetConnected(device string, id uint32) (bool, error) {
	return a.GetBool(device, id, "connected")
}

/*
 * https://ascom-standards.org/api/#/ASCOM%20Methods%20Common%20To%20All%20Devices/get__device_type___device_number__supportedactions
 */
func (a *Alpaca) GetSupportedActions(device string, id uint32) ([]string, error) {
	return a.GetStringList(device, id, "supportedactions")
}

type putResponse struct {
	ClientTransactionID uint32 `json:"ClientTransactionID"`
	ServerTransactionID uint32 `json:"ServerTransactionID"`
	ErrorNumber         int32  `json:"ErrorNumber"`
	ErrorMessage        string `json:"ErrorMessage"`
}

func (a *Alpaca) Put(device string, id uint32, api string, form map[string]string) error {
	url := a.url(device, id, api)
	resp, err := a.client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetResult(&putResponse{}).
		SetFormData(form).
		Put(url)
	if err != nil {
		return err
	}

	/*
		// Explore response object
		fmt.Println("Response Info:")
		fmt.Println("  Error      :", err)
		fmt.Println("  Status Code:", resp.StatusCode())
		fmt.Println("  Status     :", resp.Status())
		fmt.Println("  Proto      :", resp.Proto())
		fmt.Println("  Time       :", resp.Time())
		fmt.Println("  Received At:", resp.ReceivedAt())
		fmt.Println("  Body       :\n", resp)
		fmt.Println()
	*/

	result := (resp.Result().(*putResponse))
	if result.ErrorNumber != 0 {
		return fmt.Errorf("%d: %s", result.ErrorNumber, result.ErrorMessage)
	}
	log.Debugf("%v", result)
	return nil
}

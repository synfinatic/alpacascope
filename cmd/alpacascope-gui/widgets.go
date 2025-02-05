package main

import (
	"fmt"

	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"github.com/synfinatic/alpacascope/utils"
)

func NewWidgets(config *AlpacaScopeConfig) *Widgets {
	w := Widgets{}

	// TelescopeMount
	w.TelescopeMount = widget.NewSelect(
		[]string{"Alt-Az", "EQ North", "EQ South"},
		func(val string) {
			config.TelescopeMount = val
		},
	)
	w.TelescopeMount.Selected = config.TelescopeMount

	// Telescope Protocol
	w.TelescopeProtocol = widget.NewSelect([]string{"NexStar", "LX200"},
		func(proto string) {
			config.TelescopeProtocol = proto
			if proto == "NexStar" {
				w.TelescopeMount.Enable()
				w.HighPrecisionLX200.Disable()
			} else {
				// only NexStar supports the mountType
				w.TelescopeMount.Disable()
				// only LX200 supports high precision
				w.HighPrecisionLX200.Enable()
			}
			w.form.Refresh()
		},
	)
	w.TelescopeProtocol.Selected = config.TelescopeProtocol
	if config.TelescopeProtocol == "LX200" {
		w.TelescopeMount.Disable()
	}

	// AutoTracking
	w.AutoTracking = widget.NewCheck("", func(enabled bool) {
		config.AutoTracking = enabled
		w.form.Refresh()
	})
	w.AutoTracking.Checked = config.AutoTracking

	// HighPrecisionLX200
	w.HighPrecisionLX200 = widget.NewCheck("", func(enabled bool) {
		config.HighPrecisionLX200 = enabled
		w.form.Refresh()
	})
	w.HighPrecisionLX200.Checked = config.HighPrecisionLX200
	if config.TelescopeProtocol == "LX200" {
		w.HighPrecisionLX200.Enable()
	} else {
		w.HighPrecisionLX200.Disable()
	}

	// ListenIp
	ips, err := utils.GetLocalIPs()
	if err != nil {
		ips = []string{config.ListenIP}
		sbox.AddLine(fmt.Sprintf("Unable to detect interfaces: %s", err.Error()))
	}
	w.ListenIP = widget.NewSelect(ips, func(ip string) {
		config.ListenIP = ip
	})
	w.ListenIP.Selected = config.ListenIP

	// ListenPort
	w.ListenPort = widget.NewEntry()
	w.ListenPort.SetText(config.ListenPort)
	w.ListenPort.Validator = validation.NewRegexp("^[1-9][0-9]+$",
		"Invalid TCP Port number")
	w.ListenPort.OnChanged = func(val string) {
		config.ListenPort = val
	}

	// AscomIp
	w.AscomIP = widget.NewEntry()
	w.AscomIP.SetText(config.AscomIP)
	w.AscomIP.Validator = validation.NewRegexp("^([0-9]+\\.){3}[0-9]+$",
		"Must be a valid IPv4 address")
	w.AscomIP.OnChanged = func(val string) {
		config.AscomIP = val
	}

	// AscomPort
	w.AscomPort = widget.NewEntry()
	w.AscomPort.SetText(config.AscomPort)
	w.AscomPort.Validator = validation.NewRegexp("^[1-9][0-9]+$",
		"Must be a valid integer > 1")
	w.AscomPort.OnChanged = func(val string) {
		config.AscomPort = val
	}

	// AscomAuto
	w.AscomAuto = widget.NewCheck("", func(enabled bool) {
		switch enabled {
		case true:
			config.AscomAuto = true
			w.AscomIP.Disable()
			w.AscomPort.Disable()
		case false:
			config.AscomAuto = false
			w.AscomIP.Enable()
			w.AscomPort.Enable()
		}
		w.form.Refresh()
	})
	w.AscomAuto.Checked = config.AscomAuto
	if config.AscomAuto {
		w.AscomIP.Disable()
		w.AscomPort.Disable()
	}

	// AscomTelescope
	w.AscomTelescope = widget.NewSelect(
		[]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		func(val string) {
			config.AscomTelescope = val
		},
	)
	w.AscomTelescope.Selected = config.AscomTelescope

	// AutoConnectAttempts
	w.AutoConnectAttempts = widget.NewSelect(
		[]string{"3", "10", "60", "300", "900", "Unlimited"},
		func(val string) {
			config.AutoConnectAttempts = val
		},
	)
	w.AutoConnectAttempts.Selected = config.AutoConnectAttempts

	// AutoStart
	w.AutoStart = widget.NewCheck("", func(enabled bool) {
		config.AutoStart = enabled
	})
	w.AutoStart.Checked = config.AutoStart

	// status field
	w.Status = widget.NewTextGrid()
	w.Status.SetText(STOPPED)
	return &w
}

func (w *Widgets) Enable() {
	w.Save.Enable()
	w.Delete.Enable()
	w.TelescopeProtocol.Enable()
	w.TelescopeMount.Enable()
	w.HighPrecisionLX200.Enable()
	w.AutoTracking.Enable()
	w.ListenIP.Enable()
	w.ListenPort.Enable()
	w.AscomAuto.Enable()
	w.AutoConnectAttempts.Enable()
	w.AscomIP.Enable()
	w.AscomPort.Enable()
	w.AscomTelescope.Enable()
	w.AutoStart.Enable()
}

func (w *Widgets) Disable() {
	w.Save.Disable()
	w.Delete.Disable()
	w.TelescopeProtocol.Disable()
	w.TelescopeMount.Disable()
	w.HighPrecisionLX200.Disable()
	w.AutoTracking.Disable()
	w.ListenIP.Disable()
	w.ListenPort.Disable()
	w.AscomAuto.Disable()
	w.AutoConnectAttempts.Disable()
	w.AscomIP.Disable()
	w.AscomPort.Disable()
	w.AscomTelescope.Disable()
	w.AutoStart.Disable()
}

func (w *Widgets) Set(config *AlpacaScopeConfig) {
	w.TelescopeProtocol.SetSelected(config.TelescopeProtocol)
	w.HighPrecisionLX200.SetChecked(config.HighPrecisionLX200)
	w.TelescopeMount.SetSelected(config.TelescopeMount)
	w.AutoTracking.SetChecked(config.AutoTracking)
	w.ListenIP.SetSelected(config.ListenIP)
	w.ListenPort.SetText(config.ListenPort)
	w.AscomAuto.SetChecked(config.AscomAuto)
	w.AutoConnectAttempts.SetSelected(config.AutoConnectAttempts)
	w.AscomIP.SetText(config.AscomIP)
	w.AscomPort.SetText(config.AscomPort)
	w.AscomTelescope.SetSelected(config.AscomTelescope)
	w.AutoStart.SetChecked(config.AutoStart)
}

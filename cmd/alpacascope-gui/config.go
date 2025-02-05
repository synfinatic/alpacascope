package main

/*
 * AlpacaScope
 * Copyright (c) 2020-2021 Aaron Turner  <synfinatic at gmail dot com>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"strings"

	"github.com/synfinatic/alpacascope/alpaca"
)

// Our actual application config
type AlpacaScopeConfig struct {
	TelescopeProtocol   string `json:"TelescopeProtocol"`
	TelescopeMount      string `json:"TelescopeMount"`
	AutoTracking        bool   `json:"AutoTracking"`
	ListenIP            string `json:"ListenIp"`
	ListenPort          string `json:"ListenPort"`
	AscomAuto           bool   `json:"AscomAuto"`
	AutoConnectAttempts string `json:"AutoConnectAttempts"`
	AutoStart           bool   `json:"AutoStart"`
	AscomIP             string `json:"AscomIp"`
	AscomPort           string `json:"AscomPort"`
	AscomTelescope      string `json:"AscomTelescope"`
	HighPrecisionLX200  bool   `json:"HighPrecisionLX200"`
	isRunning           bool
	Quit                chan bool      `json:"-"` // have to hide since public
	EnableButtons       chan bool      `json:"-"`
	store               *SettingsStore // platform specific
}

// Loads the config from our SettingsStore (if it exists),
// otherwise will return our defaults.  Errors are informational
// so you know why loading settings failed.
func NewAlpacaScopeConfig() *AlpacaScopeConfig {
	config := &AlpacaScopeConfig{
		TelescopeProtocol:   "NexStar",
		TelescopeMount:      "Alt-Az",
		HighPrecisionLX200:  false,
		AutoTracking:        true,
		AscomAuto:           true,
		AutoConnectAttempts: "3",
		AutoStart:           false,
		ListenIP:            "All-Interfaces/0.0.0.0",
		ListenPort:          "4030",
		AscomIP:             "127.0.0.1",
		AscomPort:           alpaca.DEFAULT_PORT_STR,
		AscomTelescope:      "0",
		Quit:                make(chan bool),
		EnableButtons:       make(chan bool),
		store:               NewSettingsStore(),
	}

	return config
}

// Loads our saved settings
func (a *AlpacaScopeConfig) Load() error {
	s := NewSettingsStore()

	// load config.  maybe it worked?  Don't care really....
	err := s.GetSettings(a)
	a.SetStore(s)
	return err
}

// pass through these call to the underlying SettingsStore
func (a *AlpacaScopeConfig) Save() error {
	if a.store != nil {
		return a.store.SaveSettings(a)
	}
	return fmt.Errorf("No valid SettingsStore")
}

func (a *AlpacaScopeConfig) Delete() error {
	if a.store != nil {
		return a.store.Delete()
	}
	return fmt.Errorf("No valid SettingsStore")
}

func (c *AlpacaScopeConfig) ListenIPAddress() string {
	ips := strings.SplitN(c.ListenIP, "/", 2)
	if len(ips) == 2 {
		return ips[1]
	}
	return ips[0]
}

func (c *AlpacaScopeConfig) IsRunning() bool {
	return c.isRunning
}

func (c *AlpacaScopeConfig) SetStore(store *SettingsStore) {
	c.store = store
}

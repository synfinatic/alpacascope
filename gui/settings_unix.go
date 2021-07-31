// +build !windows

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
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	STORE_DIR  = "~/.alpacascope"
	STORE_FILE = "config.json"
)

func getPath(path string) string {
	return strings.Replace(path, "~", os.Getenv("HOME"), 1)
}

type SettingsStore struct {
	path  string
	jdata []byte
}

func NewSettingsStore() (*SettingsStore, error) {
	os.MkdirAll(getPath(STORE_DIR), 0755)
	path := getPath(path.Join(STORE_DIR, STORE_FILE))

	settingBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// missing file
		settingBytes = []byte("")
	}

	return &SettingsStore{
		path:  path,
		jdata: settingBytes,
	}, nil
}

func (ss *SettingsStore) GetSettings(config *AlpacaScopeConfig) error {
	return json.Unmarshal(ss.jdata, config)
}

func (ss *SettingsStore) SaveSettings(config *AlpacaScopeConfig) error {
	jdata, err := json.Marshal(config)
	if err != nil {
		return err
	}
	ss.jdata = jdata
	return ioutil.WriteFile(ss.path, ss.jdata, 0600)
}

func (ss *SettingsStore) Delete() error {
	return os.Remove(ss.path)
}

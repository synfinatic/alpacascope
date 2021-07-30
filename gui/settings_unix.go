// +build darwin linux freebsd netbsd openbsd
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
//	"encoding/json"
//	"fmt"
//	"path/filepath"
)

const (
	STORE_FILE = "~/.alpacascope/config.json"
)

type SettingsStore struct {
	path string
}

func NewSettingsStore() (*SettingsStore, error) {

	return &SettingsStore{}, nil
}

func (ss *SettingsStore) GetSettings(*AlpacaScopeConfig) error {

	return nil
}

func (ss *SettingsStore) SetSettings(*AlpacaScopeConfig) error {

	return nil
}

func (ss *SettingsStore) Delete() error {

	return nil
}

func (ss *SettingsStore) Close() error {

	return nil
}

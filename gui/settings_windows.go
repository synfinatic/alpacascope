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
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	REGISTRY_PATH = `SOFTWARE\AlpacaScope`
	JSON_KEY      = `JSON_CONFIG`
)

type SettingsStore struct {
	jdata string
}

func NewSettingsStore() (*SettingsStore, error) {
	store := &SettingsStore{}
	key, err := store.GetKey()
	if err != nil {
		return store, err
	}
	defer key.Close()

	val, valtype, err := key.GetStringValue(JSON_KEY)
	switch err {
	case registry.ErrNotExist:
		return store, fmt.Errorf("No saved settings")

	case nil:
		break

	default:
		return store, err
	}

	// correct value type?
	if valtype != registry.SZ {
		return store, fmt.Errorf(`Invalid registry type for %s\%s\%s: 0x%04x`,
			"CURRENT_USER", REGISTRY_PATH, JSON_KEY, valtype)
	}

	return &SettingsStore{
		jdata: val,
	}, nil
}

// Loads our settings from the Registry or our defaults.  Returns
// and error if there was a problem loading the Registry
func (ss *SettingsStore) GetSettings(config *AlpacaScopeConfig) error {
	return json.Unmarshal([]byte(ss.jdata), config)
}

func (ss *SettingsStore) SaveSettings(config *AlpacaScopeConfig) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	key, err := ss.GetKey()
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(JSON_KEY, string(bytes))
}

func (ss *SettingsStore) Delete() error {
	key, err := ss.GetKey()
	if err != nil {
		return err
	}
	defer key.Close()
	return key.DeleteValue(JSON_KEY)
}

func (ss *SettingsStore) GetKey() (registry.Key, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, REGISTRY_PATH,
		registry.ALL_ACCESS)
	if err != registry.ErrNotExist {
		key, _, err = registry.CreateKey(registry.CURRENT_USER,
			REGISTRY_PATH, registry.ALL_ACCESS)
		if err != nil {
			return key, fmt.Errorf(`Unable to open registry: %s\%s`,
				"CURRENT_USER", REGISTRY_PATH)
		}
	}

	return key, nil
}

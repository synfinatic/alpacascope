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
	JSON_KEY      = `SOFTWARE\AlpacaScope\Json`
)

type SettingsStore struct {
	key registry.Key
}

func NewSettingsStore() (SettingsStore, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		REGISTRY_PATH, registry.ALL_ACCESS)
	if err != nil {
		key, err = registry.CreateKey(registry.CURRENT_USER,
			REGISTRY_PATH, registry.ALL_ACCESS)
		if err != nil {
			return nil, err
		}
	}

	return SettingsStore{
		key: key,
	}, nil
}

// Loads our settings from the Registry or our defaults.  Returns
// and error if there was a problem loading the Registry
func (ss *SettingsStore) GetSettings(config *AlpacaScopeConfig) error {
	val, valtype, err := ss.key.GetStringValue(JSON_KEY)
	switch err {
	case registry.ErrNotExist:
		return nil

	case nil:
		break

	default:
		return err
	}

	if valtype != registry.SZ {
		return config, fmt.Errorf("Invalid registry type for %s: 0x%04x",
			JSON_KEy, valtype)
	}

	err = json.Unmarshal([]byte(val), config)
	if err != nil {
		return err
	}

	return nil
}

func (ss *SettingsStore) SetSettings(settings *AlpacaScopeConfig) error {
	bytes, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return ss.key.SetStringValue(JSON_KEY, string(bytes))
}

func (ss *SettingsStore) Delete() error {
	return ssn.key.DeleteValue(JSON_KEY)
}

func (ss *SettingsStore) Close() error {
	return ss.key.Close()
}

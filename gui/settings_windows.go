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
	"reflect"

	"golang.org/x/sys/windows/registry"
)

const (
	REGISTRY_PATH = `SOFTWARE\AlpacaScope`
)

type SettingsStore struct{}

func NewSettingsStore() *SettingsStore {
	return &SettingsStore{}
}

// Loads our settings from the Registry or our defaults.  Returns
// and error if there was a problem loading the Registry
func (ss *SettingsStore) GetSettings(config *AlpacaScopeConfig) error {
	key, err := ss.GetKey()
	if err != nil {
		return fmt.Errorf("GetKey: %s", err.Error())
	}
	defer key.Close()

	s := reflect.ValueOf(config).Elem()
	for i := 0; i < s.NumField(); i++ {
		field := s.Type().Field(i).Name
		kind := s.Type().Field(i).Type.Kind()
		val, valtype, err := key.GetStringValue(field)
		if err != nil || valtype != registry.SZ {
			continue // skip errors
		}
		switch kind {
		case reflect.String:
			s.Field(i).SetString(val)

		case reflect.Bool:
			switch val {
			case "true":
				s.Field(i).SetBool(true)
			default:
				s.Field(i).SetBool(false)
			}
		default:
			return fmt.Errorf("Unsupported type for %s: %v", field, kind)
		}
	}

	return nil
}

func (ss *SettingsStore) SaveSettings(config *AlpacaScopeConfig) error {
	key, err := ss.GetKey()
	if err != nil {
		return fmt.Errorf("GetKey: %s", err)
	}
	defer key.Close()

	s := reflect.ValueOf(config).Elem()
	for i := 0; i < s.NumField(); i++ {
		n := s.Type().Field(i).Name
		t := s.Type().Field(i).Type

		tag := string(s.Type().Field(i).Tag.Get("json"))
		if tag == "" || tag == "-" {
			continue // skip fields without the `json` tag
		}

		switch t.Kind() {
		case reflect.String:
			err = key.SetStringValue(n, s.Field(i).String())
			if err != nil {
				return fmt.Errorf("Set %s = %s", n, s.Field(i).String())
			}
		case reflect.Bool:
			if s.Field(i).Bool() {
				err = key.SetStringValue(n, "true")
			} else {
				err = key.SetStringValue(n, "false")
			}

		default:
			return fmt.Errorf("Unsupported type for field %s: %v", n, t.Kind())
		}

		if err != nil {
			return fmt.Errorf("Unable to save %s in settings %s", n, err.Error())
		}
	}

	return nil
}

func (ss *SettingsStore) Delete() error {
	return registry.DeleteKey(registry.CURRENT_USER, REGISTRY_PATH)
}

func (ss *SettingsStore) GetKey() (registry.Key, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, REGISTRY_PATH,
		registry.ALL_ACCESS)

	if err != nil {
		key, _, err = registry.CreateKey(registry.CURRENT_USER,
			REGISTRY_PATH, registry.ALL_ACCESS)
	}

	return key, nil
}

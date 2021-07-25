package utils

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
	"net"
	"sort"
)

func GetLocalIPs() ([]string, error) {
	ips := []string{"All-Interfaces/0.0.0.0"}

	ifaces, err := net.Interfaces()
	if err != nil {
		return ips, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return ips, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP.String()
				ips = append(ips, fmt.Sprintf("%s/%s", i.Name, ip))
			case *net.IPAddr:
				ip := v.IP.String()
				ips = append(ips, fmt.Sprintf("%s/%s", i.Name, ip))
			}
		}
	}

	sort.Strings(ips)
	return ips, nil
}

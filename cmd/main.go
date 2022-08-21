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
	"math/rand"
	"net"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/synfinatic/alpacascope/alpaca"
	"github.com/synfinatic/alpacascope/skyfi"
	"github.com/synfinatic/alpacascope/telescope"
	"github.com/synfinatic/alpacascope/utils"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"
var Delta = ""

type TeleComms int

const (
	NexStar TeleComms = iota
	LX200
)

func main() {
	var lport int32        // listen port
	var lip string         // listen IP
	var clientid uint32    // alpaca client id
	var debug bool         // enable debugging
	var sport int32        // Alpaca server port
	var shost string       // Alpaca server IP
	var version bool       //  Version info
	var _mode string       // Comms mode
	var highPrecision bool // used for LX200
	var noAutoTrack bool
	var mode TeleComms
	var telescopeId uint32 // Alpaca telescope id.  Usually 0-10
	var _mount_type string // mount type
	var tracking_mode alpaca.TrackingMode

	flag.StringVar(&shost, "alpaca-host", "auto", "FQDN or IP address of Alpaca server")
	flag.Int32Var(&sport, "alpaca-port", 11111, "TCP port of the Alpaca server")
	flag.Uint32Var(&clientid, "clientid", 0, "Override Alpaca ClientID used for debugging")
	flag.Int32Var(&lport, "listen-port", 4030, "TCP port to listen on for clients")
	flag.StringVar(&lip, "listen-ip", "0.0.0.0", "IP to listen on for clients")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.StringVar(&_mode, "mode", "nexstar", "Comms mode: [nexstar|lx200]")
	flag.Uint32Var(&telescopeId, "telescope-id", 0, "Alpaca Telescope ID (default 0)")
	flag.StringVar(&_mount_type, "mount-type", "altaz", "Mount type: [altaz|eqn|eqs]")
	flag.BoolVar(&noAutoTrack, "no-auto-track", false, "Do not enable auto-track")
	flag.BoolVar(&highPrecision, "high-precision", false, "Default to High Precision LX200 mode")

	flag.Parse()

	if clientid == 0 {
		clientid = rand.Uint32()
		log.Debugf("Selecting random ClientID: %d", clientid)
	}

	// turn on debugging?
	if debug == true {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	} else {
		// pretty console output
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}

	ips, err := utils.GetLocalIPs()
	if err != nil {
		log.WithError(err).Errorf("Unable to GetLocalIPs()")
	} else {
		log.Errorf("found ips: %s", strings.Join(ips, ", "))
	}

	if version == true {
		delta := ""
		if len(Delta) > 0 {
			delta = fmt.Sprintf(" [%s delta]", Delta)
			Tag = "Unknown"
		}
		fmt.Printf("AlpacaScope Version %s -- Copyright 2021 Aaron Turner\n", Version)
		fmt.Printf("%s (%s)%s built at %s\n", CommitID, Tag, delta, Buildinfos)
		os.Exit(0)
	}

	switch _mode {
	case "nexstar":
		mode = NexStar
	case "lx200":
		mode = LX200
	default:
		log.Fatalf("Invalid mode: %s", _mode)
	}
	switch _mount_type {
	case "altaz":
		tracking_mode = alpaca.Alt_Az
	case "eqn":
		tracking_mode = alpaca.EQ_North
	case "eqs":
		tracking_mode = alpaca.EQ_South
	default:
		log.Fatalf("Invalid mount type: %s", _mount_type)
	}

	listen := fmt.Sprintf("%s:%d", lip, lport)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("Error listening on %s: %s", listen, err.Error())
	}
	defer ln.Close()

	if shost == "auto" {
		// first look locally since we can't rely on UDP broadcast to work locally on windows
		shost = alpaca.IsRunningLocal(sport)
		if shost == "" {
			shost, sport, err = alpaca.DiscoverServer(3)
			if err != nil {
				log.Fatalf("Unable to auto discover Alpaca Remote Server.  Please specify --alpaca-host and --alpaca-port")
			}
		}
	}

	// Act like a SkyFi for discovery
	go skyfi.ReplyDiscover()

	a := alpaca.NewAlpaca(clientid, shost, sport)
	scope := alpaca.NewTelescope(telescopeId, tracking_mode, a)

	connected, err := scope.GetConnected()
	if err != nil {
		log.Fatalf("Unable to determine status of telescope: %s", err.Error())
	}

	if !connected {
		err = scope.PutConnected(true)
		if err != nil {
			log.Fatalf("Unable to connect to telescope ID %d: %s", telescopeId, err.Error())
		}
	}

	name, err := scope.GetName()
	if err != nil {
		log.Warnf("Unable to determine name of telescope: %s", err.Error())
	} else {
		log.Infof("Connected to telescope %d: %s", telescopeId, name)
	}

	actions, err := scope.GetSupportedActions()
	if err != nil {
		log.Fatalf("Unable to determine supportedactions of telescope: %s", err.Error())
	}
	log.Debugf("SupportedActions: %s", actions)

	var tscope telescope.TelescopeProtocol
	switch mode {
	case LX200:
		minmax, err := scope.GetAxisRates(alpaca.AxisAzmRa)
		if err != nil {
			log.Errorf("Unable to query axis rates: %s", err.Error())
		}
		tscope = telescope.NewLX200(!noAutoTrack, highPrecision, true, minmax, 100000)
	case NexStar:
		tscope = telescope.NewNexStar(!noAutoTrack)
	default:
		log.Fatalf("Unsupported mode value: %d", mode)
	}

	log.Infof("Waiting for %s clients on %s:%d\n", _mode, lip, lport)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("Error calling Accept(): %s", err.Error())
			continue
		}

		tscope.HandleConnection(conn, scope)

		clientid += 1
	}
}

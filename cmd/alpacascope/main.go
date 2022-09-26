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

	"github.com/alecthomas/kong"
	"github.com/mattn/go-colorable"
	"github.com/synfinatic/alpacascope/alpaca"
	"github.com/synfinatic/alpacascope/skyfi"
	"github.com/synfinatic/alpacascope/telescope"
	"github.com/synfinatic/alpacascope/utils"

	log "github.com/sirupsen/logrus"
)

var Version = "unknown"
var CopyrightYears = "2021-2022"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"
var Delta = ""

type TeleComms int

const (
	NexStar TeleComms = iota
	LX200
)

type CLI struct {
	AlpacaHost    string `default:"auto" short:"H" help:"FQDN or IP address of Alpaca server"`
	AlpacaPort    int32  `default:"11111" short:"P" help:"TCP port of the Alpaca server"`
	ClientID      uint32 `default:"0" short:"c" help:"Override Alpaca ClientID used for debugging"`
	TelescopeID   uint32 `default:"0" short:"t" help:"Alpaca TelescopeID"`
	ListenIP      string `default:"0.0.0.0" help:"IP to listen on for clients"`
	ListenPort    int32  `default:"4030" help:"TCP port to listen on for clients (default: 4030)"`
	SerialPort    string `default:"/dev/alpacascope" short:"p" help:"Specify serial port to listen for connections"`
	Serial        bool   `short:"s" help:"Listen on serial port instead of network"`
	Mode          string `short:"m" default:"nexstar" enum:"nexstar,lx200" help:"Comms mode: [nexstar|lx200]"`
	MountType     string `default:"altaz" enum:"altaz,eqn,eqs" help:"Mount type: [altaz|eqn|eqs]"`
	HighPrecision bool   `help:"Default to High Precision in LX200 mode"`
	NoAutoTrack   bool   `help:"Do not enable auto-track"`
	Debug         bool   `help:"Enable debug logging"`
	Version       bool   `help:"Print version and exit"`
}

type RunContext struct {
	Kctx *kong.Context
	Cli  *CLI
}

func main() {
	cli := CLI{}
	parser := kong.Must(
		&cli,
		kong.Name("alpacascope"),
		kong.Description("Alpaca to Telescope Gateway"),
		kong.UsageOnError(),
		kong.Vars{},
	)
	_, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	var mode TeleComms
	var tracking_mode alpaca.TrackingMode

	if cli.ClientID == 0 {
		cli.ClientID = rand.Uint32()
		log.Debugf("Selecting random ClientID: %d", cli.ClientID)
	}

	// turn on debugging?
	if cli.Debug == true {
		log.SetFormatter(&log.TextFormatter{
			// DisableColors: true,
			FullTimestamp: true,
		})
		log.SetLevel(log.DebugLevel)
		// log.SetReportCaller(true)
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

	if cli.Version {
		delta := ""
		if len(Delta) > 0 {
			delta = fmt.Sprintf(" [%s delta]", Delta)
			Tag = "Unknown"
		}
		fmt.Printf("AlpacaScope Version %s -- Copyright %s Aaron Turner\n", Version, CopyrightYears)
		fmt.Printf("%s (%s)%s built at %s\n", CommitID, Tag, delta, Buildinfos)
		os.Exit(0)
	}

	switch cli.Mode {
	case "nexstar":
		mode = NexStar
	case "lx200":
		mode = LX200
	}
	switch cli.MountType {
	case "altaz":
		tracking_mode = alpaca.Alt_Az
	case "eqn":
		tracking_mode = alpaca.EQ_North
	case "eqs":
		tracking_mode = alpaca.EQ_South
	}

	var ln net.Listener
	if !cli.Serial {
		listen := fmt.Sprintf("%s:%d", cli.ListenIP, cli.ListenPort)
		ln, err = net.Listen("tcp", listen)
		if err != nil {
			log.Fatalf("Error listening on %s: %s", listen, err.Error())
		}
	} else {
		// do the serial port needful
	}
	defer ln.Close()

	if cli.AlpacaHost == "auto" {
		// first look locally since we can't rely on UDP broadcast to work locally on windows
		cli.AlpacaHost = alpaca.IsRunningLocal(cli.AlpacaPort)
		if cli.AlpacaHost == "" {
			cli.AlpacaHost, cli.AlpacaPort, err = alpaca.DiscoverServer(3)
			if err != nil {
				log.Fatalf("Unable to auto discover Alpaca Remote Server.  Please specify --alpaca-host and --alpaca-port")
			}
		}
	}

	// Act like a SkyFi for discovery
	go skyfi.ReplyDiscover()

	a := alpaca.NewAlpaca(cli.ClientID, cli.AlpacaHost, cli.AlpacaPort)
	scope := alpaca.NewTelescope(cli.TelescopeID, tracking_mode, a)

	connected, err := scope.GetConnected()
	if err != nil {
		log.Fatalf("Unable to determine status of telescope: %s", err.Error())
	}

	if !connected {
		err = scope.PutConnected(true)
		if err != nil {
			log.Fatalf("Unable to connect to telescope ID %d: %s", cli.TelescopeID, err.Error())
		}
		connected, err = scope.GetConnected()
		if err != nil {
			log.Fatalf("Unable to determine status of telescope: %s", err.Error())
		}
		if !connected {
			log.Fatalf("Telescope is not connected to ASCOM Remote")
		}
	}

	name, err := scope.GetName()
	if err != nil {
		log.Warnf("Unable to determine name of telescope: %s", err.Error())
	} else {
		log.Infof("Connected to telescope %d: %s", cli.TelescopeID, name)
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
		tscope = telescope.NewLX200(!cli.NoAutoTrack, cli.HighPrecision, true, minmax, 100000)
	case NexStar:
		tscope = telescope.NewNexStar(!cli.NoAutoTrack)
	default:
		log.Fatalf("Unsupported mode value: %d", mode)
	}

	log.Infof("Waiting for %s clients on %s:%d\n", cli.Mode, cli.ListenIP, cli.ListenPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("Error calling Accept(): %s", err.Error())
			continue
		}

		tscope.HandleConnection(conn, scope)
	}
}

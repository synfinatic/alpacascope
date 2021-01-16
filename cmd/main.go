package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/synfinatic/alpacascope/alpaca"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"

type TeleComms int

const (
	NexStar TeleComms = iota
	LX200
)

func main() {
	var lport int32     // listen port
	var lip string      // listen IP
	var clientid uint32 // alpaca client id
	var debug bool      // enable debugging
	var info bool       // enable info
	var sport int32     // Alpaca server port
	var shost string    // Alpaca server IP
	var version bool    //  Version info
	var _mode string    // Comms mode
	var mode TeleComms
	var telescopeId uint32 // Alpaca telescope id.  Usually 0-10

	flag.StringVar(&shost, "alpaca-host", "auto", "FQDN or IP address of Alpaca server")
	flag.Int32Var(&sport, "alpaca-port", 11111, "TCP port of the Alpaca server")
	flag.Uint32Var(&clientid, "clientid", 1, "Alpaca ClientID used for debugging")
	flag.Int32Var(&lport, "listen-port", 4030, "TCP port to listen on for clients")
	flag.StringVar(&lip, "listen-ip", "0.0.0.0", "IP to listen on for clients")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&info, "info", false, "Enable info logging")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.StringVar(&_mode, "mode", "nexstar", "Comms mode: [nexstar|lx200]")
	flag.Uint32Var(&telescopeId, "telescope-id", 0, "Alpaca Telescope ID")

	flag.Parse()

	// turn on debugging?
	if debug == true {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	} else if info == true {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if version == true {
		fmt.Printf("AlpacaScope Version %s -- Copyright 2020 Aaron Turner\n", Version)
		fmt.Printf("%s (%s) built at %s\n", CommitID, Tag, Buildinfos)
		os.Exit(0)
	}

	if strings.Compare(_mode, "nexstar") == 0 {
		mode = NexStar
	} else if strings.Compare(_mode, "lx200") == 0 {
		mode = LX200
	} else {
		log.Fatalf("Invalid mode: %s", _mode)
	}

	listen := fmt.Sprintf("%s:%d", lip, lport)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("Error listening on %s: %s", listen, err.Error())
	}

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

	a := alpaca.NewAlpaca(clientid, shost, sport)
	telescope := alpaca.NewTelescope(telescopeId, a)

	connected, err := telescope.GetConnected()
	if err != nil {
		log.Fatalf("Unable to determine status of telescope: %s", err.Error())
	}

	if !connected {
		log.Fatalf("Unable to connect to telescope ID %d: %s", telescopeId, a.ErrorMessage)
	}

	name, err := telescope.GetName()
	if err != nil {
		log.Warnf("Unable to determine name of telescope: %s", err.Error())
	} else {
		log.Infof("Connected to telescope %d: %s", telescopeId, name)
	}

	actions, err := telescope.GetSupportedActions()
	if err != nil {
		log.Fatalf("Unable to determine supportedactions of telescope: %s", err.Error())
	}
	log.Debugf("SupportedActions: %s", actions)

	var lxstate LX200State
	if mode == LX200 {
		minmax, err := telescope.GetAxisRates(alpaca.AxisAzmRa)
		if err != nil {
			log.Errorf("Unable to query axis rates: %s", err.Error())
		}
		lxstate = LX200State{
			HighPrecision:  true,
			TwentyFourHour: true,
			MaxSlew:        minmax["Maximum"],
			MinSlew:        minmax["Minimum"],
			SlewRate:       int(minmax["Maximum"]),
		}
	}

	fmt.Printf("Waiting for %s clients on %s:%d\n", _mode, lip, lport)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("Error calling Accept(): %s", err.Error())
			continue
		}

		if mode == LX200 {
			go handleLX200Conn(conn, telescope, &lxstate)
		} else if mode == NexStar {
			go handleNexstar(conn, telescope)
		} else {
			log.Fatalf("Unsupported mode value: %d", mode)
		}
		clientid += 1
	}
}

package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/synfinatic/alpaca-gateway/alpaca"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"

const (
	NexStar int = iota
	LX200
)

func main() {
	var lport int32     // listen port
	var lip string      // listen IP
	var clientid uint32 // alpaca client id
	var debug bool      // enable debugging
	var sport int32     // Alpaca server port
	var sip string      // Alpaca server IP
	var version bool    //  Version info
	var _mode string    // Comms mode
	var mode int

	flag.StringVar(&sip, "alpaca-ip", "127.0.0.1", "IP address of Alpaca server")
	flag.Int32Var(&sport, "alpaca-port", 11111, "TCP port of the Alpaca server")
	flag.Uint32Var(&clientid, "clientid", 0, "Alpaca ClientID used for debugging")
	flag.Int32Var(&lport, "listen-port", 5150, "TCP port to listen on for clients")
	flag.StringVar(&lip, "listen-ip", "0.0.0.0", "IP to listen on for clients")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.StringVar(&_mode, "mode", "nexstar", "Comms mode: [nexstar|lx200]")

	flag.Parse()

	// turn on debugging?
	if debug == true {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if version == true {
		fmt.Printf("alpaca-gateway Version %s -- Copyright 2020 Aaron Turner\n", Version)
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

	a := alpaca.NewAlpaca(clientid, sip, sport)
	telescope := alpaca.NewTelescope(0, a)

	connected, err := telescope.GetConnected()
	if err != nil {
		log.Fatalf("Unable to determine status of telescope: %s", err.Error())
	}

	name, err := telescope.GetName()
	if err != nil {
		log.Fatalf("Unable to determine name of telescope: %s", err.Error())
	}
	if connected {
		log.Infof("Connected to telescope: %s", name)
	} else {
		log.Warnf("Not connected to telescope: %s", name)
	}

	actions, err := telescope.GetSupportedActions()
	if err != nil {
		log.Fatalf("Unable to determine supportedactions of telescope: %s", err.Error())
	}
	log.Debugf("SupportedActions: %s", actions)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("Error calling Accept(): %s", err.Error())
			continue
		}

		if mode == LX200 {
			go handleLX200Conn(conn, telescope)
		} else if mode == NexStar {
			go handleNexStar(conn, telescope)
		} else {
			log.Fatalf("Unsupported mode value: %d", mode)
		}
		clientid += 1
	}
}

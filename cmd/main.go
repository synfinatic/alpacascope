package main

import (
	"fmt"
	"net"
	"os"
	"strings"

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
	var lport int32  // listen port
	var lip string   // listen IP
	var debug bool   // enable debugging
	var sport int32  // Alpaca server port
	var sip string   // Alpaca server IP
	var version bool //  Version info
	var _mode string // Comms mode
	var mode int

	flag.StringVar(&sip, "alpaca-ip", "127.0.0.1", "IP address of Alpaca server")
	flag.Int32Var(&sport, "alpaca-port", 11111, "TCP port of the Alpaca server")
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

	var clientid int32 = 1
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("Error calling Accept(): %s", err.Error())
			continue
		}

		if mode == LX200 {
			go handleLX200Conn(conn, clientid)
		} else if mode == NexStar {
			go handleNexStar(conn, clientid)
		} else {
			log.Fatalf("Unsupported mode value: %d", mode)
		}
		clientid += 1
	}
}

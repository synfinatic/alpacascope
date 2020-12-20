package main

import (
	"github.com/synfinatic/alpaca-gateway/alpaca"
	"net"

	log "github.com/sirupsen/logrus"
)

func handleLX200Conn(conn net.Conn, t *alpaca.Telescope) {
	buf := make([]byte, 1024)
	len, err := conn.Read(buf)
	if err != nil {
		log.Errorf("reading from LX200 client: %s", err.Error())
	} else {
		log.Debugf("read %d bytes from NexStar client", len)
	}
}

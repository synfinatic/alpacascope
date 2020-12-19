package main

import (
	"net"

	log "github.com/sirupsen/logrus"
)

func handleNexStar(conn net.Conn, clientid int32) {
	buf := make([]byte, 1024)
	len, err := conn.Read(buf)
	if err != nil {
		log.Errorf("reading from NexStar client: %s", err.Error())
	} else {
		log.Debugf("read %d bytes from NexStar client", len)
	}
}

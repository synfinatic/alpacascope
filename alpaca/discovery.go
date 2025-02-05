package alpaca

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ALPACA_DISCOVERY_VERSION = 1
	DEFAULT_PORT             = 11111
	DEFAULT_PORT_STR         = "11111"
)

type AlpacaDiscoveryMessage struct {
	Fixed    []byte // must be 'alpacadiscovery'
	Version  byte
	Reserved []byte
}

type AlpacaResponseMessage struct {
	AlpacaPort uint16 `json:"AlpacaPort"`
}

func (a *AlpacaDiscoveryMessage) Bytes() []byte {
	var buf []byte = []byte{}

	buf = append(buf[:], a.Fixed[:]...)
	buf = append(buf[:], a.Version)
	buf = append(buf[:], a.Reserved[:]...)
	return buf
}

func (a AlpacaDiscoveryMessage) String() string {
	return fmt.Sprintf("%s%c", string(a.Fixed), a.Version)
}

func NewAlpacaDiscoveryMessage(version int) *AlpacaDiscoveryMessage {
	a := AlpacaDiscoveryMessage{
		Fixed:    []byte("alpacadiscovery"),
		Version:  byte(fmt.Sprintf("%d", version)[0]),
		Reserved: make([]byte, 48),
	}
	return &a
}

// checks if server is on localhost.  returns IP or empty string
func IsRunningLocal(port int32) string {
	log.Infof("Looking for Alpaca Remote Server locally on port %d...", port)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Errorf("Unable to determine local interface addresses: %s", err.Error())
		return ""
	}
	log.Debugf("local addrs: %v", addrs)
	for _, addr := range addrs {
		ips := strings.Split(addr.String(), "/")
		ip := net.ParseIP(ips[0])
		if ip.To4() == nil {
			continue // skip non-IPv4
		}

		if tryAlpaca(ip.String(), port) {
			log.Infof("Found Alpaca on %s:%d", ip.String(), port)
			return ip.String()
		} else {
			log.Debugf("Alpaca is not running on %s:%d", ip.String(), port)
		}
	}
	log.Info("No local Alpaca Remote Servers found")
	return ""
}

// Send a discovery packet to the given IP to see if it's Alpaca
func tryAlpaca(ip string, port int32) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*1)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Discover any alpaca servers.  returns IP as string and port
func DiscoverServer(tries int) (string, int32, error) {
	pc, err := net.ListenPacket("udp4", ":32227")
	if err != nil {
		return "", 0, fmt.Errorf("Unable to open Alpaca discovery listen socket: %s", err.Error())
	}
	defer pc.Close()

	sendAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:32227")
	if err != nil {
		return "", 0, fmt.Errorf("Unable to resolve Alpaca broadcast address: %s", err.Error())
	}

	adm := NewAlpacaDiscoveryMessage(ALPACA_DISCOVERY_VERSION)
	msgBytes := adm.Bytes()
	buf := make([]byte, 1024)

	for i := 0; i < tries; i++ {
		_, err = pc.WriteTo(msgBytes, sendAddr)
		if err != nil {
			return "", 0, fmt.Errorf("Unable to send Alpaca discovery message: %s", err.Error())
		}

		deadline := time.Now().Add(time.Second * 1)
		for {
			if err = pc.SetReadDeadline(deadline); err != nil {
				log.Warnf("Unable to set read deadline: %s", err.Error())
				break
			}
			n, addr, err := pc.ReadFrom(buf)
			log.Debugf("receved %d bytes", n)
			if err != nil {
				log.Warnf("Failed to discover Alpaca server: %s", err.Error())
				break // don't try reading again this cycle
			} else if n == 64 && strings.HasPrefix(string(buf[:n]), adm.String()) {
				continue // this is the message we sent
			}

			ip := strings.Split(addr.String(), ":")
			log.Debugf("receved %d bytes via discovery: %v", n, buf[:n])
			var a AlpacaResponseMessage
			err = json.Unmarshal(buf[:n], &a)
			if err != nil {
				log.Warnf("Unable to decode message: %s", err.Error())
				break
			} else {
				log.Infof("Discovered Alpaca Server on %s:%d", ip[0], a.AlpacaPort)
				return ip[0], int32(a.AlpacaPort), nil
			}
		}
	}
	return "", 0, fmt.Errorf("No reply from Alpaca Server")
}

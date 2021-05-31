package telescope

import (
	"net"

	alpaca "github.com/synfinatic/alpacascope/alpaca"
)

type TelescopeProtocol interface {
	HandleConnection(net.Conn, *alpaca.Telescope)
}

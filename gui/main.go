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
	"net"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/synfinatic/alpacascope/alpaca"
	"github.com/synfinatic/alpacascope/skyfi"
	"github.com/synfinatic/alpacascope/telescope"
	"github.com/synfinatic/alpacascope/utils"
)

var sbox *StatusBox

func main() {
	app := app.New()
	w := app.NewWindow("AlpacaScope")

	sbox = NewStatusBox(6)
	config, err := NewAlpacaScopeConfig()
	if err != nil {
		sbox.AddLine("Using default settings.")
	} else {
		sbox.AddLine("Loaded your saved settings.")
	}

	mountType := widget.NewSelect(
		[]string{"Alt-Az", "EQ North", "EQ South"},
		func(val string) {
			config.TelescopeMount = val
		},
	)
	mountType.Selected = config.TelescopeMount

	protocolDropdown := widget.NewSelect([]string{"NexStar", "LX200"},
		func(proto string) {
			config.TelescopeProtocol = proto
			if proto == "NexStar" {
				mountType.Enable()
			} else {
				// only NexStar supports the mountType
				mountType.Disable()
			}
		},
	)
	protocolDropdown.Selected = config.TelescopeProtocol

	ips, err := utils.GetLocalIPs()
	if err != nil {
		ips = []string{config.ListenIp}
		sbox.AddLine(fmt.Sprintf("Unable to detect interfaces: %s", err.Error()))
	}
	ipEntry := widget.NewSelect(ips, func(ip string) {})
	ipEntry.Selected = config.ListenIp

	portEntry := widget.NewEntry()
	portEntry.SetText(config.ListenPort)
	portEntry.Validator = validation.NewRegexp("^[1-9][0-9]+$",
		"Invalid TCP Port number")

	ascomEntryIP := widget.NewEntry()
	ascomEntryIP.SetText(config.AscomIp)
	ascomEntryIP.Validator = validation.NewRegexp("^([0-9]+\\.){3}[0-9]+$",
		"Must be a valid IPv4 address")

	ascomEntryPort := widget.NewEntry()
	ascomEntryPort.SetText(config.AscomPort)
	ascomEntryPort.Validator = validation.NewRegexp("^[1-9][0-9]+$",
		"Must be a valid integer > 1")

	autodiscover := widget.NewCheck("", func(enabled bool) {
		switch enabled {
		case true:
			config.AscomAuto = true
			ascomEntryIP.Disable()
			ascomEntryPort.Disable()
		case false:
			config.AscomAuto = false
			ascomEntryIP.Enable()
			ascomEntryPort.Enable()
		}

	})
	autodiscover.Checked = config.AscomAuto
	if config.AscomAuto {
		ascomEntryIP.Disable()
		ascomEntryPort.Disable()
	}

	telescopeId := widget.NewSelect(
		[]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		func(val string) {
			config.AscomTelescope = val
		},
	)
	telescopeId.Selected = config.AscomTelescope

	sbox.AddLine("Press 'Start' when ready.")

	top := widget.NewForm(
		widget.NewFormItem("Telescope Protocol", protocolDropdown),
		widget.NewFormItem("Mount Type", mountType),
		widget.NewFormItem("Listen IP", ipEntry),
		widget.NewFormItem("Listen Port", portEntry),
		widget.NewFormItem("Auto Discover ASCOM Remote", autodiscover),
		widget.NewFormItem("ASCOM Remote Server IP", ascomEntryIP),
		widget.NewFormItem("ASCOM Remote Port", ascomEntryPort),
		widget.NewFormItem("ASCOM Telescope ID", telescopeId),
	)
	top.OnSubmit = func() {
		config.Validate()
		go config.Run()
	}
	top.OnCancel = func() {
		if config.IsRunning() {
			config.Quit <- true
		} else {
			sbox.AddLine("AlpacaScope isn't running.")
		}
	}
	top.SubmitText = "Start AlpacaScope Services"
	top.CancelText = "Stop AlpacaScope Services"

	padded := container.NewPadded()
	spacer := layout.NewSpacer()

	quit := widget.NewButton("Quit", func() { os.Exit(0) })
	save := widget.NewButton("Save Settings", func() {
		err := config.Save()
		if err != nil {
			sbox.AddLine(err.Error())
		}
	})

	bottom := container.NewHBox(spacer, save, spacer, quit, spacer)

	w.SetContent(container.NewPadded(
		container.NewBorder(
			top, bottom, padded, padded,
			container.NewGridWithColumns(1, sbox.Widget()),
		)))

	w.ShowAndRun()
}

func (c *AlpacaScopeConfig) Run() {
	accptedFirstConnection := false
	var clientid uint32 = 1
	var sport int32
	var shost string
	var err error

	if c.isRunning {
		sbox.AddLine("AlpacaScope services are already running!")
		return
	}
	sbox.AddLine("Starting AlpacaScope services.")
	c.isRunning = true

	if c.AscomAuto {
		// first look locally since we can't rely on UDP broadcast to work locally on windows
		sport = alpaca.DEFAULT_PORT
		shost = alpaca.IsRunningLocal(sport)
		if shost == "" {
			for i := 0; i < 3; i++ {
				shost, sport, err = alpaca.DiscoverServer(1)
				if err == nil {
					sbox.AddLine(fmt.Sprintf("Found ASCOM Remote: %s:%d", shost, sport))
					break
				} else {
					sbox.AddLine(err.Error())
				}
			}
		} else {
			sbox.AddLine(fmt.Sprintf("Found ASCOM Remote: %s:%d", shost, sport))
		}
	} else {
		// Use user provided values
		shost = c.AscomIp
		x, _ := strconv.ParseInt(c.AscomPort, 10, 32)
		sport = int32(x)
	}

	if shost == "" {
		sbox.AddLine("Unable to auto-discover ASCOM Remote Server")
		sbox.AddLine("Check configuration and press 'Start'")
		c.isRunning = false
		return
	}

	var tracking_mode alpaca.TrackingMode
	switch c.TelescopeMount {
	case "Alt-Az":
		tracking_mode = alpaca.Alt_Az
	case "EQ North":
		tracking_mode = alpaca.EQ_North
	case "EQ South":
		tracking_mode = alpaca.EQ_South
	}

	a := alpaca.NewAlpaca(clientid, shost, sport)
	tid, _ := strconv.ParseUint(c.AscomTelescope, 10, 32)
	scope := alpaca.NewTelescope(uint32(tid), tracking_mode, a)

	connected, err := scope.GetConnected()
	if err != nil {
		sbox.AddLine(fmt.Sprintf("Unable to determine status of telescope: %s", err.Error()))
		sbox.AddLine("Check configuration and press 'Start'")
		c.isRunning = false
		return
	}

	if !connected {
		sbox.AddLine(fmt.Sprintf("Unable to connect to telescope ID %s: %s", c.AscomTelescope, a.ErrorMessage))
		sbox.AddLine("Check configuration and press 'Start'")
		c.isRunning = false
		return
	}

	name, err := scope.GetName()
	if err != nil {
		sbox.AddLine(fmt.Sprintf("Connected to unknown telescope: %s", err.Error()))
	} else {
		sbox.AddLine(fmt.Sprintf("Connected to telescope %s: %s", c.AscomTelescope, name))
	}

	var tscope telescope.TelescopeProtocol
	switch c.TelescopeProtocol {
	case "LX200":
		minmax, err := scope.GetAxisRates(alpaca.AxisAzmRa)
		if err != nil {
			sbox.AddLine(fmt.Sprintf("Unable to query axis rates: %s", err.Error()))
		}
		tscope = telescope.NewLX200(true, true, minmax, 100000)

	case "NexStar":
		tscope = telescope.NewNexStar()
	}

	// Act like SkyFi
	shutdownSkyFi := make(chan bool)
	go skyfi.ReplyDiscoverWithShutdown(shutdownSkyFi)
	// go skyfi.ReplyDiscover()

	newConns := make(chan net.Conn)

	listen := fmt.Sprintf("%s:%s", c.ListenIp, c.ListenPort)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		sbox.AddLine(fmt.Sprintf("Error listening on %s: %s", listen, err.Error()))
		sbox.AddLine("Check configuration and press 'Start'")
		c.isRunning = false
		return
	}
	defer ln.Close()
	sbox.AddLine(fmt.Sprintf("Ready to accept connections on %s:%s", c.ListenIp, c.ListenPort))

	// goroutine for our listener
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				// handle error (and then for example indicate acceptor is down)
				if !c.isRunning {
					break
				}
				sbox.AddLine(fmt.Sprintf("Error calling Accept(): %s", err.Error()))
				newConns <- nil
				continue
			}
			newConns <- conn
		}
	}(ln)

	for {
		select {
		case <-c.Quit:
			sbox.AddLine("Shutting down AlpacaScope services.")
			shutdownSkyFi <- true
			c.isRunning = false
			return

		case conn := <-newConns:
			if !accptedFirstConnection {
				sbox.AddLine(fmt.Sprintf("Accepted connection from: %s", conn.RemoteAddr().String()))
				accptedFirstConnection = true
			}

			tscope.HandleConnection(conn, scope)

			clientid += 1
		}
	}
}

type StatusBox struct {
	TextGrid *widget.TextGrid
	Lines    int
	numLines int
	lines    []string
}

func NewStatusBox(lineCount int) *StatusBox {
	status := widget.NewTextGrid()
	var zeroValue []string

	for i := 0; i < lineCount; i++ {
		zeroValue = append(zeroValue, "")
	}
	status.SetText(strings.Join(zeroValue, "\n"))

	sbox := StatusBox{
		TextGrid: status,
		Lines:    lineCount,
		numLines: 0,
		lines:    []string{},
	}
	return &sbox
}

func (sb *StatusBox) AddLine(line string) {
	sb.lines = append(sb.lines, line)
	for len(sb.lines) > sb.Lines {
		sb.lines = sb.lines[1:]
	}
	displayLines := sb.lines
	for len(displayLines) < sb.Lines {
		displayLines = append(displayLines, "")
	}

	lines := strings.Join(displayLines, "\n")
	sb.TextGrid.SetText(lines)
}

func (sb *StatusBox) Widget() *widget.TextGrid {
	return sb.TextGrid
}

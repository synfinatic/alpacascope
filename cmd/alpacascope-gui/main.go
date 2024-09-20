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
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/synfinatic/alpacascope/alpaca"
	"github.com/synfinatic/alpacascope/skyfi"
	"github.com/synfinatic/alpacascope/telescope"

	log "github.com/sirupsen/logrus"
)

const (
	RUNNING                = "Status: AlpacaScope is running!"
	STOPPED                = "Status: AlpacaScope is stopped."
	CHECK                  = "Check configuration and press 'Start'"
	DEFAULT_DISCOVER_TRIES = 3
)

var sbox *StatusBox

type Widgets struct {
	form                *widget.Form
	TelescopeProtocol   *widget.Select
	TelescopeMount      *widget.Select
	AutoTracking        *widget.Check
	HighPrecisionLX200  *widget.Check
	ListenIp            *widget.Select
	ListenPort          *widget.Entry
	AscomAuto           *widget.Check
	AutoConnectAttempts *widget.Select
	AutoStart           *widget.Check
	AscomIp             *widget.Entry
	AscomPort           *widget.Entry
	AscomTelescope      *widget.Select
	Status              *widget.TextGrid
	Save                *widget.Button
	Delete              *widget.Button
}

func main() {
	app := app.New()
	w := app.NewWindow("AlpacaScope")

	sbox = NewStatusBox(6)
	config := NewAlpacaScopeConfig()
	err := config.Load()
	if err != nil {
		sbox.AddLine(err.Error())
		sbox.AddLine("Using default settings.")
	} else {
		sbox.AddLine("Loaded your saved settings.")
	}

	ourWidgets := NewWidgets(config)
	if !config.AutoStart {
		sbox.AddLine("Press 'Start' when ready.")
	} else {
		sbox.AddLine("Automatically connecting...")
		go config.Run()
		go ourWidgets.Manager(config)
	}

	top := widget.NewForm(
		widget.NewFormItem("Telescope Protocol", ourWidgets.TelescopeProtocol),
		widget.NewFormItem("NexStar Mount Type", ourWidgets.TelescopeMount),
		widget.NewFormItem("LX200 default to High Precision", ourWidgets.HighPrecisionLX200),
		widget.NewFormItem("Auto Tracking", ourWidgets.AutoTracking),
		widget.NewFormItem("Listen IP", ourWidgets.ListenIp),
		widget.NewFormItem("Listen Port", ourWidgets.ListenPort),
		widget.NewFormItem("Auto Discover Alpaca Mount", ourWidgets.AscomAuto),
		widget.NewFormItem("ASCOM Remote Server IP", ourWidgets.AscomIp),
		widget.NewFormItem("ASCOM Remote Port", ourWidgets.AscomPort),
		widget.NewFormItem("ASCOM Telescope ID", ourWidgets.AscomTelescope),
		widget.NewFormItem("Automatically Connect on Start", ourWidgets.AutoStart),
		widget.NewFormItem("Connect Attempts", ourWidgets.AutoConnectAttempts),
	)
	ourWidgets.form = top

	ourWidgets.Save = widget.NewButton("Save Settings", func() {
		err := config.Save()
		if err != nil {
			sbox.AddLine(err.Error())
		} else {
			sbox.AddLine("Saved config settings.")
		}
	})
	ourWidgets.Delete = widget.NewButton("Reset Settings", func() {
		config = NewAlpacaScopeConfig()
		ourWidgets.Set(config)
		sbox.AddLine("Current settings reset to defaults.")
	})

	top.OnSubmit = func() {
		if config.IsRunning() {
			sbox.AddLine("AlpacaScope services are already running!")
			return
		}

		go config.Run()
		go ourWidgets.Manager(config)
	}

	top.OnCancel = func() {
		if config.IsRunning() {
			config.Quit <- true
		} else {
			sbox.AddLine("AlpacaScope isn't running.")
		}
		ourWidgets.Save.Enable()
		ourWidgets.Delete.Enable()
	}
	top.SubmitText = "Start AlpacaScope Services"
	top.CancelText = "Stop AlpacaScope Services"

	padded := container.NewPadded()
	spacer := layout.NewSpacer()

	quit := widget.NewButton("Quit", func() { os.Exit(0) })

	bottom := container.NewHBox(ourWidgets.Delete, ourWidgets.Save, spacer, quit, spacer)

	w.SetContent(container.NewPadded(
		container.NewBorder(
			top, bottom, padded, padded,
			container.NewVBox(
				container.NewHBox(spacer, ourWidgets.Status, spacer),
				padded,
				padded,
				container.NewGridWithColumns(1, sbox.Widget())),
		)))

	w.ShowAndRun()
}

// Waits until we are no longer running and then re-enables the buttons
func (w *Widgets) Manager(config *AlpacaScopeConfig) {
	// disable the buttons
	w.Disable()
	w.Status.SetText(RUNNING)

	// wait until we are Quitting our main Run() loop
	select {
	case <-config.EnableButtons:
		break
	}
	w.Enable()
	w.Status.SetText(STOPPED)
}

// preConnectQuit does a stop before we have connected or started skyfi
func preConnectQuit(c *AlpacaScopeConfig, stop chan bool) {
	for {
		select {
		case <-stop:
			return

		case <-c.Quit:
			c.isRunning = false
			c.EnableButtons <- true
		}
	}
}

func (c *AlpacaScopeConfig) Run() {
	accptedFirstConnection := false
	var clientid uint32 = rand.Uint32()
	var sport int32
	var shost string
	var err error

	sbox.AddLine(fmt.Sprintf("Using Alpaca ClientID: %d", clientid))

	tempQuit := make(chan bool)
	go preConnectQuit(c, tempQuit)
	c.isRunning = true

	sbox.Clear()

	if c.AscomAuto {
		// first look locally since we can't rely on UDP broadcast to work locally on windows
		sport = alpaca.DEFAULT_PORT
		shost = alpaca.IsRunningLocal(sport)
		if shost == "" {
			count := math.MaxUint32
			if c.AutoConnectAttempts != "Unlimited" {
				x, err := strconv.ParseInt(c.AutoConnectAttempts, 10, 32)
				if err != nil {
					count = 3
					log.Errorf("Error parsing AutoConnectAttempts '%s', using default %d", c.AutoConnectAttempts,
						DEFAULT_DISCOVER_TRIES)
				}
				count = int(x)
			}
			for i := 1; i <= count && c.isRunning; i++ {
				shost, sport, err = alpaca.DiscoverServer(1)
				if err == nil {
					sbox.AddLine(fmt.Sprintf("Found ASCOM Remote: %s:%d", shost, sport))
					break
				} else {
					if c.AutoConnectAttempts != "Unlimited" {
						sbox.AddLine(fmt.Sprintf("%d/%d %s", i, count, err.Error()))
					} else {
						sbox.AddLine(fmt.Sprintf("%d/Unlimited %s", i, err.Error()))
					}
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
		sbox.AddLine("Unable to auto-discover Alpaca/ASCOM Remote Server")
		sbox.AddLine(CHECK)
		tempQuit <- true
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
	var connected bool = false
	var connectAttempts int64 = 1
	if c.AutoStart {
		connectAttempts, _ = strconv.ParseInt(c.AutoConnectAttempts, 10, 32)
		sbox.AddLine(fmt.Sprintf("Attempting connecting to TelescopeID=%s %d times",
			c.AscomTelescope, connectAttempts))
	}

	for i := 1; !connected && int64(i) <= connectAttempts && c.isRunning; i++ {
		connected, err = scope.GetConnected()
		if err != nil {
			line := fmt.Sprintf("%d/%d Unable to connect to TelescopeID=%s: %s", i, connectAttempts, c.AscomTelescope, err.Error())
			sbox.AddLine(line)
			time.Sleep(time.Second)
		}
	}

	if !connected {
		// Manually connect
		err = scope.PutConnected(true)
		if err != nil {
			sbox.AddLine(fmt.Sprintf("Unable to connect to TelescopeID=%s: %s", c.AscomTelescope, err.Error()))
			sbox.AddLine(err.Error())
			sbox.AddLine(CHECK)
			tempQuit <- true
			return
		}
		connected, err = scope.GetConnected()
		if err != nil || !connected {
			sbox.AddLine(fmt.Sprintf("Unable to connect to TelescopeID=%s: %s", c.AscomTelescope, err.Error()))
			sbox.AddLine(err.Error())
			sbox.AddLine(CHECK)
			tempQuit <- true
			return
		}
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
		tscope = telescope.NewLX200(c.AutoTracking, true, true, minmax, 100000)

	case "NexStar":
		tscope = telescope.NewNexStar(c.AutoTracking)
	}

	// Act like SkyFi
	shutdownSkyFi := make(chan bool)
	go skyfi.ReplyDiscoverWithShutdown(shutdownSkyFi)

	newConns := make(chan net.Conn)

	listen := fmt.Sprintf("%s:%s", c.ListenIpAddress(), c.ListenPort)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		sbox.AddLine(fmt.Sprintf("Error listening on %s: %s", listen, err.Error()))
		sbox.AddLine(CHECK)
		tempQuit <- true
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

	// stop our temp quit handler
	tempQuit <- true

	// main loop
	for {
		select {
		case <-c.Quit:
			sbox.AddLine("Shutting down...")
			c.isRunning = false
			c.EnableButtons <- true
			shutdownSkyFi <- true
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

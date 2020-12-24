# AlpacaScope

## Why?

TL;DR: I have a [Celestron Evolution EdgeHD 800](
https://www.celestron.com/products/nexstar-evolution-8-hd-telescope-with-starsense)
and I'd like to be able to control it via [SkySafari](https://skysafariastronomy.com).

Unfortunately, the WiFi on the Evolution mount is known to be a bit flakey and
it's annoying when it drops out and you have to exit the app, reconnect to WiFi,
restart the app, and re-connect to the telescope.

So I decided to create this little application which allows me to control the
telescope via [CWPI](
https://www.celestron.com/pages/celestron-pwi-telescope-control-software)
which runs on the PC I have connected to the telescope via USB.  Of course,
SkySafari can't talk directly to CWPI.  CWPI supports [ASCOM](
https://ascom-standards.org) but that only allows IPC via Windows COM
which doesn't even support talking to programs on other computers (or my iPad).

But then in 2019, ASCOM introduced [Alpaca](
https://ascom-standards.org/Developer/Alpaca.htm) which via 
[Alpaca Remote Server](https://github.com/ASCOMInitiative/ASCOMRemote/releases)
exposes the ASCOM API via REST.  Of course, SkySafari doesn't support this (yet)
so I decided to write a service which emulates a telescope SkySafari supports
and talks to Alpaca Remote Server.  The result is that SkySafari can now control
my Celestron Evolution mount, or any mount that supports the ASCOM standard.

## Usage

So basically, my setup looks sorta like this:

```
Celestron Evolution <-> CWPI <-> ASCOM <-> Alpaca Server <-> AlpacaScope <-> SkySafari
             Focuser  <-'           `-> Sharp Cap
                                     `-> ZWO Camera
```

Basically, just download the binary for your system (easist to run on the same Windows
box as ASCOM & the Alpaca Remote Server) and run it.  By default it will try to connect
to Alpaca Remote Server on the local host (127.0.0.0) on port 11111 and listen on port 4030.

Configure SkySafari or other remote control software to connect to AlpacaScope on port
4030 using the Celestron Nexstar (I use the Nexstar/Advanced GT) protocol.

## FAQ

#### What do I need at minimum?

 1. A telescope connected to a Windows machine
 2. ASCOM installed an configured for your mount
 3. Alpaca Remote Server installed, configured & running
 4. AlpacaScope installed and running
 5. Some kind astronomy software which talks the LX200 or Celestron Nexstar protocols
    (SkySafari, etc)

#### Does this support SkySafari on Mac, iPad, Windows, Android, etc?
Yes, this supports all versions of SkySafari which allow for controlling telescopes.
Typically this is SkySafari Plus and Pro.

#### What about other astronomy software?
Yep, anything that can do Celestron Nexstar or LX200 protocols over TCP/IP
should work.  

#### What features work?

 * Manual slewing
 * Controlling slew speed
 * Goto a target
 * Align on target (maybe?)

#### Does AlpacaScope support [INDI](https://www.indilib.org)?
No it doesn't.  There's probably no reason it can't support INDI, but CWPI
doesn't support INDI so I have no easy way of developing/testing the code.

#### Does AlpacaScope need to run on the same computer as CWPI/ASCOM?
No, but that is probably the most common solution.  Alpaca-Gateway just needs
to be able to talk to the Alpaca Server running on the same computer as the
ASCOM driver connected to your telescope mount.

#### My software only supports the Meade LX200 protocol.  Can I use that?
Yes!  Slewing and GoTo works.  Be sure to specify `--mode lx200`.
I believe align/sync should work but it doesn't seem to work with my mount/CWPI.
Not sure if it's a bug on my end or a limitation with CWPI?

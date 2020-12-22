# Alpaca-Gateway

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
https://ascom-standards.org/Developer/Alpaca.htm) which basically exposes
the COM API via REST.  Of course, SkySafari doesn't support this (yet) so
I decided to write a service which emulates a telescope SkySafari supports
and talks to Alpaca.  The result is that SkySafari can now control my Celestron
Evolution mount, or any mount that supports the ASCOM standard.

## Usage

So basically, my setup looks sorta like this:

```
Celestron Evolution <-> CWPI <-> ASCOM <-> Alpaca Server <-> Alpaca-Gateway <-> SkySafari
             Focuser  <-'           `-> Sharp Cap
                                     `-> ZWO Camera
```

Basically, just download the binary for your system (easist to run on the same Windows
box as ASCOM & the Alpaca Remote Server) and run it.  By default it will try to connect
to Alpaca on the local host (127.0.0.0) on port 11111 and listen on port 4030.

Configure SkySafari or other remote control software to connect to Alpaca-Gateway on port
4030 using the Celestron Nexstar (I use the Nexstar/Advanced GT) protocol.

## FAQ

#### Does this support SkySafari on Mac, iPad, Windows, Android, etc?
Yes, this supports all versions of SkySafari which allow for controlling telescopes.
Typically this is SkySafari Plus and Pro.

#### What about other astronomy software?
Yep, anything that can do Celestron Nexstar protocol over TCP/IP should work.
That said, I haven't yet implimented the complete Nexstar protocol so there
may be issues.  Please open a bug report if you find any issues!

#### Does Alpaca-Gateway support [INDI](https://www.indilib.org)?
No it doesn't.  There's probably no reason it can't support INDI, but CWPI
doesn't support INDI so I have no easy way of developing/testing the code.

#### Does Alpaca-Gateway need to run on the same computer as CWPI/ASCOM?
No, but that is probably the most common solution.  Alpaca-Gateway just needs
to be able to talk to the Alpaca Server running on the same computer as the
ASCOM driver connected to your telescope mount.

#### My software only supports the Meade LX200 protocol.  Can I use that?
Not yet.  I plan on adding that in a future version.

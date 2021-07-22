# AlpacaScope Changelog

## Unreleased

## v2.0.0 - 2021-07-24

Added:

 - Add GUI option for Windows & OSX users
 - All releases will now be cryptographically signed with GPG

## v1.0.0 - 2021-01-23

Added:

 - Support for Nexstar tracking commands: t and T
 - Add date/time processing for LX200 GPS

Fixed:

 - Fix crash when setting Lat/Long due to index out of range

Changed:

 - Clean up NexStar cmd processing to be cleaner
 - Huge update to README.md

## v0.0.5 - 2021-01-17

Changed:

 - Removed --info flag. We now default to Info level instead of Warning.

Added:

 - SkyFi discovery support.  Now SkySafari can auto-discover AlpacaScope as if it
    was a SkyFi device.
 - Windows output logging now supports a cleaner/colorized look

Fixed:

 - Debug logging for Nexstar no longer prints the whole buffer
 - Alpaca discovery comparison logic is now more accurate.
 - Fixed crash when received invalid messages < 3 bytes in LX200 mode

## v0.0.4 - 2021-01-16

Added:

 - Info about viruses and how to build on Windows 
 - Git workflow for building & testing
 - Add Linux-ARM binary for RasPi
 - Add support for Alpaca discovery via: --alpaca-host auto
 - Add support for choosing from between multiple telescopes via --telescope-id

Fixed:

 - Formatting of changelog

## v0.0.3 - 2020-12-30

Changed:

 - --alpaca-ip is now --alpaca-host since hostname/FQDN is supported

Added:

 - Refactor code to make it easier for other projects to use
 - Add support for additional Nexstar commands
 - More unit tests
 - Add LX200 support

Fixed:

 - Bug with enum for Alpaca Axis control
 - Fix bug in Nexstar location (ABCDEFGH) math
 - Fix issue building amd64 binaries looking like arm64

## v0.0.2 - 2020-12-23

Fixed: 

 - Properly close Nexstar client TCP sockets that are no longer in use.

## v0.0.1 - 2020-12-22

Initial release

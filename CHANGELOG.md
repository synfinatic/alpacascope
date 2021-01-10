# AlpacaScope Changelog

## Unreleased

Added:

 - Info about viruses and how to build on Windows 
 - Git workflow for building & testing

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

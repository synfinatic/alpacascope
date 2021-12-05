# AlpacaScope Changelog

## v2.2.3 - 2021-12-05

Changed:

 - LX200 mode now defaults to disabling "high precision" mode #51
 - Both GUI and CLI now support toggling high precision mode to be on or 
    off by default on startup so that you can enable high precision 
    by default to replicate the old behavior.

## v2.2.2 - 2021-11-11

Added:
 - Include a Windows CLI verion of the binary for people who don't have OpenGL
    compatible hardware.
 - Update Fyne to 2.1.1

## v2.2.1 - 2021-10-03

Changed:

 - ClientId is now randomized per Alpaca protocol spec
 - Will attempt to Connect to the telescope if Alpaca reports the connected
    state is false.
 - Update to Fyne 2.1.0
 - Use SO_REUSEPORT to avoid bugs. #44

## v2.2.0 - 2021-08-02

Added:

 - SkySafari's "Stop" action will disable tracking which for some mounts
    will prevent future goto's until tracking is re-enabled.  AlpacaScope 
    will now optionally check for this situation and re-enable tracking 
    for goto's to be successful. #41

Changed:

 - Use more correct "Alpaca Mount" terminology instead of "ASCOM Remote" #38

Fixed:

 - "Reset Settings" no longer deletes your saved settings. If you want to
	reset your saved settings, please re-save your settings. #40

## v2.1.0 - 2021-07-31

Added:

 - Ability to save/read configuration settings across executions.

Changed:

 - Selecting mount type is only supported with NexStar protocol 
 - Updated makefile targets
 - Updated readme
 - Add improved docs for configuration

Fixed:

 - Moved the Icon.png to the standard Fyne location for better compatibility
    with Fyne tooling

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

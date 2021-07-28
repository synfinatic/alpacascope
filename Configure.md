# Configuring AlpacaScope

This how-to will walk you through setting up ASCOM Remote, AlpacaScope and 
SkySafari for the most common use cases.

## What you need

 * Your telescope mount configured in [ASCOM](https://ascom-standards.org).
 * Download & install [ASCOM Remote Server](
    https://github.com/ASCOMInitiative/ASCOMRemote/releases) to convert ASCOM
    to Alpaca.
 * [AlpacaScope](https://github.com/synfinatic/alpacascope/releases) downloaded
    for your computer to convert Alpaca to NexStar/LX200.
 * SkySafari or some other astronomy software which speaks Celestron NexStar or
    Meade LX200 protocol over TCP/IP.

This document will assume you have your telescope mount already setup via ASCOM
and walk you through configuring ASCOM Remote, AlpacaScope and SkySafari. If 
your mount is supported natively via Alpaca you can skip to step #7!

## Let's Go!

1. Connect your telescope mount to your PC as usual and start up whatever ASCOM 
    driver you normally use. (For the purposes of this documentation, I will be
    using the ASCOM Telescope Simulator.)
2. Download and install [ASCOM Remote Server](
    https://github.com/ASCOMInitiative/ASCOMRemote/releases).
3. Start ASCOM Remote Server and click "Setup"
	![](https://user-images.githubusercontent.com/1075352/127172229-8550cf99-98f1-4b5b-8eaf-fa48f05fec7f.png)
4. Configure your telescope as shown:
   ![](https://user-images.githubusercontent.com/1075352/127172241-aca0e0ea-620d-4135-a5fd-542cd58449a7.png)
5. Switch to the "Server Configuration" tab:
   ![](https://user-images.githubusercontent.com/1075352/127172246-3e040da9-e6b6-4c1a-8424-4a25223ee666.png)
6. Choose which network interface to listen on:
  * `localhost` - More secure. Use if running AlpacaScope on the same computer
  * `+` - Less secure. Use if running AlpacaScope on a different computer
    (remote observatory, etc).
  ![](https://user-images.githubusercontent.com/1075352/127172250-e9376f78-77fc-4d09-9826-51c06ba28632.png)
7. Start AlpacaScope:
  * `Telescope Protocol` - This must match your setting in SkySafari.
  * `Mount Type` - Only for NexStar. Set to your mount type.
  * `Listen IP` - Default is to listen on all interfaces so you can use
        SkySafari on a remote device (iPad, Android tablet, etc). Change if
        you wish.
  * `Listen Port` - 4030 is the default port SkySafari uses.
  * `Auto Discover ASCOM Remote` - Keep checked unless you have a complex
        network setup (then you can manually set the values).
  * `ASCOM Telescope ID` - Must match the telescope ID set in Step #4.
  * `Start AlpacaScope Services` - Click this when settings are correct!
     You should then see the following messages in the box below showing it is
     ready for SkySafari/etc.
  ![](https://user-images.githubusercontent.com/1075352/127172252-be00b81b-75ae-47f5-a944-389272d8227b.png)
8. SkySafari configuration.  Since AlpacaScope emulates a 
    [SkyFi](https://skysafariastronomy.com/skyfi-3-professional-astronomy-telescope-control.html) 
    so you should not need to configure the IP or Port information.  Note: If
    SkySafari is running on the same machine as AlpacaScope, SkySafari may
    report the IP address is "Not Found"- this should not be a problem if
    you are using the default port "4030".
  ![](https://user-images.githubusercontent.com/1075352/127172258-1ff8cb2c-989e-41c9-b60d-eefe3d709b97.png)
9. When ready, click "Connect" in SkySafari and enjoy!

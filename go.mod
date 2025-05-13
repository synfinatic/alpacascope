module github.com/synfinatic/alpacascope

// go version must be <= https://github.com/fyne-io/fyne-cross-images/blob/main/base/Dockerfile#L3
go 1.23.5

// when you update the Fyne version, also update the Makefile target
require (
	fyne.io/fyne/v2 v2.6.1
	github.com/alecthomas/kong v1.8.1
	github.com/go-resty/resty/v2 v2.16.5
	github.com/libp2p/go-reuseport v0.4.0
	github.com/mattn/go-colorable v0.1.14
	github.com/relvacode/iso8601 v1.6.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	golang.org/x/image v0.24.0 // indirect; security
	golang.org/x/sys v0.30.0
	golang.org/x/text v0.22.0 // indirect; security
)

require github.com/davecgh/go-spew v1.1.1

require (
	fyne.io/systray v1.11.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/fredbi/uri v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fyne-io/gl-js v0.1.0 // indirect
	github.com/fyne-io/glfw-js v0.2.0 // indirect
	github.com/fyne-io/image v0.1.1 // indirect
	github.com/fyne-io/oksvg v0.1.0 // indirect
	github.com/go-gl/gl v0.0.0-20231021071112-07e5d0ea2e71 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20240506104042-037f3cc74f2a // indirect
	github.com/go-text/render v0.2.0 // indirect
	github.com/go-text/typesetting v0.2.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/hack-pad/go-indexeddb v0.3.2 // indirect
	github.com/hack-pad/safejs v0.1.0 // indirect
	github.com/jeandeaual/go-locale v0.0.0-20241217141322-fcc2cadd6f08 // indirect
	github.com/jsummers/gobmp v0.0.0-20230614200233-a9de23ed2e25 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.5.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rymdport/portal v0.4.1 // indirect
	github.com/srwiley/oksvg v0.0.0-20221011165216-be6e8873101c // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/yuin/goldmark v1.7.8 // indirect
	golang.org/x/net v0.36.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

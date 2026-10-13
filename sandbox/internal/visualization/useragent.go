package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// The device, browser and operating system of a view are worked out from its
// User-Agent by the substring rules below, matched in lower case and in order:
// the first rule that matches wins. They are written by hand — no parser of a
// third party — and they only need to be right for the browsers people use:
// anything else is Other.

// botMarkers mark the user agent of a program rather than a person: a search
// engine, a monitor, a script, a headless browser. A view made by one is
// recorded with device bot and left out of every metric unless bots are asked
// for.
var botMarkers = []string{
	"bot", "crawler", "spider", "slurp", "curl", "wget", "python-requests", "python-urllib", "go-http-client",
	"headless", "httpclient", "okhttp", "java/", "libwww", "scrapy", "phantomjs", "facebookexternalhit",
	"lighthouse", "pingdom", "uptime", "monitor", "preview", "feedfetcher", "axios", "node-fetch",
}

// Agent is what a User-Agent says about the client that sent it.
type Agent struct {
	// Device is one of Devices.
	Device string
	// Browser is one of Browsers.
	Browser string
	// Os is one of OperatingSystems.
	Os string
}

// ParseUserAgent works out the device, browser and operating system of
// userAgent. An empty one is a bot: every browser sends one.
func ParseUserAgent(sandbox *api.Sandbox, userAgent string) Agent {
	text := sandbox.Deps.StringsDeps.ToLower(userAgent)
	agent := Agent{Device: device(sandbox, text), Browser: browser(sandbox, text), Os: operatingSystem(sandbox, text)}
	if agent.Device == DeviceBot {
		agent.Browser = "Other"
	}
	return agent
}

// IsBot tells whether userAgent is a program's rather than a person's.
func IsBot(sandbox *api.Sandbox, userAgent string) bool {
	return device(sandbox, sandbox.Deps.StringsDeps.ToLower(userAgent)) == DeviceBot
}

// device is the device the lower-case user agent text names.
func device(sandbox *api.Sandbox, text string) string {
	strings := sandbox.Deps.StringsDeps
	if strings.TrimSpace(text) == "" {
		return DeviceBot
	}
	for _, marker := range botMarkers {
		if strings.Contains(text, marker) {
			return DeviceBot
		}
	}
	switch {
	case strings.Contains(text, "ipad") || strings.Contains(text, "tablet") || strings.Contains(text, "kindle") ||
		strings.Contains(text, "silk/") || strings.Contains(text, "playbook") ||
		strings.Contains(text, "android") && !strings.Contains(text, "mobile"):
		return DeviceTablet
	case strings.Contains(text, "mobi") || strings.Contains(text, "iphone") || strings.Contains(text, "ipod") ||
		strings.Contains(text, "android") || strings.Contains(text, "windows phone"):
		return DeviceMobile
	case strings.Contains(text, "windows") || strings.Contains(text, "macintosh") || strings.Contains(text, "mac os x") ||
		strings.Contains(text, "x11") || strings.Contains(text, "linux") || strings.Contains(text, "cros"):
		return DeviceDesktop
	}
	return DeviceOther
}

// browser is the browser the lower-case user agent text names. Each browser
// built on another names it too — Edge and Opera say Chrome, Chrome says
// Safari — so the most specific is tried first.
func browser(sandbox *api.Sandbox, text string) string {
	strings := sandbox.Deps.StringsDeps
	switch {
	case strings.Contains(text, "edg/") || strings.Contains(text, "edge/") || strings.Contains(text, "edga/") || strings.Contains(text, "edgios/"):
		return "Edge"
	case strings.Contains(text, "opr/") || strings.Contains(text, "opera") || strings.Contains(text, "opios/"):
		return "Opera"
	case strings.Contains(text, "samsungbrowser/"):
		return "Samsung Internet"
	case strings.Contains(text, "firefox/") || strings.Contains(text, "fxios/"):
		return "Firefox"
	case strings.Contains(text, "chrome/") || strings.Contains(text, "crios/") || strings.Contains(text, "chromium/"):
		return "Chrome"
	case strings.Contains(text, "safari/") && (strings.Contains(text, "version/") || strings.Contains(text, "mobile/")):
		return "Safari"
	}
	return "Other"
}

// operatingSystem is the operating system the lower-case user agent text
// names. iOS says it is "like Mac OS X" and Android that it is Linux, so both
// are tried before the system they borrow the name of.
func operatingSystem(sandbox *api.Sandbox, text string) string {
	strings := sandbox.Deps.StringsDeps
	switch {
	case strings.Contains(text, "iphone") || strings.Contains(text, "ipad") || strings.Contains(text, "ipod"):
		return "iOS"
	case strings.Contains(text, "android"):
		return "Android"
	case strings.Contains(text, "cros"):
		return "ChromeOS"
	case strings.Contains(text, "windows"):
		return "Windows"
	case strings.Contains(text, "macintosh") || strings.Contains(text, "mac os x"):
		return "macOS"
	case strings.Contains(text, "linux") || strings.Contains(text, "x11"):
		return "Linux"
	}
	return "Other"
}

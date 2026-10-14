package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/geoipdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
)

// The country of a view is looked up locally, in an IP-to-country database
// held in memory: the visitor's ip is never sent anywhere. The database is the
// DB-IP "IP to Country Lite" embedded at EmbeddedGeoip — free, CC BY 4.0,
// which asks for the credit GeoipCredit wherever its data is shown — unless
// start-server's --visualization-geoip names another .mmdb.

const (
	// EmbeddedGeoip is the country database compiled into the binary, as the
	// embedded asset tree names it.
	EmbeddedGeoip = "visualization/dbip-country-lite.mmdb"
	// GeoipSourceEmbedded is the source of the embedded database.
	GeoipSourceEmbedded = "embedded"
	// GeoipCredit is the credit the license of the embedded database asks
	// for, linked to GeoipCreditUrl.
	GeoipCredit = "IP Geolocation by DB-IP"
	// GeoipCreditUrl is where GeoipCredit links to.
	GeoipCreditUrl = "https://db-ip.com"
)

// GeoipInfo is the country database this process looks views up in.
type GeoipInfo struct {
	// Open is true once OpenGeoip opened one: before, every view is
	// UnknownCountry.
	Open bool
	// Source is GeoipSourceEmbedded, or the file --visualization-geoip named.
	Source string
	// BuiltAt is when the database was built, in Unix seconds; 0 unknown.
	BuiltAt int64
	// Kind is the database type its metadata names ("DBIP-Country-Lite").
	Kind string
}

// geoipState is the database OpenGeoip opened, guarded by geoipLock.
type geoipState struct {
	info     GeoipInfo
	database geoipdeps.Database
}

// geoip is the database this process looks views up in.
var geoip geoipState

// geoipLock is held while geoip is read or written.
var geoipLock = make(chan struct{}, 1)

// OpenGeoip opens the country database views are looked up in: the file
// sandbox.Config.VisualizationGeoipFile names, else the one embedded at
// EmbeddedGeoip. A file that cannot be read, or that is no country database,
// is an error naming it, and the database open before stays.
func OpenGeoip(sandbox *api.Sandbox) error {
	file := sandbox.Config.VisualizationGeoipFile
	source := GeoipSourceEmbedded
	var content []byte
	var err error
	if file == "" {
		content, err = sandbox.Deps.EmbedDeps.ReadFile(EmbeddedGeoip)
		if err != nil {
			return sandbox.Deps.StdDeps.Errorf("the embedded country database %s cannot be read: %s", EmbeddedGeoip, err.Error())
		}
	} else {
		source = file
		content, err = sandbox.Deps.IoDeps.ReadFile(file)
		if err != nil {
			return sandbox.Deps.StdDeps.Errorf("the country database %s cannot be read: %s", file, err.Error())
		}
	}
	database, err := sandbox.Deps.GeoipDeps.Open(content)
	if err != nil {
		return sandbox.Deps.StdDeps.Errorf("%s is not a country database (.mmdb): %s", source, err.Error())
	}

	geoipLock <- struct{}{}
	defer func() { <-geoipLock }()
	geoip = geoipState{
		info:     GeoipInfo{Open: true, Source: source, BuiltAt: database.BuiltAt, Kind: database.Kind},
		database: database,
	}
	return nil
}

// CurrentGeoip is the country database this process looks views up in.
func CurrentGeoip(sandbox *api.Sandbox) GeoipInfo {
	geoipLock <- struct{}{}
	defer func() { <-geoipLock }()
	return geoip.info
}

// Country is the ISO 3166-1 alpha-2 code of the country ip is located in, as
// the open database says, or UnknownCountry: for an ip that is not one, a
// private or loopback one, one the database does not hold, or before any
// database was opened.
func Country(sandbox *api.Sandbox, ip string) string {
	if Unrouted(sandbox, ip) {
		return UnknownCountry
	}
	address, _ := backofficehttp.ParseIp(sandbox, ip)
	geoipLock <- struct{}{}
	state := geoip
	<-geoipLock
	if !state.info.Open {
		return UnknownCountry
	}
	code := sandbox.Deps.StringsDeps.ToUpper(state.database.Country(backofficehttp.FormatIp(sandbox, address)))
	if !isCountryCode(code) {
		return UnknownCountry
	}
	return code
}

// IsCountryCode tells whether code is two letters A to Z: an ISO 3166-1
// alpha-2 code, or UnknownCountry.
func IsCountryCode(sandbox *api.Sandbox, code string) bool {
	return isCountryCode(code)
}

// isCountryCode tells whether code is two letters A to Z.
func isCountryCode(code string) bool {
	return len(code) == 2 && code[0] >= 'A' && code[0] <= 'Z' && code[1] >= 'A' && code[1] <= 'Z'
}

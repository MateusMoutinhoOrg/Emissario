package maxmindgeoip

import (
	"net"
	"strings"

	"github.com/oschwald/maxminddb-golang"

	geoipdeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/geoipdeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// record is the part of a country record this adapter reads, in every layout
// a country database spells it: country.iso_code in DB-IP Lite and MaxMind
// GeoLite2-Country, registered_country.iso_code when an ip carries no country
// of its own (an anycast network), and the flat country_code of IPinfo Lite.
type record struct {
	Country struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	RegisteredCountry struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"registered_country"`
	CountryCode string `maxminddb:"country_code"`
}

// open fills geoipdeps.Contract.Open: content read as a MaxMind DB, kept in
// memory as it is, with its build date and type out of its metadata.
func open(content []byte) (geoipdeps.Database, error) {
	reader, err := maxminddb.FromBytes(content)
	if err != nil {
		return geoipdeps.Database{}, err
	}
	return geoipdeps.Database{
		Country: func(ip string) string {
			return country(reader, ip)
		},
		BuiltAt: int64(reader.Metadata.BuildEpoch),
		Kind:    reader.Metadata.DatabaseType,
	}, nil
}

// country is the ISO code reader holds for ip, upper case, "" when ip is not
// an ip or the database knows no country for it.
func country(reader *maxminddb.Reader, ip string) string {
	address := net.ParseIP(ip)
	if address == nil {
		return ""
	}
	var found record
	if err := reader.Lookup(address, &found); err != nil {
		return ""
	}
	for _, code := range []string{found.Country.IsoCode, found.CountryCode, found.RegisteredCountry.IsoCode} {
		if code != "" {
			return strings.ToUpper(code)
		}
	}
	return ""
}

// Bind fills deps.Deps.GeoipDeps with github.com/oschwald/maxminddb-golang.
func Bind(deps *deps.Deps) {
	deps.GeoipDeps = geoipdeps.Contract{
		Open: func(content []byte) (geoipdeps.Database, error) {
			return open(content)
		},
	}
}

package geoipdeps

// This package is the sandbox's *copy* of the api an IP-to-country database
// library exposes — the same mechanic as hashdeps and jwtdeps, for the same
// reason: decoding a database file format is a third-party library, which may
// not appear inside the sandbox. The contract is restated here, and the
// adapter — which lives outside the sandbox — is what fills it.
//
// The database is handed over as bytes, never as a path: where they come from
// — a file compiled into the binary, a file on disk — is the sandbox's to
// say, so the library never touches the filesystem. Every lookup is local:
// the ip is never sent anywhere.
// Only builtin types cross this boundary — no `net.IP`, no type of the
// concrete library.

// Contract is the IP-to-country library injected whole as the Deps.GeoipDeps
// field. It is named after the capability, not after whoever reads it, so any
// part of the project that needs the country of an ip reuses it.
type Contract struct {
	// Open reads a country database — a MaxMind DB (.mmdb) whose records
	// carry an ISO 3166-1 alpha-2 country code: DB-IP Lite, MaxMind
	// GeoLite2-Country, IPinfo Lite — out of content, whole, and answers it
	// ready for lookups. content is kept, so it is never changed afterwards.
	// The error reports content that is no such database.
	Open func(content []byte) (Database, error)
}

// Database is one opened country database.
type Database struct {
	// Country answers the ISO 3166-1 alpha-2 code, upper case ("BR", "US"),
	// of the country ip is located in, or "" when ip is not an ip or the
	// database holds no country for it. It is safe to call from several
	// goroutines at once.
	Country func(ip string) string
	// BuiltAt is when the database was built, in seconds since the Unix
	// epoch, as its metadata says; 0 when it says nothing.
	BuiltAt int64
	// Kind is the database type its metadata names ("DBIP-Country-Lite",
	// "GeoLite2-Country").
	Kind string
}

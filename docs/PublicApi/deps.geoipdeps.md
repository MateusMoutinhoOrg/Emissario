# `deps.GeoipDeps`

`sandbox/deps/geoipdeps`

## `Contract`

Contract is the IP-to-country library injected whole as the Deps.GeoipDeps field. It is named after the capability, not after whoever reads it, so any part of the project that needs the country of an ip reuses it.

| Field | Type | Description |
| --- | --- | --- |
| `Open` | `func(content []byte) (Database, error)` | Open reads a country database — a MaxMind DB (.mmdb) whose records carry an ISO 3166-1 alpha-2 country code: DB-IP Lite, MaxMind GeoLite2-Country, IPinfo Lite — out of content, whole, and answers it ready for lookups. content is kept, so it is never changed afterwards. The error reports content that is no such database. |

## `Database`

Database is one opened country database.

| Field | Type | Description |
| --- | --- | --- |
| `Country` | `func(ip string) string` | Country answers the ISO 3166-1 alpha-2 code, upper case ("BR", "US"), of the country ip is located in, or "" when ip is not an ip or the database holds no country for it. It is safe to call from several goroutines at once. |
| `BuiltAt` | `int64` | BuiltAt is when the database was built, in seconds since the Unix epoch, as its metadata says; 0 when it says nothing. |
| `Kind` | `string` | Kind is the database type its metadata names ("DBIP-Country-Lite", "GeoLite2-Country"). |

[every contract](doc.md)

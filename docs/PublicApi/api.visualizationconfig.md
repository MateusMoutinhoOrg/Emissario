# `sandbox/api/visualizationconfig.go`

## `VisualizationConfig`

VisualizationConfig is the part of the Config the visualization mechanic reads — the page views of the public website, recorded by the front route and shown in the backoffice at /admin/visualization — embedded in api.Config so each field is read as sandbox.Config.<Field>. The visualization-start-server middleware fills it in front of start-server, before the first request is served.

| Field | Type | Description |
| --- | --- | --- |
| `VisualizationGeoipFile` | `string` | VisualizationGeoipFile is the country database the ip of a view is looked up in — any MaxMind DB (.mmdb) of countries: a newer DB-IP Lite, MaxMind GeoLite2-Country, IPinfo Lite — as start-server's --visualization-geoip names it. Empty, the DB-IP Lite database embedded at assets/visualization/dbip-country-lite.mmdb is used. |

[every contract](doc.md)

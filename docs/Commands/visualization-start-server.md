# `visualization-start-server`

Open the country database and start recording the page views of the public website in front of start-server

A middleware: it runs on rung 52, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It opens the country database every page view of the public website is looked up in - the DB-IP Lite one embedded in the binary, or the .mmdb --visualization-geoip names - and starts the two goroutines of the visualization: the writer, which turns each view the front route queued into a record of visualization-db (holding them while a backup is taken or restored), and the retention cleanup, which removes every hour the views older than the retention a root set at /admin/root/set-visualization-settings (90 days by default). The ip of a view is looked up locally: it is never sent anywhere. Behind a reverse proxy, run with --allow-x-forwarded-for so the ip is the visitor's.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--visualization-geoip` | string |  | a country database (.mmdb) to look the ip of each page view up in instead of the embedded DB-IP Lite: a newer DB-IP Lite, MaxMind GeoLite2-Country, IPinfo Lite; a file that does not open stops the server at start | — |

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

```bash
emissario start-server --visualization-geoip GeoLite2-Country.mmdb
```

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)

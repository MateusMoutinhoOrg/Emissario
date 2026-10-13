package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// exportColumns are the columns of an export, as its first row names them.
var exportColumns = []string{"time", "path", "status", "ip", "country", "device", "browser", "os", "referrer", "user_agent"}

// ExportName is the file an export of filter is downloaded as:
// visualization-views-<from>-<to>.csv.
func ExportName(sandbox *api.Sandbox, filter Filter) string {
	return "visualization-views-" + filter.From + "-" + filter.To + ".csv"
}

// ExportCsv is the views of selection as CSV — newest first, MaxExportRows at
// most, every ip whole: only a root exports — with the time in the server's
// time zone.
func ExportCsv(sandbox *api.Sandbox, selection Selection) []byte {
	lines := []string{csvLine(sandbox, exportColumns)}
	for index, view := range selection.Views {
		if index >= MaxExportRows {
			break
		}
		lines = append(lines, csvLine(sandbox, []string{
			FormatLocal(sandbox, view.At, TimeLayout),
			view.Path,
			sandbox.Deps.StringsDeps.FormatInt(view.Status, 10),
			view.Ip,
			view.Country,
			view.Device,
			view.Browser,
			view.Os,
			view.Referrer,
			view.UserAgent,
		}))
	}
	return []byte(sandbox.Deps.StringsDeps.Join(lines, "\r\n") + "\r\n")
}

// csvLine is cells as one line of CSV.
func csvLine(sandbox *api.Sandbox, cells []string) string {
	escaped := []string{}
	for _, cell := range cells {
		escaped = append(escaped, csvCell(sandbox, cell))
	}
	return sandbox.Deps.StringsDeps.Join(escaped, ",")
}

// csvCell is cell as one field of CSV: quoted when it holds a comma, a quote
// or a line break, its quotes doubled. A cell a spreadsheet would run as a
// formula — starting with = + - @, a tab or a carriage return — is led by a
// quote mark, so a user agent or a path sent by a visitor is never run.
func csvCell(sandbox *api.Sandbox, cell string) string {
	strings := sandbox.Deps.StringsDeps
	if cell != "" && strings.ContainsAny(cell[:1], "=+-@\t\r") {
		cell = "'" + cell
	}
	if strings.ContainsAny(cell, ",\"\r\n") {
		return "\"" + strings.ReplaceAll(cell, "\"", "\"\"") + "\""
	}
	return cell
}

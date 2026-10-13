package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// VisualizationSettingsPage is what backoffice/visualization_settings.html is
// rendered with.
type VisualizationSettingsPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	// Action is where the settings form posts, PurgeAction the Delete views
	// one.
	Action      string
	PurgeAction string
	// Error is shown above the form, "" for none.
	Error string
	// Notice is the outcome of the last save or delete.
	Notice Notice
	// Enabled, Retention and IgnoredIps fill the form.
	Enabled    bool
	Retention  string
	IgnoredIps string
	// GeoipSource names the country database in use, GeoipBuilt its build
	// date; GeoipEmbedded shows its credit.
	GeoipSource   string
	GeoipKind     string
	GeoipBuilt    string
	GeoipEmbedded bool
	CreditUrl     string
	// Stored is how many views are stored, Oldest the oldest one.
	Stored string
	Oldest string
	// Dropped is how many views the full queue dropped since the server
	// started.
	Dropped string
	// Running is true when the writer runs in this process.
	Running bool
	// YourIp is the ip of the request, which the Add my ip button adds.
	YourIp string
	// Today fills the date of the Delete views form.
	Today string
	Zone  string
}

// VisualizationSettingsForm is what the settings form holds: what was saved,
// or what was sent and refused.
type VisualizationSettingsForm struct {
	Enabled    bool
	Retention  string
	IgnoredIps string
}

// VisualizationSettingsFormOf is the form filled with settings.
func VisualizationSettingsFormOf(sandbox *api.Sandbox, settings visualization.Settings) VisualizationSettingsForm {
	form := VisualizationSettingsForm{Enabled: settings.Enabled, IgnoredIps: sandbox.Deps.StringsDeps.Join(settings.IgnoredIps, "\n")}
	if settings.RetentionDays > 0 {
		form.Retention = sandbox.Deps.StringsDeps.FormatInt(settings.RetentionDays, 10)
	}
	return form
}

// visualizationNotice is how the settings page words a notice code, removed
// being how many views a delete removed.
func visualizationNotice(sandbox *api.Sandbox, code string, removed int) Notice {
	switch code {
	case visualization.NoticeSaved:
		return Notice{Text: "Settings saved.", Kind: "ok"}
	case visualization.NoticePurged:
		return Notice{Text: plural(sandbox, removed, "view", "views") + " deleted.", Kind: "ok"}
	}
	return Notice{}
}

// RenderVisualizationSettingsPage answers, under status, the settings page
// for user, the form filled with form and message above it — or the sentence
// of notice — beside the country database in use and what is stored. ip is
// the ip of the request.
func RenderVisualizationSettingsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, form VisualizationSettingsForm, notice string, removed int, message string, ip string) error {
	stored, oldest, err := visualization.StoredViews(sandbox)
	if err != nil {
		return err
	}
	geoip := visualization.CurrentGeoip(sandbox)
	viewer := viewerOf(sandbox, user)
	page := VisualizationSettingsPage{
		Viewer:        viewer,
		Tabs:          visualizationTabs(sandbox, visualization.SettingsPath, visualization.Filter{}, false, viewer.IsRoot),
		Action:        visualization.SettingsPath,
		PurgeAction:   visualization.PurgePath,
		Error:         message,
		Notice:        visualizationNotice(sandbox, notice, removed),
		Enabled:       form.Enabled,
		Retention:     form.Retention,
		IgnoredIps:    form.IgnoredIps,
		GeoipSource:   "Not opened in this process: it is opened by start-server",
		GeoipKind:     geoip.Kind,
		GeoipEmbedded: geoip.Source == visualization.GeoipSourceEmbedded,
		CreditUrl:     visualization.GeoipCreditUrl,
		Stored:        formatCount(sandbox, stored),
		Oldest:        formatInstant(sandbox, oldest),
		Dropped:       formatCount(sandbox, int(visualization.Dropped(sandbox))),
		Running:       visualization.Running(sandbox),
		YourIp:        ip,
		Today:         visualization.Today(sandbox),
		Zone:          visualization.ZoneName(sandbox),
	}
	switch {
	case geoip.Source == visualization.GeoipSourceEmbedded:
		page.GeoipSource = "Embedded DB-IP IP to Country Lite"
	case geoip.Open:
		page.GeoipSource = geoip.Source
	}
	page.GeoipBuilt = "—"
	if geoip.BuiltAt > 0 {
		page.GeoipBuilt = visualization.FormatLocal(sandbox, geoip.BuiltAt, visualization.DayLayout)
	}
	return RenderHTML(sandbox, response, status, "backoffice/visualization_settings.html", page)
}

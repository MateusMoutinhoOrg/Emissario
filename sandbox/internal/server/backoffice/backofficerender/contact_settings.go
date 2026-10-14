package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The page of the contact requests only a root reaches, under /admin/root:
// the settings of the public route, beside what it holds, and the danger zone
// that deletes every contact request before a date.

// ContactSettingsForm is what the settings form is filled with.
type ContactSettingsForm struct {
	Enabled       bool
	MaxBytes      string
	RatePerMinute string
	IpDailyLimit  string
}

// ContactSettingsFormOf is the form filled with settings; a limit of 0 is
// blank.
func ContactSettingsFormOf(sandbox *api.Sandbox, settings contact.Settings) ContactSettingsForm {
	blank := func(value int64) string {
		if value <= 0 {
			return ""
		}
		return sandbox.Deps.StringsDeps.FormatInt(value, 10)
	}
	return ContactSettingsForm{
		Enabled:       settings.Enabled,
		MaxBytes:      blank(settings.MaxBytes),
		RatePerMinute: blank(settings.RatePerMinute),
		IpDailyLimit:  blank(settings.IpDailyLimit),
	}
}

// ContactSettingsPage is what backoffice/contact_settings.html is rendered
// with.
type ContactSettingsPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	Action string
	Notice Notice
	Error  string
	Form   ContactSettingsForm
	// The state of the contact requests.
	Running bool
	Stored  string
	Inbox   string
	Dropped string
	// Endpoint is the public route, MaxMaxBytes the largest body a root may
	// set.
	Endpoint    string
	MaxMaxBytes int
	// RemoveAction is where the danger zone deletes every contact request
	// before a date.
	RemoveAction string
	Today        string
}

// RenderContactSettingsPage answers, under status, the settings form filled
// with form and with message above it — or the sentence of notice, count
// being how many contact requests it touched — beside the state of the
// contact requests.
func RenderContactSettingsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, form ContactSettingsForm, notice string, count int, message string) error {
	state, err := contact.CurrentState(sandbox)
	if err != nil {
		return err
	}
	page := ContactSettingsPage{
		Viewer:       viewerOf(sandbox, user),
		Tabs:         contactTabs(sandbox, contact.SettingsPath, nil, true),
		Action:       contact.SettingsPath,
		Notice:       contactNotice(sandbox, notice, count),
		Error:        message,
		Form:         form,
		Running:      state.Running,
		Stored:       formatCount(sandbox, state.Stored),
		Inbox:        formatCount(sandbox, state.Inbox),
		Dropped:      formatCount(sandbox, int(state.Dropped)),
		Endpoint:     contact.PublicPath,
		MaxMaxBytes:  contact.MaxMaxBytes,
		RemoveAction: contact.RemovePath,
		Today:        visualization.Today(sandbox),
	}
	return RenderHTML(sandbox, response, status, "backoffice/contact_settings.html", page)
}

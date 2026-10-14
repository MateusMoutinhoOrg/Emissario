package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// The pages of the agent only a root reaches, under /admin/root: the settings,
// the urls the agent can call and the agent source. The api key and the
// values of the headers are write-only: no page shows them.

// AgentSettingsForm is what the settings form is filled with.
type AgentSettingsForm struct {
	Enabled       bool
	BaseUrl       string
	Model         string
	MaxTokens     string
	DailyLimit    string
	IpDailyLimit  string
	RatePerMinute string
}

// AgentSettingsFormOf is the form filled with settings; a limit of 0 is blank.
func AgentSettingsFormOf(sandbox *api.Sandbox, settings agent.Settings) AgentSettingsForm {
	blank := func(value int64) string {
		if value <= 0 {
			return ""
		}
		return sandbox.Deps.StringsDeps.FormatInt(value, 10)
	}
	return AgentSettingsForm{
		Enabled:       settings.Enabled,
		BaseUrl:       settings.BaseUrl,
		Model:         settings.Model,
		MaxTokens:     blank(settings.MaxTokens),
		DailyLimit:    blank(settings.DailyLimit),
		IpDailyLimit:  blank(settings.IpDailyLimit),
		RatePerMinute: blank(settings.RatePerMinute),
	}
}

// AgentSettingsPage is what backoffice/agent_settings.html is rendered with.
type AgentSettingsPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	Action string
	Notice Notice
	Error  string
	Form   AgentSettingsForm
	// KeyHint is how the key in use is shown, "" for none; KeyFromEnv names
	// the variable it comes from.
	KeyHint    string
	KeySaved   bool
	KeyFromEnv bool
	KeyEnv     string
	Answering  bool
	Problem    string
	Readme     bool
	Summary    bool
	Running    bool
	Dropped    string
	Zone       string
	Endpoint   string
	Retention  int
}

// RenderAgentSettingsPage answers, under status, the settings form filled with
// form and with message above it — or the sentence of notice — beside the
// state of the agent.
func RenderAgentSettingsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, form AgentSettingsForm, notice string, message string) error {
	state, settings, err := agent.CurrentStatus(sandbox)
	if err != nil {
		return err
	}
	page := AgentSettingsPage{
		Viewer:     viewerOf(sandbox, user),
		Tabs:       agentTabs(sandbox, agent.SettingsPath, true),
		Action:     agent.SettingsPath,
		Notice:     agentNotice(sandbox, notice, -1),
		Error:      message,
		Form:       form,
		KeyHint:    state.KeyHint,
		KeySaved:   settings.ApiKey != "",
		KeyFromEnv: state.KeyFromEnv,
		KeyEnv:     agent.ApiKeyEnv(sandbox),
		Answering:  state.Answering,
		Problem:    state.Problem,
		Readme:     state.Readme,
		Summary:    state.Summary,
		Running:    state.Running,
		Dropped:    formatCount(sandbox, int(state.Dropped)),
		Zone:       visualization.ZoneName(sandbox),
		Retention:  agent.UsageRetentionDays,
	}
	if settings.BaseUrl != "" {
		page.Endpoint = agent.CompletionsUrl(sandbox, settings)
	}
	return RenderHTML(sandbox, response, status, "backoffice/agent_settings.html", page)
}

// AgentUrlLine is one url of the list.
type AgentUrlLine struct {
	Id          string
	Name        string
	Method      string
	Url         string
	Description string
	Headers     string
	EditHref    string
}

// AgentUrlsPage is what backoffice/agent_urls.html is rendered with.
type AgentUrlsPage struct {
	Viewer       Viewer
	Tabs         []VisualizationLink
	Notice       Notice
	Error        string
	AddHref      string
	RemoveAction string
	Rows         []AgentUrlLine
	Max          int
}

// RenderAgentUrlsPage answers, under status, every url the agent can call,
// with message above them — or the sentence of notice.
func RenderAgentUrlsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, notice string, message string) error {
	urls, err := agent.ListUrls(sandbox)
	if err != nil {
		return err
	}
	page := AgentUrlsPage{
		Viewer:       viewerOf(sandbox, user),
		Tabs:         agentTabs(sandbox, agent.UrlsPath, true),
		Notice:       agentNotice(sandbox, notice, -1),
		Error:        message,
		AddHref:      agent.AddUrlPath,
		RemoveAction: agent.RemoveUrlPath,
		Max:          agent.MaxUrls,
	}
	for _, url := range urls {
		id := sandbox.Deps.StringsDeps.FormatInt(url.Id, 10)
		page.Rows = append(page.Rows, AgentUrlLine{
			Id:          id,
			Name:        url.Name,
			Method:      url.Method,
			Url:         url.Url,
			Description: url.Description,
			Headers:     sandbox.Deps.StringsDeps.Join(agent.HeaderNames(sandbox, url.Headers), ", "),
			EditHref:    agent.SetUrlPath + "?id=" + id,
		})
	}
	return RenderHTML(sandbox, response, status, "backoffice/agent_urls.html", page)
}

// AgentUrlForm is what the url form is filled with. The headers are never
// filled with what is saved: their names are shown beside.
type AgentUrlForm struct {
	Id          string
	Name        string
	Description string
	Method      string
	Url         string
	Headers     string
}

// AgentUrlFormOf is the form of the url record, its headers left blank.
func AgentUrlFormOf(sandbox *api.Sandbox, record agent_db.AgentUrlRecord) AgentUrlForm {
	return AgentUrlForm{
		Id:          sandbox.Deps.StringsDeps.FormatInt(record.Id, 10),
		Name:        record.Name,
		Description: record.Description,
		Method:      record.Method,
		Url:         record.Url,
	}
}

// AgentUrlFormPage is what backoffice/agent_url_form.html is rendered with.
type AgentUrlFormPage struct {
	Viewer Viewer
	Tabs   []VisualizationLink
	// Editing is true for a url saved, false for a new one.
	Editing bool
	Action  string
	Error   string
	Form    AgentUrlForm
	// SavedHeaders are the names of the headers saved, when editing.
	SavedHeaders string
	Methods      []LabeledOption
	Back         string
}

// LabeledOption is one option of a select.
type LabeledOption struct {
	Value    string
	Label    string
	Selected bool
}

// RenderAgentUrlFormPage answers, under status, the form that adds a url — or
// changes the one saved, savedHeaders the names of its headers — filled with
// form and with message above it.
func RenderAgentUrlFormPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, editing bool, form AgentUrlForm, savedHeaders string, message string) error {
	page := AgentUrlFormPage{
		Viewer:       viewerOf(sandbox, user),
		Tabs:         agentTabs(sandbox, agent.UrlsPath, true),
		Editing:      editing,
		Action:       agent.AddUrlPath,
		Error:        message,
		Form:         form,
		SavedHeaders: savedHeaders,
		Back:         agent.UrlsPath,
	}
	if editing {
		page.Action = agent.SetUrlPath
	}
	method := sandbox.Deps.StringsDeps.ToUpper(form.Method)
	if method == "" {
		method = "GET"
	}
	for _, value := range []string{"GET", "POST"} {
		page.Methods = append(page.Methods, LabeledOption{Value: value, Label: value, Selected: value == method})
	}
	return RenderHTML(sandbox, response, status, "backoffice/agent_url_form.html", page)
}

// AgentSourcePage is what backoffice/agent_source_form.html is rendered with:
// the form of the website source, for the agent source.
type AgentSourcePage struct {
	WebsiteSourceFormPage
	Tabs []VisualizationLink
	// Files is how many files the agent can read; Readme whether
	// README.md is among them.
	Files  string
	Readme bool
}

// RenderAgentSourcePage answers, under status, the form that sets the agent
// source, filled with settings and with message above it — or the sentence of
// notice — beside what its sync loop is doing and its last sync did.
func RenderAgentSourcePage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, settings websitesource.Settings, notice string, message string) error {
	form, err := sourceFormPage(sandbox, user, websitesource.AgentName, settings, notice, message)
	if err != nil {
		return err
	}
	files, _ := agent.InstructionFiles(sandbox)
	io := sandbox.Deps.IoDeps
	page := AgentSourcePage{
		WebsiteSourceFormPage: form,
		Tabs:                  agentTabs(sandbox, agent.SourcePath, true),
		Files:                 formatCount(sandbox, len(files)),
		Readme:                io.IsFile(io.Join(agent.InstructionsDir(sandbox), agent.ReadmeFile)),
	}
	return RenderHTML(sandbox, response, status, "backoffice/agent_source_form.html", page)
}

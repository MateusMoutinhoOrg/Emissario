package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// syncPath is where the Sync button of the website source page posts.
const syncPath = "/admin/root/sync-website-source"

// agentSyncPath is where the Sync button of the agent source page posts.
const agentSyncPath = "/admin/root/sync-agent-source"

// WebsiteSourceFormPage is what backoffice/website_source_form.html is
// rendered with: the form that sets a source — the website's — the Sync
// button, and what its sync loop is doing and its last sync did.
// AgentSourcePage embeds it for the agent source.
type WebsiteSourceFormPage struct {
	Viewer Viewer
	// Action is where the form posts.
	Action string
	// SyncAction is where the Sync button posts.
	SyncAction string
	// Error is shown above the form, "" for none.
	Error string
	// Notice is the outcome of the last save or sync asked for, "" for none.
	Notice string
	// NoticeIsError shows Notice as a refusal.
	NoticeIsError bool
	RepoUrl       string
	Branch        string
	SyncTime      string
	// Saved is true once the source was saved: Sync needs one.
	Saved bool
	// SiteDir is the folder of the checkout that is read: the website's
	// frontend/ folder, or the agent's instructions/ folder.
	SiteDir string
	// Running is true when the sync loop runs in this process.
	Running bool
	// Busy is true from the moment a sync is asked for until it ends: the page
	// refreshes itself meanwhile.
	Busy bool
	// Attempted is true once the loop finished a sync.
	Attempted bool
	// AttemptedAt is when the last finished sync started, in UTC.
	AttemptedAt string
	// Ok is true when the last finished sync passed.
	Ok bool
	// Commit is the short hash the checkout is at.
	Commit string
	// Output is why the last finished sync failed.
	Output string
}

// websiteSourceNotice is the sentence a notice code of the page reads as, and
// whether it is a refusal; "" for a code it does not know.
func websiteSourceNotice(sandbox *api.Sandbox, notice string) (string, bool) {
	switch notice {
	case websitesource.NoticeSaved:
		return "Saved. With a sync time set, a new repository or branch is synced within a few seconds.", false
	case websitesource.OutcomeRequested:
		return "Sync started: this page refreshes until it ends.", false
	case websitesource.OutcomeNoSource:
		return "Save a repository before syncing.", true
	case websitesource.OutcomeNotRunning:
		return "The sync loop does not run in this process: it runs in the one start-server started.", true
	}
	return "", false
}

// RenderWebsiteSourcePage answers, under status, the form that sets the
// website source, filled with settings and with message above it — or the
// sentence of notice — beside what the sync loop of this process is doing and
// its last sync did.
func RenderWebsiteSourcePage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, settings websitesource.Settings, notice string, message string) error {
	page, err := sourceFormPage(sandbox, user, websitesource.FrontendName, settings, notice, message)
	if err != nil {
		return err
	}
	return RenderHTML(sandbox, response, status, "backoffice/website_source_form.html", page)
}

// sourceFormPage is the form of the source name filled with settings, with
// message above it — or the sentence of notice — and what its sync loop is
// doing and its last sync did.
func sourceFormPage(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord, name string, settings websitesource.Settings, notice string, message string) (WebsiteSourceFormPage, error) {
	_, saved, err := websitesource.Read(sandbox, name)
	if err != nil {
		return WebsiteSourceFormPage{}, err
	}
	last := websitesource.CurrentStatus(sandbox, name)
	siteDir, sync := websitesource.SiteDir(sandbox), syncPath
	if name == websitesource.AgentName {
		siteDir = sandbox.Deps.IoDeps.Join(websitesource.RepoDir(sandbox, name), "instructions")
		sync = agentSyncPath
	}
	page := WebsiteSourceFormPage{
		Viewer:     viewerOf(sandbox, user),
		Action:     websitesource.PageOf(sandbox, name),
		SyncAction: sync,
		Error:      message,
		RepoUrl:    settings.RepoUrl,
		Branch:     settings.Branch,
		SyncTime:   settings.SyncTime,
		Saved:      saved,
		SiteDir:    siteDir,
		Running:    last.Running,
		Busy:       last.Pending || last.Syncing,
		Attempted:  last.AttemptedAt > 0,
		Ok:         last.Ok,
		Commit:     last.Commit,
		Output:     last.Output,
	}
	page.Notice, page.NoticeIsError = websiteSourceNotice(sandbox, notice)
	if page.Attempted {
		page.AttemptedAt = sandbox.Deps.TimeDeps.FormatUnix(last.AttemptedAt, "2006-01-02 15:04:05 UTC")
	}
	return page, nil
}

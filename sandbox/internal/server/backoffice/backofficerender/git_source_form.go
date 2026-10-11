package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/gitsource"
)

// GitSourceFormPage is what backoffice/git_source_form.html is rendered with:
// the form that sets the git source of the public website, and what the last
// sync did.
type GitSourceFormPage struct {
	Viewer Viewer
	// Action is where the form posts.
	Action string
	// Error is shown above the form, "" for none.
	Error string
	// Notice is the outcome of the last save, "" for none.
	Notice   string
	RepoUrl  string
	Branch   string
	SyncTime string
	// RepoDir is the folder the repository is cloned into, and the website
	// served from.
	RepoDir string
	// Running is true when the sync loop runs in this process.
	Running bool
	// Attempted is true once the loop tried a sync.
	Attempted bool
	// AttemptedAt is when the last sync started, in UTC.
	AttemptedAt string
	// Ok is true when the last sync passed.
	Ok bool
	// Commit is the short hash the checkout is at.
	Commit string
	// Output is why the last sync failed.
	Output string
}

// gitSourceNotice is the sentence a notice code of the page reads as, "" for
// one it does not know.
func gitSourceNotice(sandbox *api.Sandbox, notice string) string {
	switch notice {
	case gitsource.NoticeSaved:
		return "Saved. The repository is synced within a few seconds when a sync time is set."
	}
	return ""
}

// RenderGitSourcePage answers, under status, the form that sets the git
// source, filled with settings and with message above it — or the sentence of
// notice — beside what the last sync of this process did.
func RenderGitSourcePage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, settings gitsource.Settings, notice string, message string) error {
	last := gitsource.CurrentStatus(sandbox)
	page := GitSourceFormPage{
		Viewer:    viewerOf(sandbox, user),
		Action:    gitsource.PagePath,
		Error:     message,
		Notice:    gitSourceNotice(sandbox, notice),
		RepoUrl:   settings.RepoUrl,
		Branch:    settings.Branch,
		SyncTime:  settings.SyncTime,
		RepoDir:   gitsource.RepoDir(sandbox),
		Running:   last.Running,
		Attempted: last.AttemptedAt > 0,
		Ok:        last.Ok,
		Commit:    last.Commit,
		Output:    last.Output,
	}
	if page.Attempted {
		page.AttemptedAt = sandbox.Deps.TimeDeps.FormatUnix(last.AttemptedAt, "2006-01-02 15:04:05 UTC")
	}
	return RenderHTML(sandbox, response, status, "backoffice/git_source_form.html", page)
}

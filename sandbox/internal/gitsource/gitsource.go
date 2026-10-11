package gitsource

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/emissario_db"
)

// The git source is the repository the public website is served from. A root
// sets its url, branch and sync time in the backoffice; they are kept as the
// frontend record of the source table of emissario-db. The goroutine Start
// runs clones the repository into <FontsDir>/frontend and pulls the branch
// every sync time (sync.go), and the site is that checkout itself (site.go):
// / answers the index.html at the root of the repository. git runs with whatever credentials the user running the
// server has set up — an ssh key, a credential helper — and none is ever
// asked for, stored or put on a command line here.

const (
	// FrontendName is the name of the source record the public website is
	// served from, and of the folder under FontsDir its repository is cloned
	// into.
	FrontendName = "frontend"
	// PagePath is the backoffice page that sets the git source.
	PagePath = "/admin/root/set-git-source"
	// DefaultBranch fills the branch of a form no root saved yet.
	DefaultBranch = "main"
	// DefaultSyncTime fills the sync time of a form no root saved yet.
	DefaultSyncTime = "10m"
	// MinSyncMilliseconds is the shortest sync time accepted, so a typo never
	// pulls a remote every second.
	MinSyncMilliseconds = 10 * 1000
	// MaxSyncMilliseconds is the longest sync time accepted: 30 days.
	MaxSyncMilliseconds = 30 * 24 * 60 * 60 * 1000
	// MaxRepoUrlLength is the longest repository url accepted.
	MaxRepoUrlLength = 2048
	// MaxBranchLength is the longest branch name accepted.
	MaxBranchLength = 255
)

const (
	// NoticeSaved follows a git source saved.
	NoticeSaved = "saved"
)

// Settings is what a root sets in the backoffice: the repository to clone,
// the branch to pull, and how often to pull it — "" for never.
type Settings struct {
	RepoUrl  string
	Branch   string
	SyncTime string
}

// sourceLock is held while the frontend record is read or written, so the
// sync loop never reads the url of one save with the branch of another.
var sourceLock = make(chan struct{}, 1)

// lockSource takes sourceLock, waiting for it.
func lockSource(sandbox *api.Sandbox) {
	sourceLock <- struct{}{}
}

// unlockSource gives sourceLock back.
func unlockSource(sandbox *api.Sandbox) {
	<-sourceLock
}

// Defaults are the settings a form no root saved yet is filled with.
func Defaults(sandbox *api.Sandbox) Settings {
	return Settings{Branch: DefaultBranch, SyncTime: DefaultSyncTime}
}

// RepoDir is the folder the frontend repository is cloned into, and the one
// the public website is served from.
func RepoDir(sandbox *api.Sandbox) string {
	return sandbox.Deps.IoDeps.Join(fontsDir(sandbox), FrontendName)
}

// fontsDir is sandbox.Config.FontsDir, api.DefaultFontsDir when it was left
// empty, so a source is never cloned straight into the working directory.
func fontsDir(sandbox *api.Sandbox) string {
	if sandbox.Config.FontsDir == "" {
		return api.DefaultFontsDir
	}
	return sandbox.Config.FontsDir
}

// Read answers the settings of the frontend source, false when no root saved
// any yet.
func Read(sandbox *api.Sandbox) (Settings, bool, error) {
	lockSource(sandbox)
	defer unlockSource(sandbox)
	record, ok, err := findFrontend(sandbox)
	if err != nil || !ok {
		return Settings{}, false, err
	}
	return Settings{RepoUrl: record.Repourl, Branch: record.Branch, SyncTime: record.SyncTime}, true, nil
}

// findFrontend is the source record named FrontendName — the first one, were
// there ever two — false when there is none. The caller holds sourceLock.
func findFrontend(sandbox *api.Sandbox) (emissario_db.SourceRecord, bool, error) {
	records, err := emissario_db.New(sandbox).ListSources(emissario_db.SourceFilter{NameEquals: FrontendName})
	if err != nil {
		return emissario_db.SourceRecord{}, false, err
	}
	if len(records) == 0 {
		return emissario_db.SourceRecord{}, false, nil
	}
	return records[0], true, nil
}

// Save validates settings and writes them on the frontend source record,
// creating it the first time, then asks the sync loop for a sync on its next
// tick. A refused value comes back as message, a sentence for the form, and
// nothing is written; err is a store that failed.
func Save(sandbox *api.Sandbox, settings Settings) (string, error) {
	settings, message := Validate(sandbox, settings)
	if message != "" {
		return message, nil
	}

	lockSource(sandbox)
	err := writeFrontend(sandbox, settings)
	unlockSource(sandbox)
	if err != nil {
		return "", err
	}
	RequestSync(sandbox)
	return "", nil
}

// writeFrontend writes settings on the frontend source record, adding it when
// there is none. The caller holds sourceLock.
func writeFrontend(sandbox *api.Sandbox, settings Settings) error {
	db := emissario_db.New(sandbox)
	record, ok, err := findFrontend(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		_, err := db.AddSource(emissario_db.SourceInput{
			Name:     FrontendName,
			Repourl:  settings.RepoUrl,
			Branch:   settings.Branch,
			SyncTime: settings.SyncTime,
		})
		return err
	}
	if err := db.SetSourceRepourl(record.Id, settings.RepoUrl); err != nil {
		return err
	}
	if err := db.SetSourceBranch(record.Id, settings.Branch); err != nil {
		return err
	}
	return db.SetSourceSyncTime(record.Id, settings.SyncTime)
}

// Validate answers settings trimmed, and the sentence the form shows for the
// first value it refuses — "" when every one passed.
func Validate(sandbox *api.Sandbox, settings Settings) (Settings, string) {
	strings := sandbox.Deps.StringsDeps
	settings.RepoUrl = strings.TrimSpace(settings.RepoUrl)
	settings.Branch = strings.TrimSpace(settings.Branch)
	settings.SyncTime = strings.TrimSpace(settings.SyncTime)

	if message := validateRepoUrl(sandbox, settings.RepoUrl); message != "" {
		return settings, message
	}
	if message := validateBranch(sandbox, settings.Branch); message != "" {
		return settings, message
	}
	if settings.SyncTime != "" {
		milliseconds, ok := SyncMilliseconds(sandbox, settings.SyncTime)
		if !ok {
			return settings, "The sync time reads as a number and a unit: 30s, 10m, 1h or 1h30m — or blank to stop syncing."
		}
		if milliseconds < MinSyncMilliseconds || milliseconds > MaxSyncMilliseconds {
			return settings, "The sync time is at least 10s and at most 720h (30 days)."
		}
	}
	return settings, ""
}

// validateRepoUrl is why url cannot be cloned, "" when it can. Only the
// transports git trusts by default pass — https, http, ssh, git, file and the
// scp-like user@host:path — never a remote helper (ext::, fd::), which runs a
// command, nor anything git could read as an option. A url carrying a
// password or a token is refused too: credentials are set up for the user
// running the server, never kept in the database.
func validateRepoUrl(sandbox *api.Sandbox, url string) string {
	strings := sandbox.Deps.StringsDeps
	if url == "" {
		return "The repository URL is required."
	}
	if len(url) > MaxRepoUrlLength {
		return "The repository URL is longer than 2048 characters."
	}
	if hasSpaceOrControl(sandbox, url) {
		return "The repository URL holds a space or a control character."
	}
	if strings.HasPrefix(url, "-") || strings.Contains(url, "::") {
		return "The repository URL is not one git clones safely: use https://, ssh://, git://, file:// or user@host:path."
	}
	if matches(sandbox, `^https?://[^/]*@`, url) || matches(sandbox, `^ssh://[^/@]*:[^/@]*@`, url) {
		return "Put no password or token in the repository URL: set up the credentials for the user that runs the server (an ssh key, a git credential helper)."
	}
	for _, scheme := range []string{"https://", "http://", "ssh://", "git://", "file://"} {
		if strings.HasPrefix(url, scheme) && len(url) > len(scheme) {
			return ""
		}
	}
	if matches(sandbox, `^[A-Za-z0-9._-]+@[A-Za-z0-9.-]+:[^/].*$`, url) {
		return ""
	}
	return "The repository URL starts with https://, http://, ssh://, git://, file:// or is user@host:path."
}

// validateBranch is why branch cannot be pulled, "" when it can: a name of
// letters, digits and . _ / -, never starting with - . or /, and holding none
// of the sequences git refuses in a branch name.
func validateBranch(sandbox *api.Sandbox, branch string) string {
	strings := sandbox.Deps.StringsDeps
	if branch == "" {
		return "The branch is required."
	}
	if len(branch) > MaxBranchLength {
		return "The branch is longer than 255 characters."
	}
	if !matches(sandbox, `^[A-Za-z0-9_][A-Za-z0-9._/-]*$`, branch) ||
		strings.Contains(branch, "..") || strings.Contains(branch, "//") || strings.Contains(branch, "/.") ||
		strings.HasSuffix(branch, "/") || strings.HasSuffix(branch, ".") || strings.HasSuffix(branch, ".lock") {
		return "The branch is a name of letters, digits and . _ / - such as main or release/v2."
	}
	return ""
}

// SyncMilliseconds reads a sync time — one or more of a number followed by
// s, m or h, as 30s, 10m or 1h30m — in milliseconds, and reports false on
// anything else, "" included.
func SyncMilliseconds(sandbox *api.Sandbox, text string) (int64, bool) {
	units := map[rune]int64{'s': 1000, 'm': 60 * 1000, 'h': 60 * 60 * 1000}
	total := int64(0)
	number := int64(0)
	digits := 0
	for _, letter := range text {
		if letter >= '0' && letter <= '9' {
			number = number*10 + int64(letter-'0')
			digits++
			if digits > 9 {
				return 0, false
			}
			continue
		}
		unit, ok := units[letter]
		if !ok || digits == 0 {
			return 0, false
		}
		total += number * unit
		number = 0
		digits = 0
	}
	if digits != 0 || text == "" {
		return 0, false
	}
	return total, true
}

// hasSpaceOrControl tells whether text holds a space, a tab or any other
// control character.
func hasSpaceOrControl(sandbox *api.Sandbox, text string) bool {
	for _, letter := range text {
		if letter <= ' ' || letter == 0x7f {
			return true
		}
	}
	return false
}

// matches tells whether text matches the regular expression pattern; a
// pattern that does not compile matches nothing.
func matches(sandbox *api.Sandbox, pattern string, text string) bool {
	ok, err := sandbox.Deps.StringsDeps.MatchPattern(pattern, text)
	return err == nil && ok
}

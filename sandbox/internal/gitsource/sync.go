package gitsource

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/rundeps"
)

// tickMilliseconds is how long the sync loop waits between two looks at the
// settings: a save, a new url or a new branch is picked up within it.
const tickMilliseconds = 2000

// maxOutputLength is how much of what git wrote a failed sync keeps for the
// backoffice page, its tail, where git says why.
const maxOutputLength = 4000

// gitConfig is passed to every git call ahead of its command: symlinks are
// checked out as plain files, so no file of the repository can point the
// website at a file of the server; remote helpers, which run commands, stay
// refused; and an http transfer stalled for a minute is given up on.
var gitConfig = []string{
	"-c", "core.symlinks=false",
	"-c", "protocol.ext.allow=never",
	"-c", "http.lowSpeedLimit=1000",
	"-c", "http.lowSpeedTime=60",
}

// Status is what the last sync of this process did, for the backoffice page.
type Status struct {
	// Running is true once Start ran in this process.
	Running bool
	// AttemptedAt is when the last sync started, in seconds since the Unix
	// epoch; 0 before the first.
	AttemptedAt int64
	// Ok is true when the last sync left the checkout at the branch's head.
	Ok bool
	// Commit is the short hash the checkout is at after the last sync that
	// passed.
	Commit string
	// Output is why the last sync failed: the error, then the tail of what
	// git wrote. "" when it passed.
	Output string
}

// status is the Status CurrentStatus answers, guarded by statusLock.
var status Status

// statusLock is held while status is read or written.
var statusLock = make(chan struct{}, 1)

// requested holds a token when a save asked for a sync on the next tick.
var requested = make(chan struct{}, 1)

// started holds a token once Start ran, so a second call starts nothing.
var started = make(chan struct{}, 1)

// CurrentStatus is a copy of what the last sync of this process did.
func CurrentStatus(sandbox *api.Sandbox) Status {
	statusLock <- struct{}{}
	defer func() { <-statusLock }()
	return status
}

// setStatus records what a sync did.
func setStatus(sandbox *api.Sandbox, next Status) {
	statusLock <- struct{}{}
	defer func() { <-statusLock }()
	status = next
}

// RequestSync asks the sync loop for a sync on its next tick, whatever the
// sync time; one request pending is as good as several.
func RequestSync(sandbox *api.Sandbox) {
	select {
	case requested <- struct{}{}:
	default:
	}
}

// takeRequest reports whether a sync was asked for since the last tick, and
// forgets the request.
func takeRequest(sandbox *api.Sandbox) bool {
	select {
	case <-requested:
		return true
	default:
		return false
	}
}

// Start runs the sync loop in a goroutine of its own, once per process: it
// lives as long as the process does and stops with it. Every tick it reads the
// settings of the frontend source and, when a sync time is set, syncs the
// repository at once on the first tick, again every sync time, and whenever
// the url or the branch changed or a save asked for it. With no sync time set
// it does nothing; a sync time set later is picked up on the next tick, with
// no restart.
func Start(sandbox *api.Sandbox) {
	select {
	case started <- struct{}{}:
	default:
		return
	}
	setStatus(sandbox, Status{Running: true})
	go loop(sandbox)
}

// loop is the body of the goroutine Start runs. It never returns.
func loop(sandbox *api.Sandbox) {
	synced := Settings{}
	syncedAt := int64(0)
	for {
		forced := takeRequest(sandbox)
		settings, ok, err := Read(sandbox)
		if err == nil && ok {
			interval, timed := SyncMilliseconds(sandbox, settings.SyncTime)
			now := sandbox.Deps.StdDeps.Now() / 1000000
			if timed && (forced || syncedAt == 0 || settings != synced || now-syncedAt >= interval) {
				syncAndReport(sandbox, settings)
				synced = settings
				syncedAt = now
			}
		}
		sandbox.Deps.SleepDeps.Sleep(tickMilliseconds)
	}
}

// syncAndReport runs Sync, records what it did for the backoffice page and
// says so on stderr.
func syncAndReport(sandbox *api.Sandbox, settings Settings) {
	next := Status{Running: true, AttemptedAt: sandbox.Deps.StdDeps.Now() / 1000000000}
	commit, output, err := Sync(sandbox, settings)
	if err != nil {
		next.Output = tail(sandbox, err.Error()+"\n"+output)
		sandbox.Deps.StdDeps.Eprintf("warning: the git source %s could not be synced from %s (%s): %s\n", FrontendName, settings.RepoUrl, settings.Branch, err.Error())
	} else {
		next.Ok = true
		next.Commit = commit
		sandbox.Deps.StdDeps.Logf("git source %s synced from %s (%s) at %s\n", FrontendName, settings.RepoUrl, settings.Branch, commit)
	}
	setStatus(sandbox, next)
}

// Sync leaves RepoDir a checkout of the head of settings.Branch of
// settings.RepoUrl, and answers the short hash it is at. A checkout already
// there is pointed at settings.RepoUrl, fetches the branch and checks it out
// over whatever it held, so a url or a branch that fails to fetch leaves the
// site as it was. Anything else at RepoDir — a folder that is no checkout, or
// one git cannot read — is removed and the repository cloned. On a failure,
// output is what git wrote.
func Sync(sandbox *api.Sandbox, settings Settings) (string, string, error) {
	io := sandbox.Deps.IoDeps
	strings := sandbox.Deps.StringsDeps
	dir := RepoDir(sandbox)

	if io.IsDir(io.Join(dir, ".git")) {
		origin, err := git(sandbox, dir, "remote", "get-url", "origin")
		if err == nil {
			if strings.TrimSpace(origin) != settings.RepoUrl {
				output, err := git(sandbox, dir, "remote", "set-url", "origin", "--", settings.RepoUrl)
				if err != nil {
					return "", output, err
				}
			}
			return pull(sandbox, dir, settings)
		}
	}
	return clone(sandbox, dir, settings)
}

// clone removes whatever is at dir and clones the head of settings.Branch
// into it, a single commit deep.
func clone(sandbox *api.Sandbox, dir string, settings Settings) (string, string, error) {
	io := sandbox.Deps.IoDeps
	if io.Exists(dir) {
		io.RemoveDir(dir)
	}
	io.CreateDir(io.Dir(dir))
	output, err := git(sandbox, "", "clone", "--depth", "1", "--single-branch", "--branch", settings.Branch,
		"--config", "core.symlinks=false", "--", settings.RepoUrl, dir)
	if err != nil {
		io.RemoveDir(dir)
		return "", output, err
	}
	return head(sandbox, dir)
}

// pull fetches the head of settings.Branch into the checkout at dir and checks
// it out by force: a branch rewritten upstream, or a switch to another
// branch, lands the same way as a plain update.
func pull(sandbox *api.Sandbox, dir string, settings Settings) (string, string, error) {
	remote := "refs/remotes/origin/" + settings.Branch
	output, err := git(sandbox, dir, "fetch", "--depth", "1", "origin", "+refs/heads/"+settings.Branch+":"+remote)
	if err != nil {
		return "", output, err
	}
	output, err = git(sandbox, dir, "checkout", "--force", "-B", settings.Branch, remote)
	if err != nil {
		return "", output, err
	}
	return head(sandbox, dir)
}

// head is the short hash the checkout at dir is at.
func head(sandbox *api.Sandbox, dir string) (string, string, error) {
	output, err := git(sandbox, dir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", output, err
	}
	return sandbox.Deps.StringsDeps.TrimSpace(output), "", nil
}

// git runs git with args in dir — the working directory when "" — under
// gitConfig and gitEnv, and answers what it wrote. A git that exits non-zero
// is an error naming its command and status; one that could not start at all
// (git not installed) is the error Run gave.
func git(sandbox *api.Sandbox, dir string, args ...string) (string, error) {
	result, err := sandbox.Deps.RunDeps.Run(rundeps.RunProps{
		Dir:     dir,
		Program: "git",
		Args:    append(append([]string{}, gitConfig...), args...),
		Env:     gitEnv(sandbox),
	})
	if err != nil {
		return result.Output, err
	}
	if result.ExitCode != 0 {
		return result.Output, sandbox.Deps.StdDeps.Errorf("git %s exited with status %d", args[0], result.ExitCode)
	}
	return result.Output, nil
}

// gitEnv is what git runs under: it never prompts for a username or a
// password, and ssh never asks a question — a passphrase, an unknown host key —
// so a missing credential fails the sync instead of hanging the loop. ssh is
// left alone when the user running the server chose how git calls it
// (GIT_SSH_COMMAND, GIT_SSH or core.sshCommand).
func gitEnv(sandbox *api.Sandbox) []string {
	env := []string{"GIT_TERMINAL_PROMPT=0"}
	if sandbox.Deps.EnvDeps.Getenv("GIT_SSH_COMMAND") != "" || sandbox.Deps.EnvDeps.Getenv("GIT_SSH") != "" {
		return env
	}
	result, err := sandbox.Deps.RunDeps.Run(rundeps.RunProps{Program: "git", Args: []string{"config", "--get", "core.sshCommand"}, Env: env})
	if err == nil && result.ExitCode == 0 && sandbox.Deps.StringsDeps.TrimSpace(result.Output) != "" {
		return env
	}
	return append(env, "GIT_SSH_COMMAND=ssh -o BatchMode=yes -o ConnectTimeout=30")
}

// tail is the last maxOutputLength bytes of text, trimmed.
func tail(sandbox *api.Sandbox, text string) string {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if len(text) > maxOutputLength {
		text = "…" + text[len(text)-maxOutputLength:]
	}
	return text
}

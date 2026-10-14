package websitesource

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/rundeps"
)

// tickMilliseconds is how long a sync loop waits between two looks at the
// settings: a new url, branch or sync time is picked up within it. A sync asked
// for through RequestSync wakes it at once.
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

const (
	// OutcomeRequested follows a sync asked for: the loop runs it at once.
	OutcomeRequested = "sync-requested"
	// OutcomeNoSource follows a sync asked for before the source was saved.
	OutcomeNoSource = "no-source"
	// OutcomeNotRunning follows a sync asked for in a process whose sync loop
	// does not run: it runs in the one start-server started.
	OutcomeNotRunning = "not-running"
)

// Status is what the sync loop of one source is doing in this process and
// what its last sync did, for the backoffice page.
type Status struct {
	// Running is true once Start ran in this process.
	Running bool
	// Pending is true from the moment a sync is asked for until it starts.
	Pending bool
	// Syncing is true while a sync runs.
	Syncing bool
	// AttemptedAt is when the last finished sync started, in seconds since
	// the Unix epoch; 0 before the first.
	AttemptedAt int64
	// Ok is true when the last finished sync left the checkout at the
	// branch's head.
	Ok bool
	// Commit is the short hash the checkout is at after the last sync that
	// passed; a failed sync leaves it, as it leaves the checkout.
	Commit string
	// Output is why the last finished sync failed: the error, then the tail
	// of what git wrote. "" when it passed.
	Output string
}

// source is the sync loop of one source: what it is doing and the token a
// sync asked for leaves it.
type source struct {
	// status is the Status CurrentStatus answers, guarded by statusLock.
	status Status
	// requested holds a token from the moment a sync is asked for until the
	// loop takes it.
	requested chan struct{}
}

// sources is the loop of each of Names. It is filled once, before any
// goroutine runs, and only the sources it points to change afterwards.
var sources = newSources()

// newSources is one source per name of Names.
func newSources() map[string]*source {
	made := map[string]*source{}
	for _, name := range Names {
		made[name] = &source{requested: make(chan struct{}, 1)}
	}
	return made
}

// sourceOf is the loop of the source name. A name that is not one of Names is
// a mistake of the caller: it panics rather than syncing nothing.
func sourceOf(sandbox *api.Sandbox, name string) *source {
	found, ok := sources[name]
	if !ok {
		panic("websitesource: no source is named " + name)
	}
	return found
}

// statusLock is held while the status of any source is read or written.
var statusLock = make(chan struct{}, 1)

// started holds a token once Start ran, so a second call starts nothing.
var started = make(chan struct{}, 1)

// CurrentStatus is a copy of what the sync loop of the source name is doing
// in this process and what its last sync did.
func CurrentStatus(sandbox *api.Sandbox, name string) Status {
	found := sourceOf(sandbox, name)
	statusLock <- struct{}{}
	defer func() { <-statusLock }()
	return found.status
}

// updateStatus runs change on the status of the source name under
// statusLock.
func updateStatus(sandbox *api.Sandbox, name string, change func(current *Status)) {
	found := sourceOf(sandbox, name)
	statusLock <- struct{}{}
	defer func() { <-statusLock }()
	change(&found.status)
}

// RequestSync asks the sync loop of the source name for a sync now — whatever
// the sync time, a blank one included, and even when the checkout is already
// at the head of its branch: the loop wakes at once, fetches the branch and
// checks it out again. It answers OutcomeRequested, or why nothing was asked:
// OutcomeNoSource before the source was saved, OutcomeNotRunning when no sync
// loop runs in this process. One request pending is as good as several.
func RequestSync(sandbox *api.Sandbox, name string) (string, error) {
	_, ok, err := Read(sandbox, name)
	if err != nil {
		return "", err
	}
	if !ok {
		return OutcomeNoSource, nil
	}
	requested := sourceOf(sandbox, name).requested
	outcome := OutcomeRequested
	updateStatus(sandbox, name, func(current *Status) {
		if !current.Running {
			outcome = OutcomeNotRunning
			return
		}
		current.Pending = true
		select {
		case requested <- struct{}{}:
		default:
		}
	})
	return outcome, nil
}

// takeRequest reports whether a sync of the source name was asked for and
// not taken yet, and takes it.
func takeRequest(sandbox *api.Sandbox, name string) bool {
	select {
	case <-sourceOf(sandbox, name).requested:
		return true
	default:
		return false
	}
}

// wait sleeps tickMilliseconds, or less when a sync of the source name is
// asked for meanwhile, and reports whether one was — taking it.
func wait(sandbox *api.Sandbox, name string) bool {
	woke := make(chan struct{}, 1)
	go func() {
		sandbox.Deps.SleepDeps.Sleep(tickMilliseconds)
		woke <- struct{}{}
	}()
	select {
	case <-sourceOf(sandbox, name).requested:
		return true
	case <-woke:
		return takeRequest(sandbox, name)
	}
}

// Start runs the sync loop of every source of Names, each in a goroutine of
// its own, once per process: they live as long as the process does and stop
// with it, and a slow clone of one never holds the others. Every tick a loop
// reads the settings of its source and, when a sync time is set, syncs the
// repository at once on the first tick, again every sync time, and whenever
// the url, the branch or the sync time changed. With no sync time set it syncs
// nothing on its own; a sync time set later is picked up on the next tick,
// with no restart. A sync asked for through RequestSync runs at once in
// either case. Two syncs of one source never run at once.
func Start(sandbox *api.Sandbox) {
	select {
	case started <- struct{}{}:
	default:
		return
	}
	for _, name := range Names {
		updateStatus(sandbox, name, func(current *Status) { current.Running = true })
		go loop(sandbox, name)
	}
}

// loop is the body of the goroutine Start runs for the source name. It never
// returns.
func loop(sandbox *api.Sandbox, name string) {
	synced := Settings{}
	syncedAt := int64(0)
	forced := takeRequest(sandbox, name)
	for {
		settings, ok, err := Read(sandbox, name)
		if err == nil && ok {
			interval, timed := SyncMilliseconds(sandbox, settings.SyncTime)
			now := sandbox.Deps.StdDeps.Now() / 1000000
			if forced || timed && (syncedAt == 0 || settings != synced || now-syncedAt >= interval) {
				syncAndReport(sandbox, name, settings)
				synced = settings
				syncedAt = now
				forced = false
			}
		}
		if forced {
			updateStatus(sandbox, name, func(current *Status) { current.Pending = false })
		}
		forced = wait(sandbox, name)
	}
}

// syncAndReport runs Sync for the source name, records what it did for the
// backoffice page and says so on stderr.
func syncAndReport(sandbox *api.Sandbox, name string, settings Settings) {
	attemptedAt := sandbox.Deps.StdDeps.Now() / 1000000000
	updateStatus(sandbox, name, func(current *Status) {
		current.Pending = false
		current.Syncing = true
	})
	commit, output, err := Sync(sandbox, name, settings)
	updateStatus(sandbox, name, func(current *Status) {
		current.Syncing = false
		current.AttemptedAt = attemptedAt
		current.Ok = err == nil
		current.Output = ""
		if err != nil {
			current.Output = tail(sandbox, err.Error()+"\n"+output)
		} else {
			current.Commit = commit
		}
	})
	if err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the %s %s could not be synced from %s (%s): %s\n", LabelOf(sandbox, name), name, settings.RepoUrl, settings.Branch, err.Error())
		return
	}
	sandbox.Deps.StdDeps.Logf("%s %s synced from %s (%s) at %s\n", LabelOf(sandbox, name), name, settings.RepoUrl, settings.Branch, commit)
}

// Sync leaves the RepoDir of the source name a checkout of the head of
// settings.Branch of settings.RepoUrl, and answers the short hash it is at. A checkout already
// there is pointed at settings.RepoUrl, fetches the branch and checks it out
// over whatever it held, so a url or a branch that fails to fetch leaves the
// site as it was. Anything else at RepoDir — a folder that is no checkout, or
// one git cannot read — is removed and the repository cloned. On a failure,
// output is what git wrote.
func Sync(sandbox *api.Sandbox, name string, settings Settings) (string, string, error) {
	io := sandbox.Deps.IoDeps
	strings := sandbox.Deps.StringsDeps
	dir := RepoDir(sandbox, name)

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

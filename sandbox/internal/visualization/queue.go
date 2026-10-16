package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// A view never slows the website down nor breaks it: the front route only
// hands a Hit to Enqueue, which puts it on queue — or drops it, the queue
// being full — and returns. One goroutine, the writer, turns each Hit into a
// visualization-view; a write that fails is said on stderr, never to the
// visitor. While a backup is taken or restored the writer holds the queue, so
// no database is written under the snapshot, and writes it afterwards.

// Hit is one answer of the front route, as Enqueue is handed it.
type Hit struct {
	// At is when it was answered, in Unix seconds; 0 is now.
	At int64
	// Method is the method of the request, as the server read it — HEAD for
	// a HEAD answered as a GET.
	Method string
	// Path is the path requested.
	Path string
	// Status is the status it was answered with.
	Status int
	// ContentType is the Content-Type it was answered with.
	ContentType string
	// Ip is the client ip, props.ClientIp.
	Ip string
	// UserAgent is the User-Agent header.
	UserAgent string
	// Referer is the Referer header.
	Referer string
	// Host is the host the request was sent to.
	Host string
}

// holdMilliseconds is how long the writer and the retention wait between two
// looks at a backup being taken or restored.
const holdMilliseconds = 500

// queue holds the hits the writer has not written yet.
var queue = make(chan Hit, QueueSize)

// started holds a token once Start ran, so a second call starts nothing.
var started = make(chan struct{}, 1)

// dropped is how many hits found the queue full since the process started,
// guarded by droppedLock.
var dropped int64

// droppedLock is held while dropped is read or written.
var droppedLock = make(chan struct{}, 1)

// Running tells whether the writer runs in this process: it runs in the one
// start-server started.
func Running(sandbox *api.Sandbox) bool {
	return len(started) > 0
}

// Enqueue hands hit to the writer, when it is a view — Counted — and a writer
// runs in this process. It never waits: with the queue full, the view is
// dropped and counted in Dropped.
func Enqueue(sandbox *api.Sandbox, hit Hit) {
	if !Running(sandbox) || !Counted(sandbox, hit.Method, hit.Status, hit.ContentType, hit.Path) {
		return
	}
	if hit.At == 0 {
		hit.At = NowSeconds(sandbox)
	}
	select {
	case queue <- hit:
	default:
		droppedLock <- struct{}{}
		dropped++
		<-droppedLock
	}
}

// Dropped is how many views found the queue full since the process started.
func Dropped(sandbox *api.Sandbox) int64 {
	droppedLock <- struct{}{}
	defer func() { <-droppedLock }()
	return dropped
}

// Start runs the writer and the retention cleanup, each in a goroutine of its
// own, once per process: they live as long as the process does and stop with
// it.
func Start(sandbox *api.Sandbox) {
	select {
	case started <- struct{}{}:
	default:
		return
	}
	go write(sandbox)
	go retain(sandbox)
}

// write is the body of the writer goroutine. It never returns.
func write(sandbox *api.Sandbox) {
	for hit := range queue {
		holdForBackups(sandbox)
		if err := store(sandbox, hit); err != nil {
			sandbox.Deps.StdDeps.Eprintf("warning: the view of %s could not be recorded: %s\n", truncate(hit.Path, 128), err.Error())
		}
	}
}

// holdForBackups waits while a backup is taken or restored.
func holdForBackups(sandbox *api.Sandbox) {
	for snapshots.Creating(sandbox) || snapshots.Restoring(sandbox) {
		sandbox.Deps.SleepDeps.Sleep(holdMilliseconds)
	}
}

// store writes hit as a visualization-view, unless tracking is off or its ip
// is ignored. The country and the ignored ips read the raw ip, in memory;
// the view keeps it as the IP mode says — hashed by default, with no user
// agent but its device, browser and os — and the full user agent only in
// the full mode.
func store(sandbox *api.Sandbox, hit Hit) error {
	settings := currentSettings(sandbox)
	if !settings.Enabled || Ignored(sandbox, settings, hit.Ip) {
		return nil
	}
	privacySettings, err := privacy.CurrentSettings(sandbox)
	if err != nil {
		return err
	}
	ip, err := privacy.StoredIp(sandbox, privacySettings, hit.Ip)
	if err != nil {
		return err
	}
	userAgent := ""
	if privacySettings.IpMode == privacy.IpModeFull {
		userAgent = truncate(hit.UserAgent, MaxUserAgentLength)
	}
	agent := ParseUserAgent(sandbox, hit.UserAgent)
	_, err = visualization_db.New(sandbox).AddVisualizationView(visualization_db.VisualizationViewInput{
		At:        hit.At,
		Path:      CleanPath(sandbox, hit.Path),
		Page:      CanonicalPage(sandbox, hit.Path, hit.Status),
		Status:    int64(hit.Status),
		Ip:        ip,
		Country:   Country(sandbox, hit.Ip),
		Device:    agent.Device,
		Browser:   agent.Browser,
		Os:        agent.Os,
		UserAgent: userAgent,
		Referrer:  Referrer(sandbox, hit.Referer, hit.Host),
	})
	return err
}

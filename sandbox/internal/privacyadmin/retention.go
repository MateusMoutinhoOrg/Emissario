package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backup"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The retention cleanup runs every hour in one goroutine of start-server: it
// deletes the contact requests, views, agent summaries and access log rows
// past their retention, blanks the ip of the views, contact requests and
// agent usage past the IP retention — hashed or full — releases the holds
// whose end date passed, and removes the snapshots past the backup
// retention. A row on a legal hold is skipped, ip included, until its hold
// ends. Retention is counted from a row's at, in whole days of the server's
// time zone: with 90 days, a row of a day is kept while that day is one of
// the 90 before today, and today. It waits while a backup is taken or
// restored, as every writer does, and logs counts alone, never a value.

// cleanupMilliseconds is how long the cleanup waits between two runs.
const cleanupMilliseconds = 60 * 60 * 1000

// Retentions is every retention the cleanup goes by, in days, or
// privacy.Forever to keep that data until a root deletes it.
type Retentions struct {
	Contact   int64
	View      int64
	Summary   int64
	AccessLog int64
	Ip        int64
	Backup    int64
}

// CurrentRetentions reads the retentions saved in every database.
func CurrentRetentions(sandbox *api.Sandbox) (Retentions, error) {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return Retentions{}, err
	}
	contacts, err := contact.ReadSettings(sandbox)
	if err != nil {
		return Retentions{}, err
	}
	views, err := visualization.ReadSettings(sandbox)
	if err != nil {
		return Retentions{}, err
	}
	agentSettings, err := agent.ReadSettings(sandbox)
	if err != nil {
		return Retentions{}, err
	}
	return Retentions{
		Contact:   contacts.RetentionDays,
		View:      views.RetentionDays,
		Summary:   agentSettings.SummaryRetentionDays,
		AccessLog: settings.AccessLogRetentionDays,
		Ip:        settings.IpRetentionDays,
		Backup:    settings.BackupRetentionDays,
	}, nil
}

// Shortest is the shortest retention of the records, 0 when every one keeps
// its data forever: a snapshot older than it holds data the databases already
// deleted.
func (retentions Retentions) Shortest() int64 {
	shortest := privacy.Forever
	for _, days := range []int64{retentions.Contact, retentions.View, retentions.Summary, retentions.AccessLog, retentions.Ip} {
		if days != privacy.Forever && (shortest == privacy.Forever || days < shortest) {
			shortest = days
		}
	}
	return shortest
}

// Longest is the longest retention of the records, privacy.Forever when one
// keeps its data forever: what a backup retention past it keeps longer.
func (retentions Retentions) Longest() int64 {
	longest := int64(0)
	for _, days := range []int64{retentions.Contact, retentions.View, retentions.Summary, retentions.AccessLog, retentions.Ip} {
		if days == privacy.Forever {
			return privacy.Forever
		}
		if days > longest {
			longest = days
		}
	}
	return longest
}

// Cutoff is the instant, in Unix seconds, before which a row is past a
// retention of days: the start of the day days before today, in the server's
// time zone. 0 — nothing is past it — for privacy.Forever.
func Cutoff(sandbox *api.Sandbox, days int64) int64 {
	if days == privacy.Forever {
		return 0
	}
	start, ok := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, visualization.Today(sandbox), -int(days)))
	if !ok {
		return 0
	}
	return start
}

// Swept is what one cleanup deleted or blanked — or, counted without
// writing, would.
type Swept struct {
	Contacts   int
	Views      int
	Summaries  int
	AccessLog  int
	ContactIps int
	ViewIps    int
	UsageIps   int
	Snapshots  int
	Released   int
}

// Total is every row and snapshot it deleted, and every ip it blanked.
func (swept Swept) Total() int {
	return swept.Contacts + swept.Views + swept.Summaries + swept.AccessLog + swept.ContactIps + swept.ViewIps + swept.UsageIps + swept.Snapshots
}

// Sweep deletes what is past retentions, as the cleanup does — or, with
// apply false, writes nothing and counts what it would delete, by the same
// rules. A count with apply false treats as released every hold the cleanup
// would release first.
func Sweep(sandbox *api.Sandbox, retentions Retentions, apply bool) (Swept, error) {
	swept := Swept{}
	now := privacy.NowSeconds(sandbox)
	if apply {
		released, err := privacy.ReleaseExpired(sandbox, now)
		if err != nil {
			return swept, err
		}
		swept.Released = released
	}
	ipBefore := Cutoff(sandbox, retentions.Ip)

	heldContacts, err := privacy.HeldIds(sandbox, privacy.TableContactRequest, now)
	if err != nil {
		return swept, err
	}
	records, err := contact_db.ContactRequestRecords(sandbox, contact_db.New(sandbox))
	if err != nil {
		return swept, err
	}
	swept.Contacts, swept.ContactIps, err = sweepRecords(sandbox, records, Cutoff(sandbox, retentions.Contact), ipBefore, heldContacts, apply)
	if err != nil {
		return swept, err
	}

	records, err = visualization_db.VisualizationViewRecords(sandbox, visualization_db.New(sandbox))
	if err != nil {
		return swept, err
	}
	swept.Views, swept.ViewIps, err = sweepRecords(sandbox, records, Cutoff(sandbox, retentions.View), ipBefore, nil, apply)
	if err != nil {
		return swept, err
	}

	heldSummaries, err := privacy.HeldIds(sandbox, privacy.TableAgentSummary, now)
	if err != nil {
		return swept, err
	}
	records, err = agent_db.AgentSummaryRecords(sandbox, agent_db.New(sandbox))
	if err != nil {
		return swept, err
	}
	swept.Summaries, _, err = sweepRecords(sandbox, records, Cutoff(sandbox, retentions.Summary), 0, heldSummaries, apply)
	if err != nil {
		return swept, err
	}

	records, err = agent_db.AgentUsageRecords(sandbox, agent_db.New(sandbox))
	if err != nil {
		return swept, err
	}
	_, swept.UsageIps, err = sweepRecords(sandbox, records, 0, ipBefore, nil, apply)
	if err != nil {
		return swept, err
	}

	heldAccess, err := privacy.HeldIds(sandbox, privacy.TableAccessLog, now)
	if err != nil {
		return swept, err
	}
	records, err = privacy_db.AccessLogRecords(sandbox, privacy_db.New(sandbox))
	if err != nil {
		return swept, err
	}
	swept.AccessLog, _, err = sweepRecords(sandbox, records, Cutoff(sandbox, retentions.AccessLog)*1000, 0, heldAccess, apply)
	if err != nil {
		return swept, err
	}

	swept.Snapshots, err = sweepSnapshots(sandbox, Cutoff(sandbox, retentions.Backup), apply)
	return swept, err
}

// sweepRecords deletes every one of records whose at is before removeBefore,
// and blanks the ip of every other one whose at is before ipBefore — a bound
// of 0 is none — skipping the ids of held. With apply false it only counts.
// It answers how many it deleted and how many ips it blanked.
func sweepRecords(sandbox *api.Sandbox, records []databasedeps.Record, removeBefore int64, ipBefore int64, held map[int64]bool, apply bool) (int, int, error) {
	database := sandbox.Deps.OpinionatedAgnosDatabase
	removed, blanked := 0, 0
	for _, record := range records {
		if removeBefore == 0 && ipBefore == 0 {
			break
		}
		if held[record.ID] {
			continue
		}
		at, err := database.ReadInt(record, "at")
		if err != nil {
			return removed, blanked, err
		}
		if removeBefore > 0 && at < removeBefore {
			if apply {
				if err := database.Fail(record.Remove()); err != nil {
					return removed, blanked, err
				}
			}
			removed++
			continue
		}
		if ipBefore == 0 || at >= ipBefore {
			continue
		}
		ip, err := database.ReadString(record, "ip")
		if err != nil {
			return removed, blanked, err
		}
		if ip == "" {
			continue
		}
		if apply {
			if err := database.Fail(record.Update("ip", "")); err != nil {
				return removed, blanked, err
			}
		}
		blanked++
	}
	return removed, blanked, nil
}

// sweepSnapshots removes every snapshot taken before before — none when it
// is 0 — and answers how many; with apply false it only counts. A snapshot
// still being taken, open, or kept to put a failed restore back is left, and
// so is every one while another backup job runs: the next run removes them.
func sweepSnapshots(sandbox *api.Sandbox, before int64, apply bool) (int, error) {
	if before == 0 {
		return 0, nil
	}
	old, err := snapshotsBefore(sandbox, before)
	if err != nil {
		return 0, err
	}
	if !apply {
		return len(old), nil
	}
	removed := 0
	for _, snapshot := range old {
		outcome, err := snapshots.Remove(sandbox, snapshot.Id)
		if err != nil {
			return removed, err
		}
		if outcome == snapshots.OutcomeOk {
			removed++
		}
	}
	return removed, nil
}

// snapshotsBefore is every ready or failed snapshot taken before before.
func snapshotsBefore(sandbox *api.Sandbox, before int64) ([]backup.SnapshotRecord, error) {
	listed, err := snapshots.List(sandbox, "")
	if err != nil {
		return nil, err
	}
	old := []backup.SnapshotRecord{}
	for _, snapshot := range listed {
		if snapshot.Status != snapshots.StatusReady && snapshot.Status != snapshots.StatusFailed {
			continue
		}
		if snapshot.Data < before {
			old = append(old, snapshot)
		}
	}
	return old, nil
}

// PastRetention is every snapshot older than the shortest retention of the
// records: it holds data the databases already deleted, which lives on until
// the snapshot is removed.
func PastRetention(sandbox *api.Sandbox) ([]backup.SnapshotRecord, error) {
	retentions, err := CurrentRetentions(sandbox)
	if err != nil {
		return nil, err
	}
	shortest := retentions.Shortest()
	if shortest == privacy.Forever {
		return []backup.SnapshotRecord{}, nil
	}
	return snapshotsBefore(sandbox, Cutoff(sandbox, shortest))
}

// cleanupStarted holds a token once the cleanup goroutine runs.
var cleanupStarted = make(chan struct{}, 1)

// watchMilliseconds is how often the watcher looks for a restore.
const watchMilliseconds = 2000

// StartCleanup runs the cleanup in a goroutine of its own, once per process —
// a run at once, then one every hour, as long as the process lives — and the
// watcher of the restores beside it.
func StartCleanup(sandbox *api.Sandbox) {
	select {
	case cleanupStarted <- struct{}{}:
	default:
		return
	}
	go cleanup(sandbox)
	go watchRestores(sandbox)
}

// watchRestores is the body of the watcher goroutine: once a restore ends,
// the privacy settings and the IP key are read again from the store — a
// restore that included the backoffice put privacy-db back — and so are the
// banners, and every erasure kept is applied again, so erased data does not
// come back with the snapshot. It never returns.
func watchRestores(sandbox *api.Sandbox) {
	seen := snapshots.RestoresEnded(sandbox)
	for {
		if ended := snapshots.RestoresEnded(sandbox); ended != seen && !snapshots.Restoring(sandbox) {
			seen = ended
			privacy.InvalidateCache(sandbox)
			InvalidateBanners(sandbox)
			replayed, err := ReplayErasures(sandbox)
			if err != nil {
				sandbox.Deps.StdDeps.Eprintf("warning: the erasures could not be applied again after the restore: %s\n", err.Error())
			} else if replayed > 0 {
				sandbox.Deps.StdDeps.Logf("privacy: %d rows erased before were deleted again after the restore\n", replayed)
			}
		}
		sandbox.Deps.SleepDeps.Sleep(watchMilliseconds)
	}
}

// cleanup is the body of the cleanup goroutine. It never returns.
func cleanup(sandbox *api.Sandbox) {
	for {
		RunCleanup(sandbox)
		sandbox.Deps.SleepDeps.Sleep(cleanupMilliseconds)
	}
}

// RunCleanup is one run of the cleanup: it waits while a backup is taken or
// restored, sweeps by the retentions saved, and logs how many rows, ips and
// snapshots it touched — counts alone, never a value.
func RunCleanup(sandbox *api.Sandbox) {
	privacy.HoldForBackups(sandbox)
	retentions, err := CurrentRetentions(sandbox)
	if err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the retention cleanup could not read the retentions: %s\n", err.Error())
		return
	}
	swept, err := Sweep(sandbox, retentions, true)
	if err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the retention cleanup stopped: %s\n", err.Error())
	}
	if swept.Total() > 0 || swept.Released > 0 {
		sandbox.Deps.StdDeps.Logf("privacy: retention cleanup deleted %d contact requests, %d views, %d agent summaries, %d access log rows and %d snapshots; blanked %d contact request ips, %d view ips and %d agent usage ips; released %d holds\n",
			swept.Contacts, swept.Views, swept.Summaries, swept.AccessLog, swept.Snapshots, swept.ContactIps, swept.ViewIps, swept.UsageIps, swept.Released)
	}
	if forgotten, err := ForgetErasures(sandbox); err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the retention cleanup could not forget the old erasures: %s\n", err.Error())
	} else if forgotten > 0 {
		sandbox.Deps.StdDeps.Logf("privacy: %d erasures older than every snapshot forgotten\n", forgotten)
	}
	InvalidateBanners(sandbox)
}

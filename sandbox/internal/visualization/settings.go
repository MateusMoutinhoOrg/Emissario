package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
)

// Settings is what a root sets at SettingsPath, kept as the one record of the
// visualization-setting table.
type Settings struct {
	// Enabled records the views; off, the website is served as ever and
	// nothing is recorded.
	Enabled bool
	// RetentionDays is how many days a view is kept; 0 keeps them forever.
	RetentionDays int64
	// IgnoredIps are the ips and CIDR ranges no view is ever recorded from,
	// each in the spelling backofficehttp.CanonicalIpOrRange gives it.
	IgnoredIps []string
}

// DefaultSettings are the settings before a root saved any: tracking on,
// views kept DefaultRetentionDays, no ip ignored.
func DefaultSettings(sandbox *api.Sandbox) Settings {
	return Settings{Enabled: true, RetentionDays: DefaultRetentionDays, IgnoredIps: []string{}}
}

// settingsCacheSeconds is how long the writer trusts the settings it read: a
// save in this process drops them at once, a restore of a backup is seen
// within it.
const settingsCacheSeconds = 10

// settingsCache is the settings the writer reads, guarded by settingsLock.
type settingsCache struct {
	settings Settings
	readAt   int64
	valid    bool
}

// cached is the settings the writer last read.
var cached settingsCache

// settingsLock is held while the settings record or cached is read or
// written.
var settingsLock = make(chan struct{}, 1)

// ReadSettings is the settings a root saved, DefaultSettings before any.
func ReadSettings(sandbox *api.Sandbox) (Settings, error) {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	return readSettings(sandbox)
}

// readSettings is ReadSettings under settingsLock.
func readSettings(sandbox *api.Sandbox) (Settings, error) {
	record, ok, err := findSetting(sandbox)
	if err != nil || !ok {
		return DefaultSettings(sandbox), err
	}
	return Settings{
		Enabled:       record.Enabled == 1,
		RetentionDays: record.RetentionDays,
		IgnoredIps:    splitLines(sandbox, record.IgnoredIps),
	}, nil
}

// currentSettings is the settings the writer goes by: the ones read within
// settingsCacheSeconds, else read again. A store that fails keeps the last
// ones read, DefaultSettings when none was.
func currentSettings(sandbox *api.Sandbox) Settings {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	now := NowSeconds(sandbox)
	if cached.valid && now-cached.readAt < settingsCacheSeconds && now >= cached.readAt {
		return cached.settings
	}
	settings, err := readSettings(sandbox)
	if err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the visualization settings could not be read: %s\n", err.Error())
		if cached.valid {
			return cached.settings
		}
	}
	cached = settingsCache{settings: settings, readAt: now, valid: true}
	return settings
}

// findSetting is the settings record — the first one, were there ever two —
// false when there is none. The caller holds settingsLock.
func findSetting(sandbox *api.Sandbox) (visualization_db.VisualizationSettingRecord, bool, error) {
	records, err := visualization_db.New(sandbox).ListVisualizationSettings(visualization_db.VisualizationSettingFilter{})
	if err != nil {
		return visualization_db.VisualizationSettingRecord{}, false, err
	}
	if len(records) == 0 {
		return visualization_db.VisualizationSettingRecord{}, false, nil
	}
	return records[0], true, nil
}

// SaveSettings writes settings — already checked by CheckSettings — on the
// settings record, creating it the first time. The writer goes by them from
// the next view on.
func SaveSettings(sandbox *api.Sandbox, settings Settings) error {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	cached = settingsCache{}

	enabled := int64(0)
	if settings.Enabled {
		enabled = 1
	}
	ignored := sandbox.Deps.StringsDeps.Join(settings.IgnoredIps, "\n")
	db := visualization_db.New(sandbox)
	record, ok, err := findSetting(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		_, err := db.AddVisualizationSetting(visualization_db.VisualizationSettingInput{
			Enabled:       enabled,
			RetentionDays: settings.RetentionDays,
			IgnoredIps:    ignored,
		})
		return err
	}
	if err := db.SetVisualizationSettingEnabled(record.Id, enabled); err != nil {
		return err
	}
	if err := db.SetVisualizationSettingRetentionDays(record.Id, settings.RetentionDays); err != nil {
		return err
	}
	return db.SetVisualizationSettingIgnoredIps(record.Id, ignored)
}

// ParseRetention reads the retention a form sends: blank keeps the views
// forever (0), anything else is a whole number of days from 1 to
// MaxRetentionDays. message is why text is refused, "" when it is not.
func ParseRetention(sandbox *api.Sandbox, text string) (int64, string) {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if text == "" {
		return 0, ""
	}
	days, err := sandbox.Deps.StringsDeps.ParseInt(text, 10, 64)
	if err != nil || days < 1 || days > MaxRetentionDays {
		return 0, sandbox.Deps.StdDeps.Sprintf("The retention is a number of days from 1 to %d, or blank to keep the views forever.", MaxRetentionDays)
	}
	return days, ""
}

// RequiredNotices is every notice turning before into after asks for: the
// retention notice of the views when the retention changes past its warning.
func RequiredNotices(before Settings, after Settings) []string {
	return privacy.RetentionNotice([]string{}, privacy.DataView, before.RetentionDays, after.RetentionDays, 0)
}

// SplitIgnoredIps reads the ignored ips a form sends, one per line.
func SplitIgnoredIps(sandbox *api.Sandbox, text string) []string {
	return splitLines(sandbox, text)
}

// CheckSettings answers settings with every ignored ip in its canonical
// spelling, duplicates dropped, and the sentence for the first value it
// refuses — with the field it is about, retention_days or ignored_ips — ""
// when every one passed.
func CheckSettings(sandbox *api.Sandbox, settings Settings) (Settings, string, string) {
	if settings.RetentionDays < 0 || settings.RetentionDays > MaxRetentionDays {
		return settings, "retention_days", sandbox.Deps.StdDeps.Sprintf("The retention is a number of days from 1 to %d, or 0 to keep the views forever.", MaxRetentionDays)
	}
	clean := []string{}
	seen := map[string]bool{}
	for _, entry := range settings.IgnoredIps {
		entry = sandbox.Deps.StringsDeps.TrimSpace(entry)
		if entry == "" {
			continue
		}
		canonical, ok := backofficehttp.CanonicalIpOrRange(sandbox, entry)
		if !ok {
			return settings, "ignored_ips", "\"" + truncate(entry, 64) + "\" is neither an ip (203.0.113.7) nor a CIDR range (203.0.113.0/24)."
		}
		if seen[canonical] {
			continue
		}
		seen[canonical] = true
		clean = append(clean, canonical)
	}
	if len(clean) > MaxIgnoredIps {
		return settings, "ignored_ips", sandbox.Deps.StdDeps.Sprintf("At most %d ips and ranges are ignored.", MaxIgnoredIps)
	}
	settings.IgnoredIps = clean
	return settings, "", ""
}

// Ignored tells whether ip is one of the ignored ips of settings, or falls in
// one of their ranges.
func Ignored(sandbox *api.Sandbox, settings Settings, ip string) bool {
	for _, entry := range settings.IgnoredIps {
		if backofficehttp.IpMatches(sandbox, entry, ip) {
			return true
		}
	}
	return false
}

// splitLines is every non-blank line of text, trimmed.
func splitLines(sandbox *api.Sandbox, text string) []string {
	lines := []string{}
	for _, line := range sandbox.Deps.StringsDeps.Split(text, "\n") {
		line = sandbox.Deps.StringsDeps.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

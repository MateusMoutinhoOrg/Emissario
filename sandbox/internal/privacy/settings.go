package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
)

// The IP modes: what visualization-view.ip, contact-request.ip and
// agent-usage.ip hold.
const (
	// IpModeHashed stores "h:" and the hex of 16 bytes of HMAC-SHA256 under
	// the IP key, IPv6 grouped by its /64: the default.
	IpModeHashed = "hashed"
	// IpModeFull stores the IP as it came, and the full user agent of a
	// view; it asks for the ip-full-v1 notice.
	IpModeFull = "full"
)

// What viewers see of a full IP.
const (
	ViewerIpMasked = "masked"
	ViewerIpFull   = "full"
)

// The templates of the privacy notice draft.
const (
	DraftBr   = "br"
	DraftEu   = "eu"
	DraftBrEu = "br-eu"
)

// Drafts is every template of the privacy notice draft.
var Drafts = []string{DraftBrEu, DraftBr, DraftEu}

// IpKeyBytes is how long the IP key is.
const IpKeyBytes = 32

// Settings is the one privacy-setting record but its IP key: how IPs are
// kept and shown, the access log, the retentions privacy-db owns, the
// template of the notice draft and the preset applied.
type Settings struct {
	IpMode                 string
	ViewerIp               string
	AccessLog              bool
	AccessLogRetentionDays int64
	AccessLogTotp          bool
	AccessLogInventory     bool
	IpRetentionDays        int64
	BackupRetentionDays    int64
	NoticeDraft            string
	Preset                 string
}

// DefaultSettings are the settings of a new install: private by default.
// IPs hashed and masked for viewers, the access log off — two-factor and
// the inventory on, for when a root turns it on — IPs blanked after 184
// days, snapshots kept until a root deletes them, the notice draft for
// Brazil and the EU, and no preset.
func DefaultSettings(sandbox *api.Sandbox) Settings {
	return Settings{
		IpMode:                 IpModeHashed,
		ViewerIp:               ViewerIpMasked,
		AccessLog:              false,
		AccessLogRetentionDays: Rule(DataAccessLog).Default,
		AccessLogTotp:          true,
		AccessLogInventory:     true,
		IpRetentionDays:        Rule(DataIp).Default,
		BackupRetentionDays:    Rule(DataBackup).Default,
		NoticeDraft:            DraftBrEu,
		Preset:                 "",
	}
}

// settingsCacheSeconds is how long a request goes by the settings it read:
// every view reads them, and a save is seen within it.
const settingsCacheSeconds = 10

// cache is the settings and the IP key last read, guarded by lock.
var cache struct {
	settings Settings
	readAt   int64
	valid    bool
	ipKey    []byte
}

// lock is held while the privacy-setting record or the cache is read or
// written.
var lock = make(chan struct{}, 1)

// ReadSettings answers the settings as stored, the defaults while there is
// no record.
func ReadSettings(sandbox *api.Sandbox) (Settings, error) {
	lock <- struct{}{}
	defer func() { <-lock }()
	return readSettings(sandbox)
}

// CurrentSettings answers the settings read within settingsCacheSeconds, or
// reads them again. A store that fails keeps the last ones read, and
// answers its error when none was.
func CurrentSettings(sandbox *api.Sandbox) (Settings, error) {
	lock <- struct{}{}
	defer func() { <-lock }()
	now := NowSeconds(sandbox)
	if cache.valid && now-cache.readAt < settingsCacheSeconds && now >= cache.readAt {
		return cache.settings, nil
	}
	settings, err := readSettings(sandbox)
	if err != nil {
		if cache.valid {
			sandbox.Deps.StdDeps.Eprintf("warning: the privacy settings could not be read: %s\n", err.Error())
			return cache.settings, nil
		}
		return settings, err
	}
	cache.settings = settings
	cache.readAt = now
	cache.valid = true
	return settings, nil
}

// readSettings is ReadSettings under lock.
func readSettings(sandbox *api.Sandbox) (Settings, error) {
	record, ok, err := findRecord(sandbox)
	if err != nil || !ok {
		return DefaultSettings(sandbox), err
	}
	return settingsOf(sandbox, record), nil
}

// settingsOf is the Settings record holds, every value out of its range
// read as its default.
func settingsOf(sandbox *api.Sandbox, record privacy_db.PrivacySettingRecord) Settings {
	defaults := DefaultSettings(sandbox)
	settings := Settings{
		IpMode:                 record.IpMode,
		ViewerIp:               record.ViewerIp,
		AccessLog:              record.AccessLog == 1,
		AccessLogRetentionDays: record.AccessLogRetentionDays,
		AccessLogTotp:          record.AccessLogTotp == 1,
		AccessLogInventory:     record.AccessLogInventory == 1,
		IpRetentionDays:        record.IpRetentionDays,
		BackupRetentionDays:    record.BackupRetentionDays,
		NoticeDraft:            record.NoticeDraft,
		Preset:                 record.Preset,
	}
	if settings.IpMode != IpModeFull {
		settings.IpMode = IpModeHashed
	}
	if settings.ViewerIp != ViewerIpFull {
		settings.ViewerIp = ViewerIpMasked
	}
	for _, days := range []*int64{&settings.AccessLogRetentionDays, &settings.IpRetentionDays, &settings.BackupRetentionDays} {
		if *days < 0 || *days > MaxRetentionDays {
			*days = Forever
		}
	}
	if !validDraft(settings.NoticeDraft) {
		settings.NoticeDraft = defaults.NoticeDraft
	}
	return settings
}

// validDraft tells whether draft is a template of the notice draft.
func validDraft(draft string) bool {
	for _, known := range Drafts {
		if draft == known {
			return true
		}
	}
	return false
}

// findRecord is the privacy-setting record — the first, were there ever two
// — false when there is none. The caller holds lock.
func findRecord(sandbox *api.Sandbox) (privacy_db.PrivacySettingRecord, bool, error) {
	records, err := privacy_db.New(sandbox).ListPrivacySettings(privacy_db.PrivacySettingFilter{})
	if err != nil {
		return privacy_db.PrivacySettingRecord{}, false, err
	}
	if len(records) == 0 {
		return privacy_db.PrivacySettingRecord{}, false, nil
	}
	return records[0], true, nil
}

// boolInt is 1 for true, 0 for false.
func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

// CheckSettings refuses the first value of settings that is not one.
func CheckSettings(sandbox *api.Sandbox, settings Settings) *Refusal {
	if settings.IpMode != IpModeHashed && settings.IpMode != IpModeFull {
		return refuse("ip_mode", "The IP mode is hashed or full.")
	}
	if settings.ViewerIp != ViewerIpMasked && settings.ViewerIp != ViewerIpFull {
		return refuse("viewer_ip", "What viewers see of a full IP is masked or full.")
	}
	if !validDraft(settings.NoticeDraft) {
		return refuse("notice_draft", "The notice draft is br, eu or br-eu.")
	}
	for _, retention := range []struct {
		field string
		days  int64
	}{
		{"access_log_retention_days", settings.AccessLogRetentionDays},
		{"ip_retention_days", settings.IpRetentionDays},
		{"backup_retention_days", settings.BackupRetentionDays},
	} {
		if refusal := CheckRetention(sandbox, retention.field, retention.days); refusal != nil {
			return refusal
		}
	}
	return nil
}

// RequiredNotices is every notice turning before into after asks for: a
// value turned to one that needs a notice, never one left as it was.
// longest is the longest retention of the data a snapshot keeps.
func RequiredNotices(before Settings, after Settings, longest int64) []string {
	notices := []string{}
	if after.IpMode == IpModeFull && before.IpMode != IpModeFull {
		notices = append(notices, NoticeIpFull)
	}
	if after.ViewerIp == ViewerIpFull && before.ViewerIp != ViewerIpFull {
		notices = append(notices, NoticeViewerIp)
	}
	if after.AccessLog && !before.AccessLog {
		notices = append(notices, NoticeAccessLog)
	}
	if !after.AccessLogTotp && before.AccessLogTotp {
		notices = append(notices, NoticeAccessLogNoTotp)
	}
	if !after.AccessLogInventory && before.AccessLogInventory {
		notices = append(notices, NoticeAccessLogNoInvent)
	}
	notices = RetentionNotice(notices, DataAccessLog, before.AccessLogRetentionDays, after.AccessLogRetentionDays, 0)
	notices = RetentionNotice(notices, DataIp, before.IpRetentionDays, after.IpRetentionDays, 0)
	notices = RetentionNotice(notices, DataBackup, before.BackupRetentionDays, after.BackupRetentionDays, longest)
	return notices
}

// SaveSettings writes settings on the privacy-setting record, adding it —
// with a new IP key — the first time. Requests go by them from the next one
// on.
func SaveSettings(sandbox *api.Sandbox, settings Settings) error {
	lock <- struct{}{}
	defer func() { <-lock }()
	cache.valid = false
	db := privacy_db.New(sandbox)
	record, ok, err := findRecord(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		key, err := newIpKey(sandbox)
		if err != nil {
			return err
		}
		_, err = db.AddPrivacySetting(inputOf(settings, key, NowSeconds(sandbox)))
		return err
	}
	for _, write := range []func() error{
		func() error { return db.SetPrivacySettingIpMode(record.Id, settings.IpMode) },
		func() error { return db.SetPrivacySettingViewerIp(record.Id, settings.ViewerIp) },
		func() error { return db.SetPrivacySettingAccessLog(record.Id, boolInt(settings.AccessLog)) },
		func() error {
			return db.SetPrivacySettingAccessLogRetentionDays(record.Id, settings.AccessLogRetentionDays)
		},
		func() error { return db.SetPrivacySettingAccessLogTotp(record.Id, boolInt(settings.AccessLogTotp)) },
		func() error {
			return db.SetPrivacySettingAccessLogInventory(record.Id, boolInt(settings.AccessLogInventory))
		},
		func() error { return db.SetPrivacySettingIpRetentionDays(record.Id, settings.IpRetentionDays) },
		func() error { return db.SetPrivacySettingBackupRetentionDays(record.Id, settings.BackupRetentionDays) },
		func() error { return db.SetPrivacySettingNoticeDraft(record.Id, settings.NoticeDraft) },
		func() error { return db.SetPrivacySettingPreset(record.Id, settings.Preset) },
	} {
		if err := write(); err != nil {
			return err
		}
	}
	return nil
}

// inputOf is the record of settings with key as its IP key.
func inputOf(settings Settings, key []byte, now int64) privacy_db.PrivacySettingInput {
	return privacy_db.PrivacySettingInput{
		IpMode:                 settings.IpMode,
		ViewerIp:               settings.ViewerIp,
		AccessLog:              boolInt(settings.AccessLog),
		AccessLogRetentionDays: settings.AccessLogRetentionDays,
		AccessLogTotp:          boolInt(settings.AccessLogTotp),
		AccessLogInventory:     boolInt(settings.AccessLogInventory),
		IpKey:                  key,
		IpKeyRotatedAt:         now,
		IpRetentionDays:        settings.IpRetentionDays,
		BackupRetentionDays:    settings.BackupRetentionDays,
		NoticeDraft:            settings.NoticeDraft,
		Preset:                 settings.Preset,
	}
}

// newIpKey is IpKeyBytes random bytes.
func newIpKey(sandbox *api.Sandbox) ([]byte, error) {
	text, err := sandbox.Deps.RandDeps.Hex(IpKeyBytes)
	if err != nil {
		return nil, err
	}
	return decodeHex(text)
}

// decodeHex reads lower- or upper-case hex.
func decodeHex(text string) ([]byte, error) {
	if len(text)%2 != 0 {
		return nil, errOddHex
	}
	out := make([]byte, len(text)/2)
	for index := 0; index < len(out); index++ {
		high, ok1 := hexValue(text[2*index])
		low, ok2 := hexValue(text[2*index+1])
		if !ok1 || !ok2 {
			return nil, errOddHex
		}
		out[index] = high<<4 | low
	}
	return out, nil
}

// hexValue is the value of one hex digit.
func hexValue(digit byte) (byte, bool) {
	switch {
	case digit >= '0' && digit <= '9':
		return digit - '0', true
	case digit >= 'a' && digit <= 'f':
		return digit - 'a' + 10, true
	case digit >= 'A' && digit <= 'F':
		return digit - 'A' + 10, true
	}
	return 0, false
}

// hexError is the error of text that is not hex.
type hexError struct{}

func (hexError) Error() string { return "not hex" }

var errOddHex error = hexError{}

// EnsureSettings adds the privacy-setting record with the defaults and a new
// IP key when there is none, so the IP key exists before the first view is
// hashed. start-server calls it at start.
func EnsureSettings(sandbox *api.Sandbox) error {
	lock <- struct{}{}
	defer func() { <-lock }()
	_, ok, err := findRecord(sandbox)
	if err != nil || ok {
		return err
	}
	key, err := newIpKey(sandbox)
	if err != nil {
		return err
	}
	cache.valid = false
	_, err = privacy_db.New(sandbox).AddPrivacySetting(inputOf(DefaultSettings(sandbox), key, NowSeconds(sandbox)))
	return err
}

// InvalidateCache makes the next read go to the store again, for the
// settings and the IP key alike: a restore that put privacy-db back calls it.
func InvalidateCache(sandbox *api.Sandbox) {
	lock <- struct{}{}
	cache.valid = false
	cache.ipKey = nil
	<-lock
}

// IpKey answers the key the IPs are hashed under, adding the settings record
// with a new one when there is none.
func IpKey(sandbox *api.Sandbox) ([]byte, error) {
	lock <- struct{}{}
	if cache.ipKey != nil {
		key := cache.ipKey
		<-lock
		return key, nil
	}
	record, ok, err := findRecord(sandbox)
	<-lock
	if err != nil {
		return nil, err
	}
	if !ok || len(record.IpKey) != IpKeyBytes {
		if err := EnsureSettings(sandbox); err != nil {
			return nil, err
		}
		lock <- struct{}{}
		record, ok, err = findRecord(sandbox)
		<-lock
		if err != nil {
			return nil, err
		}
		if !ok || len(record.IpKey) != IpKeyBytes {
			return nil, sandbox.Deps.StdDeps.Errorf("the IP key is missing from the privacy settings")
		}
	}
	lock <- struct{}{}
	cache.ipKey = record.IpKey
	<-lock
	return record.IpKey, nil
}

// IpKeyRotatedAt is when the IP key was made, in Unix seconds: hashes made
// before it no longer match.
func IpKeyRotatedAt(sandbox *api.Sandbox) (int64, error) {
	lock <- struct{}{}
	defer func() { <-lock }()
	record, ok, err := findRecord(sandbox)
	if err != nil || !ok {
		return 0, err
	}
	return record.IpKeyRotatedAt, nil
}

// RotateIpKey replaces the IP key with a new random one: the hashes made
// before can no longer be matched to new visits nor confirmed against an IP.
func RotateIpKey(sandbox *api.Sandbox) error {
	if err := EnsureSettings(sandbox); err != nil {
		return err
	}
	key, err := newIpKey(sandbox)
	if err != nil {
		return err
	}
	lock <- struct{}{}
	defer func() { <-lock }()
	record, ok, err := findRecord(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		return sandbox.Deps.StdDeps.Errorf("the privacy settings are missing")
	}
	db := privacy_db.New(sandbox)
	if err := db.SetPrivacySettingIpKey(record.Id, key); err != nil {
		return err
	}
	cache.ipKey = nil
	return db.SetPrivacySettingIpKeyRotatedAt(record.Id, NowSeconds(sandbox))
}

package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// Settings is the one contact-setting record: whether the route takes contact
// requests, the largest body it takes, the limits of one ip, and how many days
// a contact request is kept. An ip daily limit of 0 is no limit; a retention
// of privacy.Forever keeps the requests until a root deletes them.
type Settings struct {
	Enabled       bool
	MaxBytes      int64
	RatePerMinute int64
	IpDailyLimit  int64
	RetentionDays int64
}

// SettingsInput is what a form or a JSON body sends to change the settings,
// with the ids of the notices it acknowledged.
type SettingsInput struct {
	Enabled       bool
	MaxBytes      int64
	RatePerMinute int64
	IpDailyLimit  int64
	RetentionDays int64
	Acknowledge   []string
}

// DefaultSettings are the settings before a root saved any: the route off —
// private by default, as the GDPR art. 25 and the LGPD art. 46 §2 ask — and
// the requests kept 180 days once a root turns it on, within the default
// limits.
func DefaultSettings(sandbox *api.Sandbox) Settings {
	return Settings{Enabled: false, MaxBytes: DefaultMaxBytes, RatePerMinute: DefaultRatePerMinute, RetentionDays: privacy.Rule(privacy.DataContact).Default}
}

// settingsCacheSeconds is how long the route goes by the settings it read:
// every contact request reads them, and a save is seen within it.
const settingsCacheSeconds = 10

// settingsCache is the settings the route reads, guarded by settingsLock.
type settingsCache struct {
	settings Settings
	readAt   int64
	valid    bool
}

// cached is the settings the route last read.
var cached settingsCache

// settingsLock is held while the contact-setting record or cached is read or
// written.
var settingsLock = make(chan struct{}, 1)

// ReadSettings answers the settings, the defaults while no root saved any.
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
	settings := Settings{
		Enabled:       record.Enabled == 1,
		MaxBytes:      record.MaxBytes,
		RatePerMinute: record.RatePerMinute,
		IpDailyLimit:  record.IpDailyLimit,
		RetentionDays: record.RetentionDays,
	}
	if settings.RetentionDays < 0 || settings.RetentionDays > privacy.MaxRetentionDays {
		settings.RetentionDays = privacy.Forever
	}
	if settings.MaxBytes < MinMaxBytes || settings.MaxBytes > MaxMaxBytes {
		settings.MaxBytes = DefaultMaxBytes
	}
	if settings.RatePerMinute < 1 {
		settings.RatePerMinute = DefaultRatePerMinute
	}
	return settings, nil
}

// currentSettings is the settings the route goes by: the ones read within
// settingsCacheSeconds, else read again. A store that fails keeps the last
// ones read, and answers its error when none was.
func currentSettings(sandbox *api.Sandbox) (Settings, error) {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	now := NowSeconds(sandbox)
	if cached.valid && now-cached.readAt < settingsCacheSeconds && now >= cached.readAt {
		return cached.settings, nil
	}
	settings, err := readSettings(sandbox)
	if err != nil {
		if cached.valid {
			sandbox.Deps.StdDeps.Eprintf("warning: the contact settings could not be read: %s\n", err.Error())
			return cached.settings, nil
		}
		return settings, err
	}
	cached = settingsCache{settings: settings, readAt: now, valid: true}
	return settings, nil
}

// findSetting is the contact-setting record — the first, were there ever two
// — false when there is none. The caller holds settingsLock.
func findSetting(sandbox *api.Sandbox) (contact_db.ContactSettingRecord, bool, error) {
	records, err := contact_db.New(sandbox).ListContactSettings(contact_db.ContactSettingFilter{})
	if err != nil {
		return contact_db.ContactSettingRecord{}, false, err
	}
	if len(records) == 0 {
		return contact_db.ContactSettingRecord{}, false, nil
	}
	return records[0], true, nil
}

// ApplySettings checks input and saves it. A refused value comes back as the
// refusal, and nothing is written. Turning the route on asks for the
// contacts-v1 notice, and a retention past its warning for the retention
// notice: until input acknowledges them the refusal is a 409 listing them.
// Nothing records that a notice was acknowledged.
func ApplySettings(sandbox *api.Sandbox, input SettingsInput) (Settings, *Refusal, error) {
	settings, refusal := CheckSettings(sandbox, input)
	if refusal != nil {
		return settings, refusal, nil
	}
	current, err := ReadSettings(sandbox)
	if err != nil {
		return settings, nil, err
	}
	if missing := privacy.NeedsAcknowledgement(RequiredNotices(current, settings), input.Acknowledge); missing != nil {
		return settings, &Refusal{Status: missing.Status, Field: missing.Field, Message: missing.Message, Notices: missing.Notices}, nil
	}
	return settings, nil, saveSettings(sandbox, settings)
}

// RequiredNotices is every notice turning before into after asks for.
func RequiredNotices(before Settings, after Settings) []string {
	notices := []string{}
	if after.Enabled && !before.Enabled {
		notices = append(notices, privacy.NoticeContacts)
	}
	return privacy.RetentionNotice(notices, privacy.DataContact, before.RetentionDays, after.RetentionDays, 0)
}

// CheckSettings answers input as the settings to save, or the refusal of the
// first value that does not hold.
func CheckSettings(sandbox *api.Sandbox, input SettingsInput) (Settings, *Refusal) {
	settings := Settings{
		Enabled:       input.Enabled,
		MaxBytes:      input.MaxBytes,
		RatePerMinute: input.RatePerMinute,
		IpDailyLimit:  input.IpDailyLimit,
		RetentionDays: input.RetentionDays,
	}
	if refusal := privacy.CheckRetention(sandbox, "retention_days", settings.RetentionDays); refusal != nil {
		return settings, refuse(refusal.Field, refusal.Message)
	}
	if settings.MaxBytes < MinMaxBytes || settings.MaxBytes > MaxMaxBytes {
		return settings, refuse("max_bytes", "The largest body is a number of bytes from 2 to 1048576.")
	}
	if settings.RatePerMinute < 1 || settings.RatePerMinute > MaxRatePerMinute {
		return settings, refuse("rate_per_minute", "The rate per minute is a number of requests from 1 to 10000.")
	}
	if settings.IpDailyLimit < 0 || settings.IpDailyLimit > MaxIpDailyLimit {
		return settings, refuse("ip_daily_limit", "The ip daily limit is a number of requests up to 1000000, or blank for no limit.")
	}
	return settings, nil
}

// saveSettings writes settings on the contact-setting record, adding it the
// first time. The route goes by them from the next request on.
func saveSettings(sandbox *api.Sandbox, settings Settings) error {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	cached = settingsCache{}
	enabled := int64(0)
	if settings.Enabled {
		enabled = 1
	}
	db := contact_db.New(sandbox)
	record, ok, err := findSetting(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		_, err := db.AddContactSetting(contact_db.ContactSettingInput{
			Enabled:       enabled,
			MaxBytes:      settings.MaxBytes,
			RatePerMinute: settings.RatePerMinute,
			IpDailyLimit:  settings.IpDailyLimit,
			RetentionDays: settings.RetentionDays,
		})
		return err
	}
	for _, write := range []func() error{
		func() error { return db.SetContactSettingEnabled(record.Id, enabled) },
		func() error { return db.SetContactSettingMaxBytes(record.Id, settings.MaxBytes) },
		func() error { return db.SetContactSettingRatePerMinute(record.Id, settings.RatePerMinute) },
		func() error { return db.SetContactSettingIpDailyLimit(record.Id, settings.IpDailyLimit) },
		func() error { return db.SetContactSettingRetentionDays(record.Id, settings.RetentionDays) },
	} {
		if err := write(); err != nil {
			return err
		}
	}
	return nil
}

// SettingsForm is the settings form as it was posted, every field text, and
// the notices it acknowledged.
type SettingsForm struct {
	Enabled       string
	MaxBytes      string
	RatePerMinute string
	IpDailyLimit  string
	RetentionDays string
	Acknowledge   []string
}

// InputOfForm reads form as a SettingsInput: a checkbox is on as 1 or on; a
// blank largest body is DefaultMaxBytes, a blank rate DefaultRatePerMinute
// and a blank ip daily limit no limit. A number that does not read is
// refused.
func InputOfForm(sandbox *api.Sandbox, form SettingsForm) (SettingsInput, *Refusal) {
	input := SettingsInput{Enabled: form.Enabled == "1" || form.Enabled == "on", Acknowledge: form.Acknowledge}
	retention, refusal := privacy.ParseRetention(sandbox, "retention_days", form.RetentionDays)
	if refusal != nil {
		return input, refuse(refusal.Field, refusal.Message)
	}
	input.RetentionDays = retention
	for _, number := range []struct {
		text     string
		fallback int64
		target   *int64
		field    string
		message  string
	}{
		{form.MaxBytes, DefaultMaxBytes, &input.MaxBytes, "max_bytes", "The largest body is a number of bytes from 2 to 1048576, or blank for 16384."},
		{form.RatePerMinute, DefaultRatePerMinute, &input.RatePerMinute, "rate_per_minute", "The rate per minute is a number of requests from 1 to 10000, or blank for 5."},
		{form.IpDailyLimit, 0, &input.IpDailyLimit, "ip_daily_limit", "The ip daily limit is a number of requests up to 1000000, or blank for no limit."},
	} {
		text := sandbox.Deps.StringsDeps.TrimSpace(number.text)
		if text == "" {
			*number.target = number.fallback
			continue
		}
		value, err := sandbox.Deps.StringsDeps.ParseInt(text, 10, 64)
		if err != nil {
			return input, refuse(number.field, number.message)
		}
		*number.target = value
	}
	return input, nil
}

// State is what the settings page and its JSON twin say beside the settings:
// how many contact requests are stored and in the inbox, whether the writer
// runs in this process, and how many requests found its queue full.
type State struct {
	Stored  int
	Inbox   int
	Running bool
	Dropped int64
}

// CurrentState is the State of the contact requests.
func CurrentState(sandbox *api.Sandbox) (State, error) {
	inbox, _, err := contact_db.CountContactRequests(sandbox, contact_db.New(sandbox), 0)
	if err != nil {
		return State{}, err
	}
	stored, err := contact_db.New(sandbox).CountContactRequest()
	if err != nil {
		return State{}, err
	}
	return State{Stored: stored, Inbox: inbox, Running: Running(sandbox), Dropped: Dropped(sandbox)}, nil
}

// SaveRetention writes days as the retention of the contact requests,
// leaving every other setting as it is: a privacy preset applies it, after
// the notices of the whole preset were acknowledged together.
func SaveRetention(sandbox *api.Sandbox, days int64) error {
	settings, err := ReadSettings(sandbox)
	if err != nil {
		return err
	}
	settings.RetentionDays = days
	return saveSettings(sandbox, settings)
}

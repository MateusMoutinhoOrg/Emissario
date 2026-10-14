package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
)

// Settings is the one contact-setting record: whether the route takes contact
// requests, the largest body it takes, and the limits of one ip. An ip daily
// limit of 0 is no limit.
type Settings struct {
	Enabled       bool
	MaxBytes      int64
	RatePerMinute int64
	IpDailyLimit  int64
}

// SettingsInput is what a form or a JSON body sends to change the settings.
type SettingsInput struct {
	Enabled       bool
	MaxBytes      int64
	RatePerMinute int64
	IpDailyLimit  int64
}

// DefaultSettings are the settings before a root saved any: the route on, so
// the form of a new website works at once, within the default limits.
func DefaultSettings(sandbox *api.Sandbox) Settings {
	return Settings{Enabled: true, MaxBytes: DefaultMaxBytes, RatePerMinute: DefaultRatePerMinute}
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
// refusal, and nothing is written.
func ApplySettings(sandbox *api.Sandbox, input SettingsInput) (Settings, *Refusal, error) {
	settings, refusal := CheckSettings(sandbox, input)
	if refusal != nil {
		return settings, refusal, nil
	}
	return settings, nil, saveSettings(sandbox, settings)
}

// CheckSettings answers input as the settings to save, or the refusal of the
// first value that does not hold.
func CheckSettings(sandbox *api.Sandbox, input SettingsInput) (Settings, *Refusal) {
	settings := Settings(input)
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
		})
		return err
	}
	for _, write := range []func() error{
		func() error { return db.SetContactSettingEnabled(record.Id, enabled) },
		func() error { return db.SetContactSettingMaxBytes(record.Id, settings.MaxBytes) },
		func() error { return db.SetContactSettingRatePerMinute(record.Id, settings.RatePerMinute) },
		func() error { return db.SetContactSettingIpDailyLimit(record.Id, settings.IpDailyLimit) },
	} {
		if err := write(); err != nil {
			return err
		}
	}
	return nil
}

// SettingsForm is the settings form as it was posted, every field text.
type SettingsForm struct {
	Enabled       string
	MaxBytes      string
	RatePerMinute string
	IpDailyLimit  string
}

// InputOfForm reads form as a SettingsInput: a checkbox is on as 1 or on; a
// blank largest body is DefaultMaxBytes, a blank rate DefaultRatePerMinute
// and a blank ip daily limit no limit. A number that does not read is
// refused.
func InputOfForm(sandbox *api.Sandbox, form SettingsForm) (SettingsInput, *Refusal) {
	input := SettingsInput{Enabled: form.Enabled == "1" || form.Enabled == "on"}
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

package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// SettingsInput is what the Privacy page or its JSON twin sends to change the
// privacy settings, with the ids of the notices it acknowledged. The preset
// applied is kept: a value that differs from it is drift, which the Presets
// page shows.
type SettingsInput struct {
	IpMode                 string
	ViewerIp               string
	AccessLog              bool
	AccessLogRetentionDays int64
	AccessLogTotp          bool
	AccessLogInventory     bool
	IpRetentionDays        int64
	BackupRetentionDays    int64
	NoticeDraft            string
	Acknowledge            []string
}

// ApplySettings checks input and saves it. A refused value comes back as the
// refusal, and nothing is written. Every value turned to one that needs a
// notice — full IPs, full IPs for viewers, the access log, the access log
// without two-factor or its inventory, a retention past its warning — asks
// for it: until input acknowledges them all the refusal is a 409 listing
// them. Nothing records that a notice was acknowledged, nor who saved.
func ApplySettings(sandbox *api.Sandbox, input SettingsInput) (privacy.Settings, *privacy.Refusal, error) {
	current, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return current, nil, err
	}
	settings := current
	settings.IpMode = input.IpMode
	settings.ViewerIp = input.ViewerIp
	settings.AccessLog = input.AccessLog
	settings.AccessLogRetentionDays = input.AccessLogRetentionDays
	settings.AccessLogTotp = input.AccessLogTotp
	settings.AccessLogInventory = input.AccessLogInventory
	settings.IpRetentionDays = input.IpRetentionDays
	settings.BackupRetentionDays = input.BackupRetentionDays
	settings.NoticeDraft = input.NoticeDraft
	if refusal := privacy.CheckSettings(sandbox, settings); refusal != nil {
		return settings, refusal, nil
	}
	retentions, err := CurrentRetentions(sandbox)
	if err != nil {
		return settings, nil, err
	}
	retentions.AccessLog = settings.AccessLogRetentionDays
	retentions.Ip = settings.IpRetentionDays
	required := privacy.RequiredNotices(current, settings, retentions.Longest())
	if missing := privacy.NeedsAcknowledgement(required, input.Acknowledge); missing != nil {
		return settings, missing, nil
	}
	if err := privacy.SaveSettings(sandbox, settings); err != nil {
		return settings, nil, err
	}
	InvalidateBanners(sandbox)
	return settings, nil, nil
}

// NoticeContextOf is what the notices a save of settings asks for are filled
// with: the retention each retention notice is about, and the longest of the
// records for the backup one.
func NoticeContextOf(sandbox *api.Sandbox, settings privacy.Settings, notice string) privacy.NoticeContext {
	context := privacy.NoticeContext{}
	switch notice {
	case privacy.RetentionNoticeId(privacy.DataAccessLog):
		context.Retention = settings.AccessLogRetentionDays
	case privacy.RetentionNoticeId(privacy.DataIp):
		context.Retention = settings.IpRetentionDays
	case privacy.RetentionNoticeId(privacy.DataBackup):
		context.Retention = settings.BackupRetentionDays
		if retentions, err := CurrentRetentions(sandbox); err == nil {
			retentions.AccessLog = settings.AccessLogRetentionDays
			retentions.Ip = settings.IpRetentionDays
			context.Longest = retentions.Longest()
		}
	}
	return context
}

// LoadSettingsNotices loads every notice of ids, each filled for settings.
func LoadSettingsNotices(sandbox *api.Sandbox, settings privacy.Settings, ids []string) ([]privacy.Notice, error) {
	notices := []privacy.Notice{}
	for _, id := range ids {
		notice, err := privacy.LoadNotice(sandbox, id, NoticeContextOf(sandbox, settings, id))
		if err != nil {
			return nil, err
		}
		notices = append(notices, notice)
	}
	return notices, nil
}

// SettingsForm is the Privacy settings form as it was posted, every field
// text, and the notices it acknowledged.
type SettingsForm struct {
	IpMode                 string
	ViewerIp               string
	AccessLog              string
	AccessLogRetentionDays string
	AccessLogTotp          string
	AccessLogInventory     string
	IpRetentionDays        string
	BackupRetentionDays    string
	NoticeDraft            string
	Acknowledge            []string
}

// InputOfForm reads form as a SettingsInput: a checkbox is on as 1 or on, a
// blank retention keeps the data until a root deletes it. A retention that
// does not read is refused.
func InputOfForm(sandbox *api.Sandbox, form SettingsForm) (SettingsInput, *privacy.Refusal) {
	checked := func(value string) bool { return value == "1" || value == "on" }
	input := SettingsInput{
		IpMode:             form.IpMode,
		ViewerIp:           form.ViewerIp,
		AccessLog:          checked(form.AccessLog),
		AccessLogTotp:      checked(form.AccessLogTotp),
		AccessLogInventory: checked(form.AccessLogInventory),
		NoticeDraft:        form.NoticeDraft,
		Acknowledge:        form.Acknowledge,
	}
	for _, retention := range []struct {
		field  string
		text   string
		target *int64
	}{
		{"access_log_retention_days", form.AccessLogRetentionDays, &input.AccessLogRetentionDays},
		{"ip_retention_days", form.IpRetentionDays, &input.IpRetentionDays},
		{"backup_retention_days", form.BackupRetentionDays, &input.BackupRetentionDays},
	} {
		days, refusal := privacy.ParseRetention(sandbox, retention.field, retention.text)
		if refusal != nil {
			return input, refusal
		}
		*retention.target = days
	}
	return input, nil
}

package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/encryption"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The paths of the Privacy page's forms.
const (
	PrivacyHashIpsPath   = "/admin/root/hash-privacy-ips"
	PrivacyRotateKeyPath = "/admin/root/rotate-privacy-ip-key"
	PrivacyReadersPath   = "/admin/root/set-access-log-readers"
)

// The outcome codes the Privacy page shows after a form.
const (
	PrivacyNoticeSaved          = "saved"
	PrivacyNoticeHashed         = "hashed"
	PrivacyNoticeRotated        = "rotated"
	PrivacyNoticeReleased       = "released"
	PrivacyNoticeHoldNotFound   = "hold-not-found"
	PrivacyNoticeReadersSaved   = "readers-saved"
	PrivacyNoticeReadersRefused = "readers-refused"
)

// PrivacySettingsForm is what the settings form of the Privacy page holds:
// what was saved, or what was sent and refused, and the notices a save asks
// for.
type PrivacySettingsForm struct {
	IpMode                 string
	ViewerIp               string
	AccessLog              bool
	AccessLogRetentionDays string
	AccessLogTotp          bool
	AccessLogInventory     bool
	IpRetentionDays        string
	BackupRetentionDays    string
	NoticeDraft            string
	Notices                []privacy.Notice
}

// PrivacySettingsFormOf is the form filled with settings.
func PrivacySettingsFormOf(sandbox *api.Sandbox, settings privacy.Settings) PrivacySettingsForm {
	return PrivacySettingsForm{
		IpMode:                 settings.IpMode,
		ViewerIp:               settings.ViewerIp,
		AccessLog:              settings.AccessLog,
		AccessLogRetentionDays: privacy.FormatRetention(sandbox, settings.AccessLogRetentionDays),
		AccessLogTotp:          settings.AccessLogTotp,
		AccessLogInventory:     settings.AccessLogInventory,
		IpRetentionDays:        privacy.FormatRetention(sandbox, settings.IpRetentionDays),
		BackupRetentionDays:    privacy.FormatRetention(sandbox, settings.BackupRetentionDays),
		NoticeDraft:            settings.NoticeDraft,
	}
}

// PrivacyKeptLine is one table of What this server keeps.
type PrivacyKeptLine struct {
	Table     string
	Fields    string
	Purpose   string
	Basis     string
	Retention string
	Count     string
	Oldest    string
}

// PrivacyHoldLine is one legal hold of the Privacy page.
type PrivacyHoldLine struct {
	Id     string
	Table  string
	Rows   string
	Reason string
	Until  string
}

// PrivacyReaderLine is one root of the access log readers form.
type PrivacyReaderLine struct {
	Id       string
	Username string
	Reader   bool
	Totp     bool
}

// PrivacySettingsPage is what backoffice/privacy_settings.html is rendered
// with.
type PrivacySettingsPage struct {
	Viewer Viewer
	Notice Notice
	Error  string
	Action string
	Form   PrivacySettingsForm
	Drafts []LabeledOption
	// Preset is the name of the preset applied, "" for none; PresetsHref is
	// the Presets page.
	Preset      string
	PresetsHref string
	// Encrypted, KeyId and Algorithm describe the encryption at rest.
	Encrypted bool
	KeyId     string
	Algorithm string
	// IpKeyMade is the day the IP key was made.
	IpKeyMade string
	// SourcePort is false when the access log cannot tell the source port
	// of a visitor: behind a proxy, with no --client-port-header.
	SourcePort bool
	// AccessLogHref is the access log page, "" for a root who is not a
	// reader.
	AccessLogHref string
	Readers       []PrivacyReaderLine
	ReadersAction string
	Kept          []PrivacyKeptLine
	Holds         []PrivacyHoldLine
	ReleaseAction string
	// InUse are the notices of the settings in use, to read again.
	InUse        []privacy.Notice
	HashAction   string
	RotateAction string
	FindHref     string
	DraftHref    string
	TotpHref     string
	// NoticeDraft is the privacy notice draft of the settings in use.
	NoticeDraft string
	Zone        string
}

// privacyNotice is how the Privacy page words an outcome code.
func privacyNotice(sandbox *api.Sandbox, code string, count int, reason string) Notice {
	switch code {
	case PrivacyNoticeSaved:
		return Notice{Text: "Privacy settings saved.", Kind: "ok"}
	case PrivacyNoticeHashed:
		return Notice{Text: plural(sandbox, count, "full IP", "full IPs") + " turned into hashes.", Kind: "ok"}
	case PrivacyNoticeRotated:
		return Notice{Text: "IP key rotated: the hashes made before no longer match new visits, nor an IP typed in a filter or in Find and erase.", Kind: "ok"}
	case PrivacyNoticeReleased:
		return Notice{Text: "Hold released: its rows follow their retention again.", Kind: "ok"}
	case PrivacyNoticeHoldNotFound:
		return Notice{Text: "That hold no longer exists.", Kind: "error"}
	case PrivacyNoticeReadersSaved:
		return Notice{Text: "Access log readers saved.", Kind: "ok"}
	case PrivacyNoticeReadersRefused:
		return Notice{Text: "Only roots read the access log.", Kind: "error"}
	}
	return holdNotice(sandbox, code, reason)
}

// RenderPrivacySettingsPage answers, under status, the Privacy page for user
// with form, message above it — or the sentence of the outcome code notice,
// count and reason telling what it did.
func RenderPrivacySettingsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, form PrivacySettingsForm, notice string, count int, reason string, message string) error {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	page := PrivacySettingsPage{
		Viewer:        viewerOf(sandbox, user),
		Notice:        privacyNotice(sandbox, notice, count, reason),
		Error:         message,
		Action:        privacy.SettingsPath,
		Form:          form,
		Drafts:        []LabeledOption{{Value: privacy.DraftBrEu, Label: "Brazil and the EU (pt-BR and en)"}, {Value: privacy.DraftBr, Label: "Brazil (pt-BR)"}, {Value: privacy.DraftEu, Label: "EU (en)"}},
		Preset:        settings.Preset,
		PresetsHref:   privacy.PresetsPath,
		Encrypted:     sandbox.Config.Encrypted,
		KeyId:         sandbox.Config.EncryptionKeyId,
		Algorithm:     encryption.Algorithm,
		SourcePort:    !sandbox.Config.AllowXForwardedFor || sandbox.Config.ClientPortHeader != "",
		ReadersAction: PrivacyReadersPath,
		ReleaseAction: ReleaseHoldPath,
		HashAction:    PrivacyHashIpsPath,
		RotateAction:  PrivacyRotateKeyPath,
		FindHref:      privacy.FindPath,
		DraftHref:     privacy.SettingsPath + "#notice-draft",
		TotpHref:      privacy.TotpPath,
		Zone:          visualization.ZoneName(sandbox),
	}
	if made, err := privacy.IpKeyRotatedAt(sandbox); err == nil && made > 0 {
		page.IpKeyMade = visualization.DayOf(sandbox, made)
	}
	if privacyadmin.CanReadAccessLog(sandbox, user, nil) == nil {
		page.AccessLogHref = privacy.AccessLogPath
	}

	readers, err := privacyadmin.Readers(sandbox)
	if err != nil {
		return err
	}
	for _, reader := range readers {
		page.Readers = append(page.Readers, PrivacyReaderLine{
			Id:       sandbox.Deps.StringsDeps.FormatInt(reader.Id, 10),
			Username: reader.Username,
			Reader:   reader.AccessLogReader == 1,
			Totp:     reader.TotpEnabledAt != 0,
		})
	}

	kept, err := privacyadmin.KeptTables(sandbox)
	if err != nil {
		return err
	}
	for _, table := range kept {
		line := PrivacyKeptLine{Table: table.Table, Fields: table.Fields, Purpose: table.Purpose, Basis: table.Basis, Retention: table.Retention, Count: formatCount(sandbox, table.Count), Oldest: "—"}
		if table.Oldest > 0 {
			line.Oldest = visualization.DayOf(sandbox, table.Oldest)
		}
		page.Kept = append(page.Kept, line)
	}

	holds, err := privacyadmin.VisibleHolds(sandbox, user, nil)
	if err != nil {
		return err
	}
	for _, hold := range holds {
		page.Holds = append(page.Holds, PrivacyHoldLine{
			Id:     sandbox.Deps.StringsDeps.FormatInt(hold.Id, 10),
			Table:  hold.Table,
			Rows:   formatCount(sandbox, len(hold.Ids)),
			Reason: hold.Reason,
			Until:  visualization.DayOf(sandbox, hold.Until-1),
		})
	}

	draft, err := privacyadmin.NoticeDraft(sandbox)
	if err != nil {
		return err
	}
	page.NoticeDraft = draft

	inUse, err := noticesInUse(sandbox, settings)
	if err != nil {
		return err
	}
	page.InUse = inUse
	return RenderHTML(sandbox, response, status, "backoffice/privacy_settings.html", page)
}

// noticesInUse is the current notice of every setting in use that needs one,
// to read again.
func noticesInUse(sandbox *api.Sandbox, settings privacy.Settings) ([]privacy.Notice, error) {
	ids := []string{}
	if contacts, err := contact.ReadSettings(sandbox); err == nil && contacts.Enabled {
		ids = append(ids, privacy.NoticeContacts)
	}
	host := ""
	if agentSettings, err := agent.ReadSettings(sandbox); err == nil {
		host = agent.ProviderHost(sandbox, agentSettings)
		if agentSettings.Enabled {
			ids = append(ids, privacy.NoticeAgent)
		}
		if agentSettings.SummariesEnabled {
			ids = append(ids, privacy.NoticeSummaries)
		}
	}
	if settings.IpMode == privacy.IpModeFull {
		ids = append(ids, privacy.NoticeIpFull)
	}
	if settings.ViewerIp == privacy.ViewerIpFull {
		ids = append(ids, privacy.NoticeViewerIp)
	}
	if settings.AccessLog {
		ids = append(ids, privacy.NoticeAccessLog)
		if !settings.AccessLogTotp {
			ids = append(ids, privacy.NoticeAccessLogNoTotp)
		}
		if !settings.AccessLogInventory {
			ids = append(ids, privacy.NoticeAccessLogNoInvent)
		}
	}
	return privacy.LoadNotices(sandbox, ids, privacy.NoticeContext{Host: host})
}

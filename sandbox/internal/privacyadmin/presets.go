package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// A preset is a named, versioned and complete set of values for the privacy
// settings, each with its reference: applying one is saving those settings,
// through the same ranges and the same notices, in one step. It says how
// data is kept, never whether it is collected — the contact requests, the
// agent and its summaries stay as a root set them — and it never sets the
// keys, who sees full IPs, the access log readers, its two-factor nor its
// inventory. The built-in ones — lgpd, gdpr and lgpd-gdpr — live in
// assets/privacy/presets, one file per version; the custom ones in the
// privacy-preset table. A version never changes: editing a preset saves the
// next one. Applying it keeps privacy-setting.preset = <id>@<version> and
// nothing else: not who applied it, nor when. A preset is a starting point,
// not a lock: a setting changed on its own page is drift, which the banners
// and check-privacy-preset show.

// presetsDir is where the built-in presets live, under assets/.
const presetsDir = "privacy/presets"

// PresetFormat is the format of a preset document.
const PresetFormat = 1

// The settings a preset holds, every one of them, in the order a page lists
// them.
const (
	SettingIpMode           = "ip-mode"
	SettingAccessLog        = "access-log"
	SettingAccessLogDays    = "access-log-retention-days"
	SettingIpDays           = "ip-retention-days"
	SettingContactDays      = "contact-retention-days"
	SettingViewDays         = "view-retention-days"
	SettingSummaryDays      = "summary-retention-days"
	SettingBackupDays       = "backup-retention-days"
	SettingNoticeDraft      = "notice-draft"
	ChecklistAlways         = "always"
	ChecklistAgent          = "agent"
	DefaultChecklistPreset  = "lgpd-gdpr"
	maxPresetChecklistItems = 50
)

// PresetSettings is every setting a preset holds.
var PresetSettings = []string{SettingIpMode, SettingAccessLog, SettingAccessLogDays, SettingIpDays, SettingContactDays, SettingViewDays, SettingSummaryDays, SettingBackupDays, SettingNoticeDraft}

// PresetValues is the value of every setting a preset holds; a retention of
// privacy.Forever keeps the data until a root deletes it.
type PresetValues struct {
	IpMode             string
	AccessLog          bool
	AccessLogRetention int64
	IpRetention        int64
	ContactRetention   int64
	ViewRetention      int64
	SummaryRetention   int64
	BackupRetention    int64
	NoticeDraft        string
}

// Retentions are the retentions of values, as the cleanup goes by them.
func (values PresetValues) Retentions() Retentions {
	return Retentions{
		Contact:   values.ContactRetention,
		View:      values.ViewRetention,
		Summary:   values.SummaryRetention,
		AccessLog: values.AccessLogRetention,
		Ip:        values.IpRetention,
		Backup:    values.BackupRetention,
	}
}

// Value is the value of setting in values: a string, a bool, an int64 days
// or nil for privacy.Forever.
func (values PresetValues) Value(setting string) any {
	days := func(days int64) any {
		if days == privacy.Forever {
			return nil
		}
		return days
	}
	switch setting {
	case SettingIpMode:
		return values.IpMode
	case SettingAccessLog:
		return values.AccessLog
	case SettingAccessLogDays:
		return days(values.AccessLogRetention)
	case SettingIpDays:
		return days(values.IpRetention)
	case SettingContactDays:
		return days(values.ContactRetention)
	case SettingViewDays:
		return days(values.ViewRetention)
	case SettingSummaryDays:
		return days(values.SummaryRetention)
	case SettingBackupDays:
		return days(values.BackupRetention)
	case SettingNoticeDraft:
		return values.NoticeDraft
	}
	return nil
}

// ChecklistItem is one obligation Emissario cannot fulfil for the operator.
// It is a reminder: nothing records whether it is done.
type ChecklistItem struct {
	Id        string
	Text      string
	Reference string
	// When is ChecklistAlways, or ChecklistAgent to show it only while the
	// agent is on.
	When string
}

// Preset is one preset, read and checked.
type Preset struct {
	Id            string
	Version       int64
	Name          string
	Description   string
	Jurisdictions []string
	Values        PresetValues
	// References holds the reference of each setting, "" for none.
	References map[string]string
	Checklist  []ChecklistItem
	Changes    string
	// BuiltIn is true for a preset of assets/, read only.
	BuiltIn bool
}

// Ref is how the preset is named: <id>@<version>.
func (preset Preset) Ref() string {
	return preset.Id + "@" + formatInt(preset.Version)
}

// formatInt spells value in base 10 with no dep: a preset is parsed before a
// sandbox is at hand in Ref.
func formatInt(value int64) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	negative := value < 0
	if negative {
		value = -value
	}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// presetRefusal is a 400 on a preset document, its field the part refused.
func presetRefusal(field string, message string) *privacy.Refusal {
	return privacy.Refuse(field, message)
}

// ParsePreset reads text as a preset document and checks it whole: the
// format, every setting present with a value in its range, no unknown key
// anywhere. A preset refused is not stored, and nothing is written.
func ParsePreset(sandbox *api.Sandbox, text string) (Preset, *privacy.Refusal) {
	strings := sandbox.Deps.StringsDeps
	document, err := sandbox.Deps.SerializableDeps.ParseJson(text)
	if err != nil || document == nil || !document.IsObject() {
		return Preset{}, presetRefusal("preset", "A preset is a JSON object.")
	}
	if refusal := onlyKeys(document, "preset", "format", "id", "version", "name", "description", "jurisdictions", "values", "checklist", "changes"); refusal != nil {
		return Preset{}, refusal
	}
	preset := Preset{References: map[string]string{}}
	format, ok := intOf(document, "format")
	if !ok || format != PresetFormat {
		return preset, presetRefusal("format", "format is 1.")
	}
	preset.Id, ok = stringOf(document, "id")
	if !ok || !matchesPattern(sandbox, `^[a-z0-9][a-z0-9-]{0,39}$`, preset.Id) {
		return preset, presetRefusal("id", "id is 1 to 40 lowercase letters, digits and dashes, starting with a letter or a digit.")
	}
	preset.Version, ok = intOf(document, "version")
	if !ok || preset.Version < 1 || preset.Version > 1000000 {
		return preset, presetRefusal("version", "version is a whole number from 1 up.")
	}
	preset.Name, ok = stringOf(document, "name")
	if !ok || strings.TrimSpace(preset.Name) == "" || len(preset.Name) > 120 {
		return preset, presetRefusal("name", "name is 1 to 120 characters.")
	}
	if document.HasKey("description") {
		if preset.Description, ok = stringOf(document, "description"); !ok || len(preset.Description) > 1000 {
			return preset, presetRefusal("description", "description is a text of at most 1000 characters.")
		}
	}
	if document.HasKey("changes") {
		if preset.Changes, ok = stringOf(document, "changes"); !ok || len(preset.Changes) > 2000 {
			return preset, presetRefusal("changes", "changes is a text of at most 2000 characters.")
		}
	}
	if document.HasKey("jurisdictions") {
		list, _ := document.GetObjectItem("jurisdictions")
		size, err := arraySize(list)
		if err != nil || size > 10 {
			return preset, presetRefusal("jurisdictions", "jurisdictions is a list of at most 10 codes, such as BR or EU.")
		}
		for index := 0; index < size; index++ {
			code, ok := itemString(list.GetArrayItem(index))
			if !ok || code == "" || len(code) > 8 {
				return preset, presetRefusal("jurisdictions", "jurisdictions is a list of at most 10 codes, such as BR or EU.")
			}
			preset.Jurisdictions = append(preset.Jurisdictions, code)
		}
	}
	if refusal := parseValues(sandbox, document, &preset); refusal != nil {
		return preset, refusal
	}
	if refusal := parseChecklist(sandbox, document, &preset); refusal != nil {
		return preset, refusal
	}
	return preset, nil
}

// parseValues reads the values of document into preset.
func parseValues(sandbox *api.Sandbox, document *serializabledeps.SerializableObject, preset *Preset) *privacy.Refusal {
	values, err := document.GetObjectItem("values")
	if err != nil || values == nil || !values.IsObject() {
		return presetRefusal("values", "values is an object holding every setting of the preset.")
	}
	if refusal := onlyKeys(values, "values", PresetSettings...); refusal != nil {
		return refusal
	}
	for _, setting := range PresetSettings {
		field := "values." + setting
		entry, err := values.GetObjectItem(setting)
		if err != nil || entry == nil || !entry.IsObject() {
			return presetRefusal(field, setting+" is missing: a preset holds every setting, {\"value\": …, \"reference\": \"…\"}.")
		}
		if refusal := onlyKeys(entry, field, "value", "reference"); refusal != nil {
			return refusal
		}
		if entry.HasKey("reference") {
			reference, ok := stringOf(entry, "reference")
			if !ok || len(reference) > 1000 {
				return presetRefusal(field+".reference", "A reference is a text of at most 1000 characters.")
			}
			preset.References[setting] = reference
		}
		if !entry.HasKey("value") {
			return presetRefusal(field+".value", setting+" has no value.")
		}
		value, _ := entry.GetObjectItem("value")
		switch setting {
		case SettingIpMode:
			mode, ok := itemString(value)
			if !ok || (mode != privacy.IpModeHashed && mode != privacy.IpModeFull) {
				return presetRefusal(field+".value", "ip-mode is hashed or full.")
			}
			preset.Values.IpMode = mode
		case SettingAccessLog:
			if value == nil || !value.IsBool() {
				return presetRefusal(field+".value", "access-log is true or false.")
			}
			preset.Values.AccessLog, _ = value.GetBool()
		case SettingNoticeDraft:
			draft, ok := itemString(value)
			if !ok || (draft != privacy.DraftBr && draft != privacy.DraftEu && draft != privacy.DraftBrEu) {
				return presetRefusal(field+".value", "notice-draft is br, eu or br-eu.")
			}
			preset.Values.NoticeDraft = draft
		default:
			days := privacy.Forever
			if value != nil && !value.IsNull() {
				if !value.IsInt() {
					return presetRefusal(field+".value", setting+" is a number of days from 1 to 36500, or null to keep the data until a root deletes it.")
				}
				days, _ = value.GetInt()
				if days < 1 || days > privacy.MaxRetentionDays {
					return presetRefusal(field+".value", setting+" is a number of days from 1 to 36500, or null to keep the data until a root deletes it.")
				}
			}
			switch setting {
			case SettingAccessLogDays:
				preset.Values.AccessLogRetention = days
			case SettingIpDays:
				preset.Values.IpRetention = days
			case SettingContactDays:
				preset.Values.ContactRetention = days
			case SettingViewDays:
				preset.Values.ViewRetention = days
			case SettingSummaryDays:
				preset.Values.SummaryRetention = days
			case SettingBackupDays:
				preset.Values.BackupRetention = days
			}
		}
	}
	return nil
}

// parseChecklist reads the checklist of document into preset.
func parseChecklist(sandbox *api.Sandbox, document *serializabledeps.SerializableObject, preset *Preset) *privacy.Refusal {
	if !document.HasKey("checklist") {
		return nil
	}
	list, _ := document.GetObjectItem("checklist")
	size, err := arraySize(list)
	if err != nil || size > maxPresetChecklistItems {
		return presetRefusal("checklist", "checklist is a list of at most 50 items.")
	}
	seen := map[string]bool{}
	for index := 0; index < size; index++ {
		item := list.GetArrayItem(index)
		field := "checklist[" + formatInt(int64(index)) + "]"
		if item == nil || !item.IsObject() {
			return presetRefusal(field, "A checklist item is {\"id\", \"text\", \"reference\", \"when\"}.")
		}
		if refusal := onlyKeys(item, field, "id", "text", "reference", "when"); refusal != nil {
			return refusal
		}
		entry := ChecklistItem{When: ChecklistAlways}
		var ok bool
		entry.Id, ok = stringOf(item, "id")
		if !ok || !matchesPattern(sandbox, `^[a-z0-9][a-z0-9-]{0,39}$`, entry.Id) || seen[entry.Id] {
			return presetRefusal(field+".id", "A checklist item has an id of its own: lowercase letters, digits and dashes.")
		}
		seen[entry.Id] = true
		entry.Text, ok = stringOf(item, "text")
		if !ok || sandbox.Deps.StringsDeps.TrimSpace(entry.Text) == "" || len(entry.Text) > 500 {
			return presetRefusal(field+".text", "A checklist item has a text of 1 to 500 characters.")
		}
		if item.HasKey("reference") {
			if entry.Reference, ok = stringOf(item, "reference"); !ok || len(entry.Reference) > 500 {
				return presetRefusal(field+".reference", "A reference is a text of at most 500 characters.")
			}
		}
		if item.HasKey("when") {
			if entry.When, ok = stringOf(item, "when"); !ok || (entry.When != ChecklistAlways && entry.When != ChecklistAgent) {
				return presetRefusal(field+".when", "when is always or agent.")
			}
		}
		preset.Checklist = append(preset.Checklist, entry)
	}
	return nil
}

// onlyKeys refuses the first key of object that keys does not hold.
func onlyKeys(object *serializabledeps.SerializableObject, field string, keys ...string) *privacy.Refusal {
	present, err := object.GetKeys()
	if err != nil {
		return presetRefusal(field, field+" is not an object.")
	}
	allowed := map[string]bool{}
	for _, key := range keys {
		allowed[key] = true
	}
	for _, key := range present {
		if !allowed[key] {
			return presetRefusal(field+"."+key, "Unknown key "+key+": a preset holds no other settings than the ones Emissario knows.")
		}
	}
	return nil
}

// stringOf is the string at key of object.
func stringOf(object *serializabledeps.SerializableObject, key string) (string, bool) {
	item, err := object.GetObjectItem(key)
	if err != nil {
		return "", false
	}
	return itemString(item)
}

// itemString is item as a string.
func itemString(item *serializabledeps.SerializableObject) (string, bool) {
	if item == nil || !item.IsString() {
		return "", false
	}
	text, err := item.GetString()
	return text, err == nil
}

// intOf is the whole number at key of object.
func intOf(object *serializabledeps.SerializableObject, key string) (int64, bool) {
	item, err := object.GetObjectItem(key)
	if err != nil || item == nil || !item.IsInt() {
		return 0, false
	}
	value, err := item.GetInt()
	return value, err == nil
}

// arraySize is the size of list, an error when it is not a list.
func arraySize(list *serializabledeps.SerializableObject) (int, error) {
	if list == nil || !list.IsArray() {
		return 0, errNotList{}
	}
	return list.GetArraySize()
}

// errNotList is the error of arraySize.
type errNotList struct{}

func (errNotList) Error() string { return "not a list" }

// matchesPattern tells whether text matches pattern.
func matchesPattern(sandbox *api.Sandbox, pattern string, text string) bool {
	ok, err := sandbox.Deps.StringsDeps.MatchPattern(pattern, text)
	return err == nil && ok
}

// PresetJSON is preset as its document: what is stored, exported and
// imported.
func PresetJSON(sandbox *api.Sandbox, preset Preset) *serializabledeps.SerializableObject {
	serial := sandbox.Deps.SerializableDeps
	values := serial.CreateObject()
	for _, setting := range PresetSettings {
		entry := serial.CreateObject()
		value := preset.Values.Value(setting)
		if value == nil {
			entry.AddItemToObject("value", serial.CreateNull())
		} else {
			entry.AddItemToObject("value", value)
		}
		if reference := preset.References[setting]; reference != "" {
			entry.AddItemToObject("reference", reference)
		}
		values.AddItemToObject(setting, entry)
	}
	jurisdictions := serial.CreateArray()
	for _, code := range preset.Jurisdictions {
		jurisdictions.AddItemToArray(code)
	}
	checklist := serial.CreateArray()
	for _, item := range preset.Checklist {
		entry := serial.CreateObject()
		entry.AddItemToObject("id", item.Id)
		entry.AddItemToObject("text", item.Text)
		entry.AddItemToObject("reference", item.Reference)
		entry.AddItemToObject("when", item.When)
		checklist.AddItemToArray(entry)
	}
	document := serial.CreateObject()
	document.AddItemToObject("format", int64(PresetFormat))
	document.AddItemToObject("id", preset.Id)
	document.AddItemToObject("version", preset.Version)
	document.AddItemToObject("name", preset.Name)
	document.AddItemToObject("description", preset.Description)
	document.AddItemToObject("jurisdictions", jurisdictions)
	document.AddItemToObject("values", values)
	document.AddItemToObject("checklist", checklist)
	document.AddItemToObject("changes", preset.Changes)
	return document
}

// BuiltInPresets answers every built-in preset, every version.
func BuiltInPresets(sandbox *api.Sandbox) ([]Preset, error) {
	names, err := sandbox.Deps.EmbedDeps.ListFiles(presetsDir)
	if err != nil {
		return nil, err
	}
	presets := []Preset{}
	for _, name := range names {
		if !sandbox.Deps.StringsDeps.HasSuffix(name, ".json") {
			continue
		}
		content, err := sandbox.Deps.EmbedDeps.ReadFile(presetsDir + "/" + name)
		if err != nil {
			return nil, err
		}
		preset, refusal := ParsePreset(sandbox, string(content))
		if refusal != nil {
			return nil, sandbox.Deps.StdDeps.Errorf("the built-in preset %s is not valid: %s %s", name, refusal.Field, refusal.Message)
		}
		preset.BuiltIn = true
		presets = append(presets, preset)
	}
	return presets, nil
}

// customRecord is a custom preset and the id of its record.
type customRecord struct {
	preset Preset
	id     int64
}

// customPresets answers every custom preset, with its record id. One that no
// longer reads is left out, with a warning.
func customPresets(sandbox *api.Sandbox) ([]customRecord, error) {
	records, err := privacy_db.New(sandbox).ListPrivacyPresets(privacy_db.PrivacyPresetFilter{})
	if err != nil {
		return nil, err
	}
	presets := []customRecord{}
	for _, record := range records {
		preset, refusal := ParsePreset(sandbox, record.Json)
		if refusal != nil {
			sandbox.Deps.StdDeps.Eprintf("warning: the custom preset %s@%d does not read: %s\n", record.PresetId, record.Version, refusal.Message)
			continue
		}
		presets = append(presets, customRecord{preset: preset, id: record.Id})
	}
	return presets, nil
}

// AllPresets answers every preset, the built-in ones first, each by id then
// version.
func AllPresets(sandbox *api.Sandbox) ([]Preset, error) {
	presets, err := BuiltInPresets(sandbox)
	if err != nil {
		return nil, err
	}
	customs, err := customPresets(sandbox)
	if err != nil {
		return nil, err
	}
	for _, custom := range customs {
		presets = append(presets, custom.preset)
	}
	sandbox.Deps.SortDeps.SliceStable(presets, func(i int, j int) bool {
		if presets[i].BuiltIn != presets[j].BuiltIn {
			return presets[i].BuiltIn
		}
		if presets[i].Id != presets[j].Id {
			return presets[i].Id < presets[j].Id
		}
		return presets[i].Version < presets[j].Version
	})
	return presets, nil
}

// FindPreset answers the preset ref names — <id>@<version>, or <id> alone for
// its latest version — false when none is.
func FindPreset(sandbox *api.Sandbox, ref string) (Preset, bool, error) {
	presets, err := AllPresets(sandbox)
	if err != nil {
		return Preset{}, false, err
	}
	ref = sandbox.Deps.StringsDeps.TrimSpace(ref)
	found, ok := Preset{}, false
	for _, preset := range presets {
		if preset.Ref() == ref {
			return preset, true, nil
		}
		if preset.Id == ref && (!ok || preset.Version > found.Version) {
			found, ok = preset, true
		}
	}
	return found, ok, nil
}

// CurrentValues reads the value of every setting a preset holds, as saved.
func CurrentValues(sandbox *api.Sandbox) (PresetValues, error) {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return PresetValues{}, err
	}
	retentions, err := CurrentRetentions(sandbox)
	if err != nil {
		return PresetValues{}, err
	}
	return PresetValues{
		IpMode:             settings.IpMode,
		AccessLog:          settings.AccessLog,
		AccessLogRetention: settings.AccessLogRetentionDays,
		IpRetention:        settings.IpRetentionDays,
		ContactRetention:   retentions.Contact,
		ViewRetention:      retentions.View,
		SummaryRetention:   retentions.Summary,
		BackupRetention:    settings.BackupRetentionDays,
		NoticeDraft:        settings.NoticeDraft,
	}, nil
}

// Difference is one setting whose current value is not the preset's.
type Difference struct {
	Setting string
	Preset  any
	Current any
}

// Differences is every setting of current that differs from preset.
func Differences(current PresetValues, preset PresetValues) []Difference {
	differences := []Difference{}
	for _, setting := range PresetSettings {
		have, want := current.Value(setting), preset.Value(setting)
		if have != want {
			differences = append(differences, Difference{Setting: setting, Preset: want, Current: have})
		}
	}
	return differences
}

// SpellValue spells a value of a setting for a page.
func SpellValue(sandbox *api.Sandbox, value any) string {
	switch typed := value.(type) {
	case nil:
		return "kept until a root deletes it"
	case bool:
		if typed {
			return "on"
		}
		return "off"
	case int64:
		return privacy.DescribeRetention(sandbox, typed)
	case string:
		return typed
	}
	return ""
}

// presetNotices is every notice turning the saved settings into values asks
// for, as a save of each setting on its own page would.
func presetNotices(sandbox *api.Sandbox, values PresetValues) ([]string, error) {
	before, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return nil, err
	}
	current, err := CurrentValues(sandbox)
	if err != nil {
		return nil, err
	}
	after := applied(before, values, "")
	notices := privacy.RequiredNotices(before, after, values.Retentions().Longest())
	notices = privacy.RetentionNotice(notices, privacy.DataContact, current.ContactRetention, values.ContactRetention, 0)
	notices = privacy.RetentionNotice(notices, privacy.DataView, current.ViewRetention, values.ViewRetention, 0)
	notices = privacy.RetentionNotice(notices, privacy.DataSummary, current.SummaryRetention, values.SummaryRetention, 0)
	return notices, nil
}

// applied is settings with the values of a preset, and preset as the one
// applied.
func applied(settings privacy.Settings, values PresetValues, preset string) privacy.Settings {
	settings.IpMode = values.IpMode
	settings.AccessLog = values.AccessLog
	settings.AccessLogRetentionDays = values.AccessLogRetention
	settings.IpRetentionDays = values.IpRetention
	settings.BackupRetentionDays = values.BackupRetention
	settings.NoticeDraft = values.NoticeDraft
	settings.Preset = preset
	return settings
}

// LoadPresetNotices loads the notices of ids, each filled for values.
func LoadPresetNotices(sandbox *api.Sandbox, values PresetValues, ids []string) ([]privacy.Notice, error) {
	host := ""
	if agentSettings, err := agent.ReadSettings(sandbox); err == nil {
		host = agent.ProviderHost(sandbox, agentSettings)
	}
	retentions := values.Retentions()
	notices := []privacy.Notice{}
	for _, id := range ids {
		context := privacy.NoticeContext{Host: host}
		switch id {
		case privacy.RetentionNoticeId(privacy.DataContact):
			context.Retention = values.ContactRetention
		case privacy.RetentionNoticeId(privacy.DataView):
			context.Retention = values.ViewRetention
		case privacy.RetentionNoticeId(privacy.DataSummary):
			context.Retention = values.SummaryRetention
		case privacy.RetentionNoticeId(privacy.DataAccessLog):
			context.Retention = values.AccessLogRetention
		case privacy.RetentionNoticeId(privacy.DataIp):
			context.Retention = values.IpRetention
		case privacy.RetentionNoticeId(privacy.DataBackup):
			context.Retention = values.BackupRetention
			context.Longest = retentions.Longest()
		}
		notice, err := privacy.LoadNotice(sandbox, id, context)
		if err != nil {
			return nil, err
		}
		notices = append(notices, notice)
	}
	return notices, nil
}

// Preview is what applying a preset would do.
type Preview struct {
	Preset      Preset
	Current     PresetValues
	Differences []Difference
	// Notices are the notices applying it asks for.
	Notices []string
	// Deletes is what the next cleanup would delete under its retentions,
	// counted now by the same rules.
	Deletes Swept
	// Effects says in sentences what happens to the data stored.
	Effects []string
}

// PreviewPreset answers what applying preset would do, writing nothing.
func PreviewPreset(sandbox *api.Sandbox, preset Preset) (Preview, error) {
	current, err := CurrentValues(sandbox)
	if err != nil {
		return Preview{}, err
	}
	notices, err := presetNotices(sandbox, preset.Values)
	if err != nil {
		return Preview{}, err
	}
	deletes, err := Sweep(sandbox, preset.Values.Retentions(), false)
	if err != nil {
		return Preview{}, err
	}
	preview := Preview{Preset: preset, Current: current, Differences: Differences(current, preset.Values), Notices: notices, Deletes: deletes}
	plural := func(count int, one string, many string) string {
		if count == 1 {
			return "1 " + one
		}
		return sandbox.Deps.StringsDeps.FormatInt(int64(count), 10) + " " + many
	}
	if deletes.Total() > 0 {
		preview.Effects = append(preview.Effects, "The next cleanup deletes "+plural(deletes.Contacts, "contact request", "contact requests")+", "+plural(deletes.Views, "view", "views")+", "+plural(deletes.Summaries, "agent summary", "agent summaries")+", "+plural(deletes.AccessLog, "access log row", "access log rows")+" and "+plural(deletes.Snapshots, "snapshot", "snapshots")+", and blanks "+plural(deletes.ContactIps+deletes.ViewIps+deletes.UsageIps, "IP", "IPs")+".")
	}
	switch {
	case preset.Values.AccessLog && !current.AccessLog:
		preview.Effects = append(preview.Effects, "From now on every view and public API call writes an access log row.")
	case !preset.Values.AccessLog && current.AccessLog:
		stored, err := privacy_db.New(sandbox).CountAccessLog()
		if err != nil {
			return preview, err
		}
		preview.Effects = append(preview.Effects, "No new access log row; the rows kept ("+sandbox.Deps.StringsDeps.FormatInt(int64(stored), 10)+") stay until their retention ends.")
	}
	switch {
	case preset.Values.IpMode == privacy.IpModeHashed && current.IpMode == privacy.IpModeFull:
		preview.Effects = append(preview.Effects, "New IPs are hashed; the full IPs stored stay until their retention ends, or until Hash the IPs already stored.")
	case preset.Values.IpMode == privacy.IpModeFull && current.IpMode == privacy.IpModeHashed:
		preview.Effects = append(preview.Effects, "Only new IPs are stored in full; the hashes stored stay hashes.")
	}
	return preview, nil
}

// ApplyPreset saves every setting of the preset ref names, after the notices
// they need: until acknowledge holds them all the refusal is a 409 listing
// them, and nothing is written. It never changes whether data is collected
// — the contact requests, the agent and its summaries — and keeps
// privacy-setting.preset = <id>@<version>, nothing about who applied it nor
// when.
func ApplyPreset(sandbox *api.Sandbox, ref string, acknowledge []string) (Preset, *privacy.Refusal, error) {
	preset, ok, err := FindPreset(sandbox, ref)
	if err != nil {
		return preset, nil, err
	}
	if !ok {
		return preset, &privacy.Refusal{Status: api.StatusNotFound, Field: "preset", Message: "No preset is named " + ref + "."}, nil
	}
	notices, err := presetNotices(sandbox, preset.Values)
	if err != nil {
		return preset, nil, err
	}
	if missing := privacy.NeedsAcknowledgement(notices, acknowledge); missing != nil {
		return preset, missing, nil
	}
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return preset, nil, err
	}
	if err := privacy.SaveSettings(sandbox, applied(settings, preset.Values, preset.Ref())); err != nil {
		return preset, nil, err
	}
	if err := contact.SaveRetention(sandbox, preset.Values.ContactRetention); err != nil {
		return preset, nil, err
	}
	if err := visualization.SaveRetention(sandbox, preset.Values.ViewRetention); err != nil {
		return preset, nil, err
	}
	if err := agent.SaveSummaryRetention(sandbox, preset.Values.SummaryRetention); err != nil {
		return preset, nil, err
	}
	InvalidateBanners(sandbox)
	return preset, nil, nil
}

// DetachPreset forgets the preset applied: the server is Custom, its
// settings as they are.
func DetachPreset(sandbox *api.Sandbox) error {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	settings.Preset = ""
	if err := privacy.SaveSettings(sandbox, settings); err != nil {
		return err
	}
	InvalidateBanners(sandbox)
	return nil
}

// PresetState is the state check-privacy-preset answers: the preset applied
// — "" for none — whether every setting still equals it, and the ones that
// do not.
type PresetState struct {
	Preset      string
	Found       bool
	Conform     bool
	Differences []Difference
	// Newer is the newest version of the preset applied, when it is newer.
	Newer *Preset
}

// CheckPreset answers the state of the preset applied.
func CheckPreset(sandbox *api.Sandbox) (PresetState, error) {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return PresetState{}, err
	}
	state := PresetState{Preset: settings.Preset}
	if settings.Preset == "" {
		return state, nil
	}
	preset, ok, err := FindPreset(sandbox, settings.Preset)
	if err != nil || !ok || preset.Ref() != settings.Preset {
		return state, err
	}
	state.Found = true
	current, err := CurrentValues(sandbox)
	if err != nil {
		return state, err
	}
	state.Differences = Differences(current, preset.Values)
	state.Conform = len(state.Differences) == 0
	latest, ok, err := FindPreset(sandbox, preset.Id)
	if err != nil {
		return state, err
	}
	if ok && latest.Version > preset.Version {
		state.Newer = &latest
	}
	return state, nil
}

// slugOf is name as a preset id: lowercase letters and digits, every other
// run of characters a dash, at most 40 characters; "custom" for none.
func slugOf(name string) string {
	slug := []byte{}
	dash := false
	for index := 0; index < len(name); index++ {
		char := name[index]
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A'
		}
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			if dash && len(slug) > 0 {
				slug = append(slug, '-')
			}
			slug = append(slug, char)
			dash = false
			continue
		}
		dash = true
	}
	if len(slug) > 40 {
		slug = slug[:40]
	}
	for len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}
	if len(slug) == 0 {
		return "custom"
	}
	return string(slug)
}

// builtInId tells whether id is the id of a built-in preset.
func builtInId(sandbox *api.Sandbox, id string) (bool, error) {
	presets, err := BuiltInPresets(sandbox)
	if err != nil {
		return false, err
	}
	for _, preset := range presets {
		if preset.Id == id {
			return true, nil
		}
	}
	return false, nil
}

// storePreset adds preset to privacy-preset.
func storePreset(sandbox *api.Sandbox, preset Preset) error {
	_, err := privacy_db.New(sandbox).AddPrivacyPreset(privacy_db.PrivacyPresetInput{
		PresetId:  preset.Id,
		Version:   preset.Version,
		Name:      preset.Name,
		Json:      sandbox.Deps.SerializableDeps.SerializeToJson(PresetJSON(sandbox, preset)),
		CreatedAt: privacy.NowSeconds(sandbox),
	})
	return err
}

// SavePreset saves the current settings as a custom preset named name: a new
// id, or the next version of the custom preset of that name — a version
// never changes. Its checklist is the one of the preset applied, the
// lgpd-gdpr one when there is none; references are left for the root to
// fill in an export. It is stored, not applied.
func SavePreset(sandbox *api.Sandbox, name string, description string) (Preset, *privacy.Refusal, error) {
	strings := sandbox.Deps.StringsDeps
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" || len(name) > 120 {
		return Preset{}, presetRefusal("name", "The name is 1 to 120 characters."), nil
	}
	if len(description) > 1000 {
		return Preset{}, presetRefusal("description", "The description is at most 1000 characters."), nil
	}
	id := slugOf(name)
	builtIn, err := builtInId(sandbox, id)
	if err != nil {
		return Preset{}, nil, err
	}
	if builtIn {
		return Preset{}, presetRefusal("name", "This name is the one of a built-in preset, which is read only: pick another one."), nil
	}
	values, err := CurrentValues(sandbox)
	if err != nil {
		return Preset{}, nil, err
	}
	preset := Preset{Id: id, Version: 1, Name: name, Description: description, Values: values, References: map[string]string{}, Changes: "First version."}
	if latest, ok, err := FindPreset(sandbox, id); err != nil {
		return preset, nil, err
	} else if ok {
		preset.Version = latest.Version + 1
		preset.Jurisdictions = latest.Jurisdictions
		preset.Changes = "Saved again from the settings of a server."
	}
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return preset, nil, err
	}
	source := DefaultChecklistPreset
	if settings.Preset != "" {
		source = settings.Preset
	}
	if from, ok, err := FindPreset(sandbox, source); err != nil {
		return preset, nil, err
	} else if ok {
		preset.Checklist = from.Checklist
		if len(preset.Jurisdictions) == 0 {
			preset.Jurisdictions = from.Jurisdictions
		}
	}
	return preset, nil, storePreset(sandbox, preset)
}

// ImportPreset checks text as a preset — its format, every setting in range,
// no unknown key — and stores it without applying it. A built-in id, or a
// version already stored, is refused: versions never change.
func ImportPreset(sandbox *api.Sandbox, text string) (Preset, *privacy.Refusal, error) {
	preset, refusal := ParsePreset(sandbox, text)
	if refusal != nil {
		return preset, refusal, nil
	}
	builtIn, err := builtInId(sandbox, preset.Id)
	if err != nil {
		return preset, nil, err
	}
	if builtIn {
		return preset, presetRefusal("id", preset.Id+" is a built-in preset, which is read only: give the copy an id of its own."), nil
	}
	if _, ok, err := findExact(sandbox, preset.Ref()); err != nil {
		return preset, nil, err
	} else if ok {
		return preset, &privacy.Refusal{Status: api.StatusConflict, Field: "version", Message: preset.Ref() + " is already stored, and a version never changes: give it the next version."}, nil
	}
	return preset, nil, storePreset(sandbox, preset)
}

// findExact answers the preset of ref, <id>@<version> exactly.
func findExact(sandbox *api.Sandbox, ref string) (Preset, bool, error) {
	preset, ok, err := FindPreset(sandbox, ref)
	if err != nil || !ok || preset.Ref() != ref {
		return Preset{}, false, err
	}
	return preset, true, nil
}

// DeletePreset deletes the custom preset of ref, <id>@<version>. A built-in
// preset, or the one applied, is refused.
func DeletePreset(sandbox *api.Sandbox, ref string) (*privacy.Refusal, error) {
	customs, err := customPresets(sandbox)
	if err != nil {
		return nil, err
	}
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return nil, err
	}
	for _, custom := range customs {
		if custom.preset.Ref() != ref {
			continue
		}
		if settings.Preset == ref {
			return presetRefusal("preset", "This preset is applied: detach it, or apply another one, first."), nil
		}
		InvalidateBanners(sandbox)
		return nil, privacy_db.New(sandbox).RemovePrivacyPreset(custom.id)
	}
	if exact, ok, err := findExact(sandbox, ref); err != nil {
		return nil, err
	} else if ok && exact.BuiltIn {
		return presetRefusal("preset", "A built-in preset is read only."), nil
	}
	return &privacy.Refusal{Status: api.StatusNotFound, Field: "preset", Message: "No custom preset is named " + ref + "."}, nil
}

// VisibleChecklist is the checklist of preset as the Presets page lists it:
// the items about the agent only while it is on.
func VisibleChecklist(sandbox *api.Sandbox, preset Preset) []ChecklistItem {
	agentOn := false
	if agentSettings, err := agent.ReadSettings(sandbox); err == nil {
		agentOn = agentSettings.Enabled
	}
	items := []ChecklistItem{}
	for _, item := range preset.Checklist {
		if item.When == ChecklistAgent && !agentOn {
			continue
		}
		items = append(items, item)
	}
	return items
}

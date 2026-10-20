package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The paths of the Presets page's forms.
const (
	PresetApplyPath  = "/admin/root/apply-privacy-preset"
	PresetSavePath   = "/admin/root/save-privacy-preset"
	PresetImportPath = "/admin/root/import-privacy-preset"
	PresetExportPath = "/admin/root/export-privacy-preset"
	PresetDeletePath = "/admin/root/delete-privacy-preset"
	PresetDetachPath = "/admin/root/detach-privacy-preset"
)

// The outcome codes the Presets page shows after a form.
const (
	PresetNoticeApplied  = "applied"
	PresetNoticeDetached = "detached"
	PresetNoticeSaved    = "saved"
	PresetNoticeImported = "imported"
	PresetNoticeDeleted  = "deleted"
	PresetNoticeRefused  = "refused"
)

// PresetLine is one preset of the Presets page.
type PresetLine struct {
	Ref           string
	Name          string
	Description   string
	Jurisdictions string
	BuiltIn       bool
	Active        bool
	PreviewHref   string
	ExportHref    string
}

// DifferenceLine is one setting on a Presets page: its value now and in the
// preset, and the preset's reference.
type DifferenceLine struct {
	Setting   string
	Current   string
	Preset    string
	Reference string
	Differs   bool
}

// PrivacyPresetsPage is what backoffice/privacy_presets.html is rendered
// with.
type PrivacyPresetsPage struct {
	Viewer Viewer
	Notice Notice
	// Active is the preset applied, "" for none; ActiveName its name and
	// Found false when it no longer exists.
	Active      string
	ActiveName  string
	Found       bool
	Conform     bool
	Differences []DifferenceLine
	// Newer is a newer version of the preset applied, "" for none, spelled
	// by its name, version and changes.
	Newer        string
	NewerHref    string
	NewerName    string
	NewerVersion string
	NewerChanges string
	Checklist   []privacyadmin.ChecklistItem
	Presets     []PresetLine
	ReapplyHref string
	DetachPath  string
	SavePath    string
	ImportPath  string
	DeletePath  string
	PrivacyHref string
}

// presetsNotice is how the Presets page words an outcome code.
func presetsNotice(sandbox *api.Sandbox, code string, preset string, reason string) Notice {
	switch code {
	case PresetNoticeApplied:
		return Notice{Text: "Preset " + preset + " applied: every setting it holds now equals it.", Kind: "ok"}
	case PresetNoticeDetached:
		return Notice{Text: "Preset detached: this server is Custom, its settings as they were.", Kind: "ok"}
	case PresetNoticeSaved:
		return Notice{Text: "The current settings were saved as " + preset + ". Export it to configure other servers the same way.", Kind: "ok"}
	case PresetNoticeImported:
		return Notice{Text: "Preset " + preset + " imported. Preview it to apply it.", Kind: "ok"}
	case PresetNoticeDeleted:
		return Notice{Text: "Preset " + preset + " deleted.", Kind: "ok"}
	case PresetNoticeRefused:
		return Notice{Text: reason, Kind: "error"}
	}
	return Notice{}
}

// RenderPrivacyPresetsPage answers the Presets page for user, with the
// outcome code notice of the last action, about preset, and its reason when
// it was refused.
func RenderPrivacyPresetsPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, notice string, preset string, reason string) error {
	state, err := privacyadmin.CheckPreset(sandbox)
	if err != nil {
		return err
	}
	presets, err := privacyadmin.AllPresets(sandbox)
	if err != nil {
		return err
	}
	page := PrivacyPresetsPage{
		Viewer:      viewerOf(sandbox, user),
		Notice:      presetsNotice(sandbox, notice, preset, reason),
		Active:      state.Preset,
		Found:       state.Found,
		Conform:     state.Conform,
		DetachPath:  PresetDetachPath,
		SavePath:    PresetSavePath,
		ImportPath:  PresetImportPath,
		DeletePath:  PresetDeletePath,
		PrivacyHref: privacy.SettingsPath,
	}
	for _, item := range presets {
		jurisdictions := sandbox.Deps.StringsDeps.Join(item.Jurisdictions, ", ")
		page.Presets = append(page.Presets, PresetLine{
			Ref:           item.Ref(),
			Name:          item.Name,
			Description:   item.Description,
			Jurisdictions: jurisdictions,
			BuiltIn:       item.BuiltIn,
			Active:        item.Ref() == state.Preset,
			PreviewHref:   visualization.Href(sandbox, privacy.PresetPreviewPth, []visualization.Param{{Key: "preset", Value: item.Ref()}}),
			ExportHref:    visualization.Href(sandbox, PresetExportPath, []visualization.Param{{Key: "preset", Value: item.Ref()}}),
		})
		if item.Ref() == state.Preset {
			page.ActiveName = item.Name
			page.Checklist = privacyadmin.VisibleChecklist(sandbox, item)
			page.ReapplyHref = visualization.Href(sandbox, privacy.PresetPreviewPth, []visualization.Param{{Key: "preset", Value: item.Ref()}})
		}
	}
	for _, difference := range state.Differences {
		page.Differences = append(page.Differences, DifferenceLine{Setting: difference.Setting, Current: privacyadmin.SpellValue(sandbox, difference.Current), Preset: privacyadmin.SpellValue(sandbox, difference.Preset), Differs: true})
	}
	if state.Newer != nil {
		page.Newer = state.Newer.Ref()
		page.NewerName, page.NewerChanges = state.Newer.Name, state.Newer.Changes
		page.NewerVersion = sandbox.Deps.StringsDeps.FormatInt(state.Newer.Version, 10)
		page.NewerHref = visualization.Href(sandbox, privacy.PresetPreviewPth, []visualization.Param{{Key: "preset", Value: state.Newer.Ref()}})
	}
	return RenderHTML(sandbox, response, status, "backoffice/privacy_presets.html", page)
}

// PresetPreviewPage is what backoffice/privacy_preset_preview.html is
// rendered with.
type PresetPreviewPage struct {
	Viewer      Viewer
	Error       string
	Ref         string
	Name        string
	Description string
	Changes     string
	BuiltIn     bool
	Lines       []DifferenceLine
	Effects     []string
	// Deletes is true when applying it deletes data at the next cleanup.
	Deletes     bool
	Notices     []privacy.Notice
	ApplyAction string
	PresetsHref string
}

// RenderPresetPreviewPage answers, under status, the preview of preset with
// message above it: every setting now and in the preset with its reference,
// what happens to the data stored, and the notices applying it asks for,
// each to tick, beside Apply.
func RenderPresetPreviewPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, preset privacyadmin.Preset, message string) error {
	preview, err := privacyadmin.PreviewPreset(sandbox, preset)
	if err != nil {
		return err
	}
	notices, err := privacyadmin.LoadPresetNotices(sandbox, preset.Values, preview.Notices)
	if err != nil {
		return err
	}
	page := PresetPreviewPage{
		Viewer:      viewerOf(sandbox, user),
		Error:       message,
		Ref:         preset.Ref(),
		Name:        preset.Name,
		Description: preset.Description,
		Changes:     preset.Changes,
		BuiltIn:     preset.BuiltIn,
		Effects:     preview.Effects,
		Deletes:     preview.Deletes.Total() > 0,
		Notices:     notices,
		ApplyAction: PresetApplyPath,
		PresetsHref: privacy.PresetsPath,
	}
	for _, setting := range privacyadmin.PresetSettings {
		current, wanted := preview.Current.Value(setting), preset.Values.Value(setting)
		page.Lines = append(page.Lines, DifferenceLine{
			Setting:   setting,
			Current:   privacyadmin.SpellValue(sandbox, current),
			Preset:    privacyadmin.SpellValue(sandbox, wanted),
			Reference: preset.References[setting],
			Differs:   current != wanted,
		})
	}
	return RenderHTML(sandbox, response, status, "backoffice/privacy_preset_preview.html", page)
}

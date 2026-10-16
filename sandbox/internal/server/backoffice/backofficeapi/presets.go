package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
)

// presetValueJSON is a value of a preset setting as JSON spells it.
func presetValueJSON(sandbox *api.Sandbox, value any) any {
	if value == nil {
		return sandbox.Deps.SerializableDeps.CreateNull()
	}
	return value
}

// PresetsJSON is {active, presets: [{preset, id, version, name, description,
// jurisdictions, built_in, active}]}: every preset, every version.
func PresetsJSON(sandbox *api.Sandbox, presets []privacyadmin.Preset, active string) *serializabledeps.SerializableObject {
	serial := sandbox.Deps.SerializableDeps
	list := serial.CreateArray()
	for _, preset := range presets {
		jurisdictions := serial.CreateArray()
		for _, code := range preset.Jurisdictions {
			jurisdictions.AddItemToArray(code)
		}
		item := serial.CreateObject()
		item.AddItemToObject("preset", preset.Ref())
		item.AddItemToObject("id", preset.Id)
		item.AddItemToObject("version", preset.Version)
		item.AddItemToObject("name", preset.Name)
		item.AddItemToObject("description", preset.Description)
		item.AddItemToObject("jurisdictions", jurisdictions)
		item.AddItemToObject("built_in", preset.BuiltIn)
		item.AddItemToObject("active", preset.Ref() == active)
		list.AddItemToArray(item)
	}
	document := serial.CreateObject()
	document.AddItemToObject("active", active)
	document.AddItemToObject("presets", list)
	return document
}

// DifferencesJSON is [{setting, preset, current}].
func DifferencesJSON(sandbox *api.Sandbox, differences []privacyadmin.Difference) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, difference := range differences {
		item := sandbox.Deps.SerializableDeps.CreateObject()
		item.AddItemToObject("setting", difference.Setting)
		item.AddItemToObject("preset", presetValueJSON(sandbox, difference.Preset))
		item.AddItemToObject("current", presetValueJSON(sandbox, difference.Current))
		list.AddItemToArray(item)
	}
	return list
}

// PreviewJSON is {preset, differences, notices, deletes: {contact_requests,
// views, agent_summaries, access_log, snapshots, ips}, effects}: what
// applying a preset would do.
func PreviewJSON(sandbox *api.Sandbox, preview privacyadmin.Preview) *serializabledeps.SerializableObject {
	serial := sandbox.Deps.SerializableDeps
	notices := serial.CreateArray()
	for _, id := range preview.Notices {
		notices.AddItemToArray(id)
	}
	deletes := serial.CreateObject()
	deletes.AddItemToObject("contact_requests", preview.Deletes.Contacts)
	deletes.AddItemToObject("views", preview.Deletes.Views)
	deletes.AddItemToObject("agent_summaries", preview.Deletes.Summaries)
	deletes.AddItemToObject("access_log", preview.Deletes.AccessLog)
	deletes.AddItemToObject("snapshots", preview.Deletes.Snapshots)
	deletes.AddItemToObject("ips", preview.Deletes.ContactIps+preview.Deletes.ViewIps+preview.Deletes.UsageIps)
	effects := serial.CreateArray()
	for _, effect := range preview.Effects {
		effects.AddItemToArray(effect)
	}
	document := serial.CreateObject()
	document.AddItemToObject("preset", preview.Preset.Ref())
	document.AddItemToObject("differences", DifferencesJSON(sandbox, preview.Differences))
	document.AddItemToObject("notices", notices)
	document.AddItemToObject("deletes", deletes)
	document.AddItemToObject("effects", effects)
	return document
}

// PresetStateJSON is {preset, conform, differences, encrypted, newer}: the
// state check-privacy-preset answers; preset is "" for none, newer the
// newest version of it when there is one, null otherwise.
func PresetStateJSON(sandbox *api.Sandbox, state privacyadmin.PresetState) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("preset", state.Preset)
	document.AddItemToObject("conform", state.Conform)
	document.AddItemToObject("differences", DifferencesJSON(sandbox, state.Differences))
	document.AddItemToObject("encrypted", sandbox.Config.Encrypted)
	if state.Newer != nil {
		document.AddItemToObject("newer", state.Newer.Ref())
	} else {
		document.AddItemToObject("newer", sandbox.Deps.SerializableDeps.CreateNull())
	}
	return document
}

// PresetRefJSON is {preset}: the preset an action touched.
func PresetRefJSON(sandbox *api.Sandbox, preset privacyadmin.Preset) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("preset", preset.Ref())
	return document
}

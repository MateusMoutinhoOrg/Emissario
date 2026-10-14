package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// ContactRequestJSON is one contact request: {id, at, ip, country, archived,
// size, body, body_text}. body is the JSON value it holds — the text as sent,
// were it ever not to parse — read back the way JSON is: its keys sorted and
// its numbers as doubles, so an integer past 2^53 is rounded; body_text is
// the text exactly as it was sent.
func ContactRequestJSON(sandbox *api.Sandbox, request contact_db.ContactRequestRecord) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", request.Ref)
	item.AddItemToObject("at", request.At)
	item.AddItemToObject("ip", request.Ip)
	item.AddItemToObject("country", request.Country)
	item.AddItemToObject("archived", request.Archived != 0)
	item.AddItemToObject("size", request.Size)
	if body, err := sandbox.Deps.SerializableDeps.ParseJson(request.Body); err == nil {
		item.AddItemToObject("body", body)
	} else {
		item.AddItemToObject("body", request.Body)
	}
	item.AddItemToObject("body_text", request.Body)
	return item
}

// ContactRequestResponseJSON is {contact_request}.
func ContactRequestResponseJSON(sandbox *api.Sandbox, request contact_db.ContactRequestRecord) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("contact_request", ContactRequestJSON(sandbox, request))
	return document
}

// ContactRequestsJSON is one page of the contact requests: {contact_requests,
// total, page, pages, limit}.
func ContactRequestsJSON(sandbox *api.Sandbox, requests []contact_db.ContactRequestRecord, paging visualization.Paging) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, request := range requests {
		list.AddItemToArray(ContactRequestJSON(sandbox, request))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("contact_requests", list)
	pagingJSON(sandbox, document, paging)
	return document
}

// ChangedJSON is {changed}: how many contact requests an archive changed.
func ChangedJSON(sandbox *api.Sandbox, changed int) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("changed", changed)
	return document
}

// ContactSettingsJSON is the settings of the contact requests.
func ContactSettingsJSON(sandbox *api.Sandbox, settings contact.Settings) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("enabled", settings.Enabled)
	item.AddItemToObject("max_bytes", settings.MaxBytes)
	item.AddItemToObject("rate_per_minute", settings.RatePerMinute)
	item.AddItemToObject("ip_daily_limit", settings.IpDailyLimit)
	return item
}

// ContactSettingsResponseJSON is {settings}.
func ContactSettingsResponseJSON(sandbox *api.Sandbox, settings contact.Settings) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", ContactSettingsJSON(sandbox, settings))
	return document
}

// ContactSettingsStateJSON is {settings, stored, inbox, running, dropped}.
func ContactSettingsStateJSON(sandbox *api.Sandbox, settings contact.Settings, state contact.State) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", ContactSettingsJSON(sandbox, settings))
	document.AddItemToObject("stored", state.Stored)
	document.AddItemToObject("inbox", state.Inbox)
	document.AddItemToObject("running", state.Running)
	document.AddItemToObject("dropped", state.Dropped)
	return document
}

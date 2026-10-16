package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// AccessRowJSON is one row of the access log: {id, at, at_utc, at_local,
// time_zone, ip, port, kind, ref}. at is in Unix milliseconds; a port of 0
// is null, one not read.
func AccessRowJSON(sandbox *api.Sandbox, row privacy_db.AccessLogRecord) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", row.Id)
	item.AddItemToObject("at", row.At)
	item.AddItemToObject("at_utc", privacyadmin.FormatUtc(sandbox, row.At))
	item.AddItemToObject("at_local", visualization.FormatLocal(sandbox, row.At/1000, visualization.TimeLayout))
	item.AddItemToObject("time_zone", visualization.ZoneName(sandbox))
	item.AddItemToObject("ip", row.Ip)
	if row.Port > 0 {
		item.AddItemToObject("port", row.Port)
	} else {
		item.AddItemToObject("port", sandbox.Deps.SerializableDeps.CreateNull())
	}
	item.AddItemToObject("kind", row.Kind)
	item.AddItemToObject("ref", row.Ref)
	return item
}

// AccessRowsJSON is every one of rows.
func AccessRowsJSON(sandbox *api.Sandbox, rows []privacy_db.AccessLogRecord) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, row := range rows {
		list.AddItemToArray(AccessRowJSON(sandbox, row))
	}
	return list
}

// AccessListJSON is one page of the access log: {rows, total, page, pages,
// limit}.
func AccessListJSON(sandbox *api.Sandbox, rows []privacy_db.AccessLogRecord, paging visualization.Paging) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("rows", AccessRowsJSON(sandbox, rows))
	document.AddItemToObject("total", paging.Total)
	document.AddItemToObject("page", paging.Page)
	document.AddItemToObject("pages", paging.Pages)
	document.AddItemToObject("limit", paging.Limit)
	return document
}

// PrivacyAccessJSON is one read of the access log the inventory kept.
func PrivacyAccessJSON(sandbox *api.Sandbox, access privacy_db.PrivacyAccessRecord) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", access.Id)
	item.AddItemToObject("at", access.At)
	item.AddItemToObject("at_utc", privacyadmin.FormatUtc(sandbox, access.At))
	item.AddItemToObject("ended_at", access.EndedAt)
	item.AddItemToObject("user_id", access.UserId)
	item.AddItemToObject("username", access.Username)
	item.AddItemToObject("action", access.Action)
	if filter, err := sandbox.Deps.SerializableDeps.ParseJson(access.Filter); err == nil {
		item.AddItemToObject("filter", filter)
	} else {
		item.AddItemToObject("filter", access.Filter)
	}
	item.AddItemToObject("rows", access.Rows)
	return item
}

// ReadersJSON is {readers: [{id, username, totp}]}: the roots who read the
// access log.
func ReadersJSON(sandbox *api.Sandbox, users []backoffice_db.BackofficeUserRecord) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, user := range users {
		if user.AccessLogReader != 1 {
			continue
		}
		item := sandbox.Deps.SerializableDeps.CreateObject()
		item.AddItemToObject("id", user.Id)
		item.AddItemToObject("username", user.Username)
		item.AddItemToObject("totp", totp.Enabled(user))
		list.AddItemToArray(item)
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("readers", list)
	return document
}

// HoldJSON is one legal hold: {id, table, ids, reason, until}, until the
// instant it ends in Unix seconds.
func HoldJSON(sandbox *api.Sandbox, hold privacy.Hold) *serializabledeps.SerializableObject {
	ids := sandbox.Deps.SerializableDeps.CreateArray()
	for _, id := range hold.Ids {
		ids.AddItemToArray(id)
	}
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", hold.Id)
	item.AddItemToObject("table", hold.Table)
	item.AddItemToObject("ids", ids)
	item.AddItemToObject("reason", hold.Reason)
	item.AddItemToObject("until", hold.Until)
	return item
}

// HoldsJSON is {holds: […]}.
func HoldsJSON(sandbox *api.Sandbox, holds []privacy.Hold) *serializabledeps.SerializableObject {
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, hold := range holds {
		list.AddItemToArray(HoldJSON(sandbox, hold))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("holds", list)
	return document
}

// PrivacySettingsJSON is the privacy settings: {ip_mode, viewer_ip,
// access_log, access_log_retention_days, access_log_totp,
// access_log_inventory, ip_retention_days, backup_retention_days,
// notice_draft, preset}; a retention of null keeps the data until a root
// deletes it, and a preset of "" is none.
func PrivacySettingsJSON(sandbox *api.Sandbox, settings privacy.Settings) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("ip_mode", settings.IpMode)
	item.AddItemToObject("viewer_ip", settings.ViewerIp)
	item.AddItemToObject("access_log", settings.AccessLog)
	item.AddItemToObject("access_log_retention_days", RetentionJSON(sandbox, settings.AccessLogRetentionDays))
	item.AddItemToObject("access_log_totp", settings.AccessLogTotp)
	item.AddItemToObject("access_log_inventory", settings.AccessLogInventory)
	item.AddItemToObject("ip_retention_days", RetentionJSON(sandbox, settings.IpRetentionDays))
	item.AddItemToObject("backup_retention_days", RetentionJSON(sandbox, settings.BackupRetentionDays))
	item.AddItemToObject("notice_draft", settings.NoticeDraft)
	item.AddItemToObject("preset", settings.Preset)
	return item
}

// PrivacySettingsResponseJSON is {settings}.
func PrivacySettingsResponseJSON(sandbox *api.Sandbox, settings privacy.Settings) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", PrivacySettingsJSON(sandbox, settings))
	return document
}

// PrivacyStateJSON is {settings, encryption: {encrypted, algorithm, key_id},
// ip_key_made_at, kept: [{table, fields, purpose, basis, retention, rows,
// oldest_at}]}: what the Privacy page shows. key_id reveals nothing about
// the key; oldest_at is null for a table with no row or no time.
func PrivacyStateJSON(sandbox *api.Sandbox, settings privacy.Settings, ipKeyMadeAt int64, kept []privacyadmin.Kept) *serializabledeps.SerializableObject {
	encryption := sandbox.Deps.SerializableDeps.CreateObject()
	encryption.AddItemToObject("encrypted", sandbox.Config.Encrypted)
	encryption.AddItemToObject("algorithm", "XChaCha20-Poly1305")
	encryption.AddItemToObject("key_id", sandbox.Config.EncryptionKeyId)
	tables := sandbox.Deps.SerializableDeps.CreateArray()
	for _, table := range kept {
		item := sandbox.Deps.SerializableDeps.CreateObject()
		item.AddItemToObject("table", table.Table)
		item.AddItemToObject("fields", table.Fields)
		item.AddItemToObject("purpose", table.Purpose)
		item.AddItemToObject("basis", table.Basis)
		item.AddItemToObject("retention", table.Retention)
		item.AddItemToObject("rows", table.Count)
		if table.Oldest > 0 {
			item.AddItemToObject("oldest_at", table.Oldest)
		} else {
			item.AddItemToObject("oldest_at", sandbox.Deps.SerializableDeps.CreateNull())
		}
		tables.AddItemToArray(item)
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", PrivacySettingsJSON(sandbox, settings))
	document.AddItemToObject("encryption", encryption)
	document.AddItemToObject("ip_key_made_at", ipKeyMadeAt)
	document.AddItemToObject("kept", tables)
	return document
}

// HashedIpsJSON is {hashed: {views, contact_requests, agent_usage, total}}.
func HashedIpsJSON(sandbox *api.Sandbox, hashed privacyadmin.HashedIps) *serializabledeps.SerializableObject {
	counts := sandbox.Deps.SerializableDeps.CreateObject()
	counts.AddItemToObject("views", hashed.Views)
	counts.AddItemToObject("contact_requests", hashed.Contacts)
	counts.AddItemToObject("agent_usage", hashed.Usages)
	counts.AddItemToObject("total", hashed.Total())
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("hashed", counts)
	return document
}

// SubjectJSON is every record of a person Find and erase found:
// {contact_requests, views, agent_usage, agent_summaries, access_log, held,
// kept_by_law, ip_key_made_at, total}. access_log is null for a root who does
// not read it; held lists the ids of each table on a legal hold — never the
// reason — and kept_by_law the access log rows still inside their term.
func SubjectJSON(sandbox *api.Sandbox, found privacyadmin.Found) *serializabledeps.SerializableObject {
	serial := sandbox.Deps.SerializableDeps
	contacts := serial.CreateArray()
	for _, request := range found.Contacts {
		contacts.AddItemToArray(ContactRequestJSON(sandbox, request, true))
	}
	views := serial.CreateArray()
	for _, view := range found.Views {
		views.AddItemToArray(VisualizationViewJSON(sandbox, view))
	}
	usages := serial.CreateArray()
	for _, usage := range found.Usages {
		item := serial.CreateObject()
		item.AddItemToObject("id", usage.Id)
		item.AddItemToObject("at", usage.At)
		item.AddItemToObject("ip", usage.Ip)
		item.AddItemToObject("kind", usage.Kind)
		item.AddItemToObject("input_tokens", usage.InputTokens)
		item.AddItemToObject("output_tokens", usage.OutputTokens)
		usages.AddItemToArray(item)
	}
	summaries := serial.CreateArray()
	for _, summary := range found.Summaries {
		summaries.AddItemToArray(AgentSummaryJSON(sandbox, summary, true))
	}
	held := serial.CreateObject()
	for _, table := range privacy.HoldTables {
		ids := serial.CreateArray()
		add := func(id int64) {
			if found.Held[table][id] {
				ids.AddItemToArray(id)
			}
		}
		switch table {
		case privacy.TableContactRequest:
			for _, request := range found.Contacts {
				add(request.Id)
			}
		case privacy.TableAgentSummary:
			for _, summary := range found.Summaries {
				add(summary.Id)
			}
		case privacy.TableAccessLog:
			for _, row := range found.Access {
				add(row.Id)
			}
		}
		held.AddItemToObject(table, ids)
	}
	keptByLaw := serial.CreateArray()
	for _, row := range found.Access {
		if found.InsideTerm[row.Id] {
			keptByLaw.AddItemToArray(row.Id)
		}
	}
	document := serial.CreateObject()
	document.AddItemToObject("contact_requests", contacts)
	document.AddItemToObject("views", views)
	document.AddItemToObject("agent_usage", usages)
	document.AddItemToObject("agent_summaries", summaries)
	if found.AccessShown {
		document.AddItemToObject("access_log", AccessRowsJSON(sandbox, found.Access))
	} else {
		document.AddItemToObject("access_log", serial.CreateNull())
	}
	document.AddItemToObject("held", held)
	document.AddItemToObject("kept_by_law", keptByLaw)
	document.AddItemToObject("ip_key_made_at", found.IpKeyMadeAt)
	document.AddItemToObject("total", found.Total())
	return document
}

// ErasedJSON is {erased: {contact_requests, views, agent_usage,
// agent_summaries, access_log, total}, kept: {held, access_log}}.
func ErasedJSON(sandbox *api.Sandbox, erased privacyadmin.Erased) *serializabledeps.SerializableObject {
	counts := sandbox.Deps.SerializableDeps.CreateObject()
	counts.AddItemToObject("contact_requests", erased.Contacts)
	counts.AddItemToObject("views", erased.Views)
	counts.AddItemToObject("agent_usage", erased.Usages)
	counts.AddItemToObject("agent_summaries", erased.Summaries)
	counts.AddItemToObject("access_log", erased.Access)
	counts.AddItemToObject("total", erased.Total())
	kept := sandbox.Deps.SerializableDeps.CreateObject()
	kept.AddItemToObject("held", erased.Held)
	kept.AddItemToObject("access_log", erased.KeptByLaw)
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("erased", counts)
	document.AddItemToObject("kept", kept)
	return document
}

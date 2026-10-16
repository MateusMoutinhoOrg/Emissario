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

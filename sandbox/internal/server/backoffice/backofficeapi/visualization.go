package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The documents of the visualization api, each the JSON twin of a page of
// /admin/visualization. Every instant is Unix seconds; every ip is the one
// the listing handed over, already masked to a viewer.

// VisualizationMetricsJSON is {summary, timeline, top}.
func VisualizationMetricsJSON(sandbox *api.Sandbox, metrics visualization.Metrics) *serializabledeps.SerializableObject {
	serializables := sandbox.Deps.SerializableDeps
	summary := serializables.CreateObject()
	summary.AddItemToObject("views", metrics.Summary.Views)
	summary.AddItemToObject("unique_visitors", metrics.Summary.Unique)
	summary.AddItemToObject("views_per_visitor", metrics.Summary.ViewsPerVisitor)
	summary.AddItemToObject("not_found", metrics.Summary.NotFound)
	summary.AddItemToObject("bots", metrics.Summary.Bots)

	timeline := serializables.CreateArray()
	for _, bucket := range metrics.Timeline {
		item := serializables.CreateObject()
		item.AddItemToObject("start", bucket.Start)
		item.AddItemToObject("views", bucket.Views)
		item.AddItemToObject("unique", bucket.Unique)
		timeline.AddItemToArray(item)
	}

	top := serializables.CreateObject()
	top.AddItemToObject("pages", topRowsJSON(sandbox, metrics.Top.Pages))
	top.AddItemToObject("countries", topRowsJSON(sandbox, metrics.Top.Countries))
	top.AddItemToObject("devices", topRowsJSON(sandbox, metrics.Top.Devices))
	top.AddItemToObject("browsers", topRowsJSON(sandbox, metrics.Top.Browsers))
	top.AddItemToObject("os", topRowsJSON(sandbox, metrics.Top.Os))
	top.AddItemToObject("referrers", topRowsJSON(sandbox, metrics.Top.Referrers))
	top.AddItemToObject("not_found", topRowsJSON(sandbox, metrics.Top.NotFound))
	top.AddItemToObject("visitors", topRowsJSON(sandbox, metrics.Top.Visitors))

	document := serializables.CreateObject()
	document.AddItemToObject("summary", summary)
	document.AddItemToObject("bucket", metrics.Bucket)
	document.AddItemToObject("timeline", timeline)
	document.AddItemToObject("top", top)
	return document
}

// topRowsJSON is rows as [{value, views, unique, percent}].
func topRowsJSON(sandbox *api.Sandbox, rows []visualization.TopRow) *serializabledeps.SerializableObject {
	array := sandbox.Deps.SerializableDeps.CreateArray()
	for _, row := range rows {
		item := sandbox.Deps.SerializableDeps.CreateObject()
		item.AddItemToObject("value", row.Value)
		item.AddItemToObject("views", row.Views)
		item.AddItemToObject("unique", row.Unique)
		item.AddItemToObject("percent", row.Percent)
		array.AddItemToArray(item)
	}
	return array
}

// pagingJSON adds total, page, pages and limit to document.
func pagingJSON(sandbox *api.Sandbox, document *serializabledeps.SerializableObject, paging visualization.Paging) {
	document.AddItemToObject("total", paging.Total)
	document.AddItemToObject("page", paging.Page)
	document.AddItemToObject("pages", paging.Pages)
	document.AddItemToObject("limit", paging.Limit)
}

// VisualizationPagesJSON is {website_pages, total, page, pages, limit}.
func VisualizationPagesJSON(sandbox *api.Sandbox, rows []visualization.PageRow, paging visualization.Paging) *serializabledeps.SerializableObject {
	pages := sandbox.Deps.SerializableDeps.CreateArray()
	for _, row := range rows {
		item := sandbox.Deps.SerializableDeps.CreateObject()
		item.AddItemToObject("page", row.Page)
		item.AddItemToObject("views", row.Views)
		item.AddItemToObject("unique", row.Unique)
		item.AddItemToObject("not_found", row.NotFound)
		item.AddItemToObject("last_at", row.LastAt)
		pages.AddItemToArray(item)
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("website_pages", pages)
	pagingJSON(sandbox, document, paging)
	return document
}

// VisualizationVisitorsJSON is {visitors, total, page, pages, limit}.
func VisualizationVisitorsJSON(sandbox *api.Sandbox, rows []visualization.VisitorRow, paging visualization.Paging) *serializabledeps.SerializableObject {
	visitors := sandbox.Deps.SerializableDeps.CreateArray()
	for _, row := range rows {
		item := sandbox.Deps.SerializableDeps.CreateObject()
		item.AddItemToObject("ip", row.Ip)
		item.AddItemToObject("country", row.Country)
		item.AddItemToObject("device", row.Device)
		item.AddItemToObject("browser", row.Browser)
		item.AddItemToObject("os", row.Os)
		item.AddItemToObject("views", row.Views)
		item.AddItemToObject("pages", row.Pages)
		item.AddItemToObject("first_at", row.FirstAt)
		item.AddItemToObject("last_at", row.LastAt)
		visitors.AddItemToArray(item)
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("visitors", visitors)
	pagingJSON(sandbox, document, paging)
	return document
}

// VisualizationViewJSON is one view: {id, at, path, page, status, ip,
// country, device, browser, os, referrer, user_agent}.
func VisualizationViewJSON(sandbox *api.Sandbox, view visualization_db.VisualizationViewRecord) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", view.Id)
	item.AddItemToObject("at", view.At)
	item.AddItemToObject("path", view.Path)
	item.AddItemToObject("page", view.Page)
	item.AddItemToObject("status", view.Status)
	item.AddItemToObject("ip", view.Ip)
	item.AddItemToObject("country", view.Country)
	item.AddItemToObject("device", view.Device)
	item.AddItemToObject("browser", view.Browser)
	item.AddItemToObject("os", view.Os)
	item.AddItemToObject("referrer", view.Referrer)
	item.AddItemToObject("user_agent", view.UserAgent)
	return item
}

// VisualizationViewsJSON is {views, total, page, pages, limit}.
func VisualizationViewsJSON(sandbox *api.Sandbox, views []visualization_db.VisualizationViewRecord, paging visualization.Paging) *serializabledeps.SerializableObject {
	array := sandbox.Deps.SerializableDeps.CreateArray()
	for _, view := range views {
		array.AddItemToArray(VisualizationViewJSON(sandbox, view))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("views", array)
	pagingJSON(sandbox, document, paging)
	return document
}

// VisualizationSettingsJSON is {enabled, retention_days, ignored_ips: [..]}.
func VisualizationSettingsJSON(sandbox *api.Sandbox, settings visualization.Settings) *serializabledeps.SerializableObject {
	ignored := sandbox.Deps.SerializableDeps.CreateArray()
	for _, entry := range settings.IgnoredIps {
		ignored.AddItemToArray(entry)
	}
	object := sandbox.Deps.SerializableDeps.CreateObject()
	object.AddItemToObject("enabled", settings.Enabled)
	object.AddItemToObject("retention_days", settings.RetentionDays)
	object.AddItemToObject("ignored_ips", ignored)
	return object
}

// VisualizationSettingsResponseJSON is {settings}.
func VisualizationSettingsResponseJSON(sandbox *api.Sandbox, settings visualization.Settings) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", VisualizationSettingsJSON(sandbox, settings))
	return document
}

// VisualizationSettingsStateJSON is {settings, geoip, stored, oldest_at,
// dropped}: geoip is {source, built_at}, source "embedded" or the
// --visualization-geoip file.
func VisualizationSettingsStateJSON(sandbox *api.Sandbox, settings visualization.Settings, geoip visualization.GeoipInfo, stored int, oldestAt int64, dropped int64) *serializabledeps.SerializableObject {
	database := sandbox.Deps.SerializableDeps.CreateObject()
	database.AddItemToObject("source", geoip.Source)
	database.AddItemToObject("built_at", geoip.BuiltAt)

	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("settings", VisualizationSettingsJSON(sandbox, settings))
	document.AddItemToObject("geoip", database)
	document.AddItemToObject("stored", stored)
	document.AddItemToObject("oldest_at", oldestAt)
	document.AddItemToObject("dropped", dropped)
	return document
}

// VisualizationPurgeJSON is {removed}.
func VisualizationPurgeJSON(sandbox *api.Sandbox, removed int) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("removed", removed)
	return document
}

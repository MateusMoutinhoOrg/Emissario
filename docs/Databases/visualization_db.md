# `VisualizationDb`

`sandbox/internal/databases/visualization_db/`, keys under `visualizationdb`. Build one with
`visualization_db.New(sandbox)` — it touches no key, so building one is free.

## `visualization-view`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `at` | `integer` | yes |  |
| `path` | `string` | yes |  |
| `page` | `string` | yes |  |
| `status` | `integer` | yes |  |
| `ip` | `string` | yes |  |
| `country` | `string` |  |  |
| `device` | `string` |  |  |
| `browser` | `string` |  |  |
| `os` | `string` |  |  |
| `user-agent` | `string` |  |  |
| `referrer` | `string` |  |  |

## `visualization-setting`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `enabled` | `integer` |  |  |
| `retention-days` | `integer` |  |  |
| `ignored-ips` | `string` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddVisualizationView(props VisualizationViewInput) (VisualizationViewRecord, error)` | inserts one visualization-view record |
| `FindVisualizationViewById(id int64) (VisualizationViewRecord, bool)` | reads one visualization-view record by its permanent id |
| `ListVisualizationViews(filter VisualizationViewFilter) ([]VisualizationViewRecord, error)` | reads every visualization-view record the filter keeps |
| `ListVisualizationViewsPage(offset int, limit int) ([]VisualizationViewRecord, error)` | reads up to limit visualization-view records after the first offset, every one past them when limit is 0 |
| `CountVisualizationView() (int, error)` | is how many visualization-view records are live |
| `SetVisualizationViewAt(id int64, value int64) error` | writes a new at on one visualization-view record |
| `SetVisualizationViewPath(id int64, value string) error` | writes a new path on one visualization-view record |
| `SetVisualizationViewPage(id int64, value string) error` | writes a new page on one visualization-view record |
| `SetVisualizationViewStatus(id int64, value int64) error` | writes a new status on one visualization-view record |
| `SetVisualizationViewIp(id int64, value string) error` | writes a new ip on one visualization-view record |
| `SetVisualizationViewCountry(id int64, value string) error` | writes a new country on one visualization-view record |
| `SetVisualizationViewDevice(id int64, value string) error` | writes a new device on one visualization-view record |
| `SetVisualizationViewBrowser(id int64, value string) error` | writes a new browser on one visualization-view record |
| `SetVisualizationViewOs(id int64, value string) error` | writes a new os on one visualization-view record |
| `SetVisualizationViewUserAgent(id int64, value string) error` | writes a new user-agent on one visualization-view record |
| `SetVisualizationViewReferrer(id int64, value string) error` | writes a new referrer on one visualization-view record |
| `RemoveVisualizationView(id int64) error` | deletes one visualization-view record and everything nested under it |
| `AddVisualizationSetting(props VisualizationSettingInput) (VisualizationSettingRecord, error)` | inserts one visualization-setting record |
| `FindVisualizationSettingById(id int64) (VisualizationSettingRecord, bool)` | reads one visualization-setting record by its permanent id |
| `ListVisualizationSettings(filter VisualizationSettingFilter) ([]VisualizationSettingRecord, error)` | reads every visualization-setting record the filter keeps |
| `ListVisualizationSettingsPage(offset int, limit int) ([]VisualizationSettingRecord, error)` | reads up to limit visualization-setting records after the first offset, every one past them when limit is 0 |
| `CountVisualizationSetting() (int, error)` | is how many visualization-setting records are live |
| `SetVisualizationSettingEnabled(id int64, value int64) error` | writes a new enabled on one visualization-setting record |
| `SetVisualizationSettingRetentionDays(id int64, value int64) error` | writes a new retention-days on one visualization-setting record |
| `SetVisualizationSettingIgnoredIps(id int64, value string) error` | writes a new ignored-ips on one visualization-setting record |
| `RemoveVisualizationSetting(id int64) error` | deletes one visualization-setting record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

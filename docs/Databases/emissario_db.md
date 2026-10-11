# `EmissarioDb`

`sandbox/internal/databases/emissario_db/`, keys under `emissariodb`. Build one with
`emissario_db.New(sandbox)` — it touches no key, so building one is free.

## `source`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `string` | yes |  |
| `repourl` | `string` | yes |  |
| `branch` | `string` |  |  |
| `sync-time` | `string` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddSource(props SourceInput) (SourceRecord, error)` | inserts one source record |
| `FindSourceById(id int64) (SourceRecord, bool)` | reads one source record by its permanent id |
| `ListSources(filter SourceFilter) ([]SourceRecord, error)` | reads every source record the filter keeps |
| `ListSourcesPage(offset int, limit int) ([]SourceRecord, error)` | reads up to limit source records after the first offset, every one past them when limit is 0 |
| `CountSource() (int, error)` | is how many source records are live |
| `SetSourceName(id int64, value string) error` | writes a new name on one source record |
| `SetSourceRepourl(id int64, value string) error` | writes a new repourl on one source record |
| `SetSourceBranch(id int64, value string) error` | writes a new branch on one source record |
| `SetSourceSyncTime(id int64, value string) error` | writes a new sync-time on one source record |
| `RemoveSource(id int64) error` | deletes one source record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

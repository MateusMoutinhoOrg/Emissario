# `BlockerDb`

`sandbox/internal/databases/blocker_db/`, keys under `blockerdb`. Build one with
`blocker_db.New(sandbox)` — it touches no key, so building one is free.

## `ip-block`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `value` | `key` | yes |  |
| `kind` | `string` | yes |  |
| `reason` | `string` |  |  |
| `created-at` | `integer` |  |  |
| `expires-at` | `integer` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddIpBlock(props IpBlockInput) (IpBlockRecord, error)` | inserts one ip-block record |
| `FindIpBlockById(id int64) (IpBlockRecord, bool)` | reads one ip-block record by its permanent id |
| `FindIpBlockByValue(value string) (IpBlockRecord, bool)` | reads one ip-block record by its indexed value |
| `ListIpBlocks(filter IpBlockFilter) ([]IpBlockRecord, error)` | reads every ip-block record the filter keeps |
| `ListIpBlocksPage(offset int, limit int) ([]IpBlockRecord, error)` | reads up to limit ip-block records after the first offset, every one past them when limit is 0 |
| `CountIpBlock() (int, error)` | is how many ip-block records are live |
| `SetIpBlockValue(id int64, value string) error` | writes a new value on one ip-block record |
| `SetIpBlockKind(id int64, value string) error` | writes a new kind on one ip-block record |
| `SetIpBlockReason(id int64, value string) error` | writes a new reason on one ip-block record |
| `SetIpBlockCreatedAt(id int64, value int64) error` | writes a new created-at on one ip-block record |
| `SetIpBlockExpiresAt(id int64, value int64) error` | writes a new expires-at on one ip-block record |
| `RemoveIpBlock(id int64) error` | deletes one ip-block record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

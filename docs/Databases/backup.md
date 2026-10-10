# `Backup`

`sandbox/internal/databases/backup/`, keys under `backup`. Build one with
`backup.New(sandbox)` — it touches no key, so building one is free.

## `blob`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `sha` | `key` | yes |  |
| `value` | `bytes` | yes |  |

## `snapshot`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `key` | yes |  |
| `data` | `integer` | yes |  |
| `content` | `object` |  |  |
| `content.path` | `string` | yes |  |
| `content.sha` | `string` | yes |  |
| `status` | `string` | yes |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddBlob(props BlobInput) (BlobRecord, error)` | inserts one blob record |
| `FindBlobById(id int64) (BlobRecord, bool)` | reads one blob record by its permanent id |
| `FindBlobBySha(value string) (BlobRecord, bool)` | reads one blob record by its indexed sha |
| `ListBlobs(filter BlobFilter) ([]BlobRecord, error)` | reads every blob record the filter keeps |
| `ListBlobsPage(offset int, limit int) ([]BlobRecord, error)` | reads up to limit blob records after the first offset, every one past them when limit is 0 |
| `CountBlob() (int, error)` | is how many blob records are live |
| `SetBlobSha(id int64, value string) error` | writes a new sha on one blob record |
| `SetBlobValue(id int64, value []byte) error` | writes a new value on one blob record |
| `RemoveBlob(id int64) error` | deletes one blob record and everything nested under it |
| `AddSnapshot(props SnapshotInput) (SnapshotRecord, error)` | inserts one snapshot record |
| `FindSnapshotById(id int64) (SnapshotRecord, bool)` | reads one snapshot record by its permanent id |
| `FindSnapshotByName(value string) (SnapshotRecord, bool)` | reads one snapshot record by its indexed name |
| `ListSnapshots(filter SnapshotFilter) ([]SnapshotRecord, error)` | reads every snapshot record the filter keeps |
| `ListSnapshotsPage(offset int, limit int) ([]SnapshotRecord, error)` | reads up to limit snapshot records after the first offset, every one past them when limit is 0 |
| `CountSnapshot() (int, error)` | is how many snapshot records are live |
| `SetSnapshotName(id int64, value string) error` | writes a new name on one snapshot record |
| `SetSnapshotData(id int64, value int64) error` | writes a new data on one snapshot record |
| `SetSnapshotStatus(id int64, value string) error` | writes a new status on one snapshot record |
| `RemoveSnapshot(id int64) error` | deletes one snapshot record and everything nested under it |
| `AddSnapshotContent(parentId int64, props SnapshotContentInput) (SnapshotContentRecord, error)` | inserts one content record under one snapshot record |
| `ListSnapshotContents(parentId int64) ([]SnapshotContentRecord, error)` | reads every content record of one snapshot record |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

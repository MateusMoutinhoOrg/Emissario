# `PrivacyDb`

`sandbox/internal/databases/privacy_db/`, keys under `privacydb`. Build one with
`privacy_db.New(sandbox)` — it touches no key, so building one is free.

## `privacy-setting`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `ip-mode` | `string` |  |  |
| `viewer-ip` | `string` |  |  |
| `access-log` | `integer` |  |  |
| `access-log-retention-days` | `integer` |  |  |
| `access-log-totp` | `integer` |  |  |
| `access-log-inventory` | `integer` |  |  |
| `ip-key` | `bytes` |  |  |
| `ip-key-rotated-at` | `integer` |  |  |
| `ip-retention-days` | `integer` |  |  |
| `backup-retention-days` | `integer` |  |  |
| `notice-draft` | `string` |  |  |
| `preset` | `string` |  |  |

## `privacy-preset`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `preset-id` | `string` |  |  |
| `version` | `integer` |  |  |
| `name` | `string` |  |  |
| `json` | `string` |  |  |
| `created-at` | `integer` |  |  |

## `access-log`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `at` | `integer` | yes |  |
| `ip` | `string` |  |  |
| `port` | `integer` |  |  |
| `kind` | `string` |  |  |
| `ref` | `integer` |  |  |

## `privacy-access`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `at` | `integer` | yes |  |
| `ended-at` | `integer` |  |  |
| `user-id` | `integer` |  |  |
| `username` | `string` |  |  |
| `action` | `string` |  |  |
| `filter` | `string` |  |  |
| `rows` | `integer` |  |  |

## `privacy-hold`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `reason` | `string` |  |  |
| `until` | `integer` |  |  |
| `table` | `string` |  |  |
| `ids` | `string` |  |  |

## `privacy-erasure`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `at` | `integer` | yes |  |
| `table` | `string` |  |  |
| `ids` | `string` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddPrivacySetting(props PrivacySettingInput) (PrivacySettingRecord, error)` | inserts one privacy-setting record |
| `FindPrivacySettingById(id int64) (PrivacySettingRecord, bool)` | reads one privacy-setting record by its permanent id |
| `ListPrivacySettings(filter PrivacySettingFilter) ([]PrivacySettingRecord, error)` | reads every privacy-setting record the filter keeps |
| `ListPrivacySettingsPage(offset int, limit int) ([]PrivacySettingRecord, error)` | reads up to limit privacy-setting records after the first offset, every one past them when limit is 0 |
| `CountPrivacySetting() (int, error)` | is how many privacy-setting records are live |
| `SetPrivacySettingIpMode(id int64, value string) error` | writes a new ip-mode on one privacy-setting record |
| `SetPrivacySettingViewerIp(id int64, value string) error` | writes a new viewer-ip on one privacy-setting record |
| `SetPrivacySettingAccessLog(id int64, value int64) error` | writes a new access-log on one privacy-setting record |
| `SetPrivacySettingAccessLogRetentionDays(id int64, value int64) error` | writes a new access-log-retention-days on one privacy-setting record |
| `SetPrivacySettingAccessLogTotp(id int64, value int64) error` | writes a new access-log-totp on one privacy-setting record |
| `SetPrivacySettingAccessLogInventory(id int64, value int64) error` | writes a new access-log-inventory on one privacy-setting record |
| `SetPrivacySettingIpKey(id int64, value []byte) error` | writes a new ip-key on one privacy-setting record |
| `SetPrivacySettingIpKeyRotatedAt(id int64, value int64) error` | writes a new ip-key-rotated-at on one privacy-setting record |
| `SetPrivacySettingIpRetentionDays(id int64, value int64) error` | writes a new ip-retention-days on one privacy-setting record |
| `SetPrivacySettingBackupRetentionDays(id int64, value int64) error` | writes a new backup-retention-days on one privacy-setting record |
| `SetPrivacySettingNoticeDraft(id int64, value string) error` | writes a new notice-draft on one privacy-setting record |
| `SetPrivacySettingPreset(id int64, value string) error` | writes a new preset on one privacy-setting record |
| `RemovePrivacySetting(id int64) error` | deletes one privacy-setting record and everything nested under it |
| `AddPrivacyPreset(props PrivacyPresetInput) (PrivacyPresetRecord, error)` | inserts one privacy-preset record |
| `FindPrivacyPresetById(id int64) (PrivacyPresetRecord, bool)` | reads one privacy-preset record by its permanent id |
| `ListPrivacyPresets(filter PrivacyPresetFilter) ([]PrivacyPresetRecord, error)` | reads every privacy-preset record the filter keeps |
| `ListPrivacyPresetsPage(offset int, limit int) ([]PrivacyPresetRecord, error)` | reads up to limit privacy-preset records after the first offset, every one past them when limit is 0 |
| `CountPrivacyPreset() (int, error)` | is how many privacy-preset records are live |
| `SetPrivacyPresetPresetId(id int64, value string) error` | writes a new preset-id on one privacy-preset record |
| `SetPrivacyPresetVersion(id int64, value int64) error` | writes a new version on one privacy-preset record |
| `SetPrivacyPresetName(id int64, value string) error` | writes a new name on one privacy-preset record |
| `SetPrivacyPresetJson(id int64, value string) error` | writes a new json on one privacy-preset record |
| `SetPrivacyPresetCreatedAt(id int64, value int64) error` | writes a new created-at on one privacy-preset record |
| `RemovePrivacyPreset(id int64) error` | deletes one privacy-preset record and everything nested under it |
| `AddAccessLog(props AccessLogInput) (AccessLogRecord, error)` | inserts one access-log record |
| `FindAccessLogById(id int64) (AccessLogRecord, bool)` | reads one access-log record by its permanent id |
| `ListAccessLogs(filter AccessLogFilter) ([]AccessLogRecord, error)` | reads every access-log record the filter keeps |
| `ListAccessLogsPage(offset int, limit int) ([]AccessLogRecord, error)` | reads up to limit access-log records after the first offset, every one past them when limit is 0 |
| `CountAccessLog() (int, error)` | is how many access-log records are live |
| `SetAccessLogAt(id int64, value int64) error` | writes a new at on one access-log record |
| `SetAccessLogIp(id int64, value string) error` | writes a new ip on one access-log record |
| `SetAccessLogPort(id int64, value int64) error` | writes a new port on one access-log record |
| `SetAccessLogKind(id int64, value string) error` | writes a new kind on one access-log record |
| `SetAccessLogRef(id int64, value int64) error` | writes a new ref on one access-log record |
| `RemoveAccessLog(id int64) error` | deletes one access-log record and everything nested under it |
| `AddPrivacyAccess(props PrivacyAccessInput) (PrivacyAccessRecord, error)` | inserts one privacy-access record |
| `FindPrivacyAccessById(id int64) (PrivacyAccessRecord, bool)` | reads one privacy-access record by its permanent id |
| `ListPrivacyAccesses(filter PrivacyAccessFilter) ([]PrivacyAccessRecord, error)` | reads every privacy-access record the filter keeps |
| `ListPrivacyAccessesPage(offset int, limit int) ([]PrivacyAccessRecord, error)` | reads up to limit privacy-access records after the first offset, every one past them when limit is 0 |
| `CountPrivacyAccess() (int, error)` | is how many privacy-access records are live |
| `SetPrivacyAccessAt(id int64, value int64) error` | writes a new at on one privacy-access record |
| `SetPrivacyAccessEndedAt(id int64, value int64) error` | writes a new ended-at on one privacy-access record |
| `SetPrivacyAccessUserId(id int64, value int64) error` | writes a new user-id on one privacy-access record |
| `SetPrivacyAccessUsername(id int64, value string) error` | writes a new username on one privacy-access record |
| `SetPrivacyAccessAction(id int64, value string) error` | writes a new action on one privacy-access record |
| `SetPrivacyAccessFilter(id int64, value string) error` | writes a new filter on one privacy-access record |
| `SetPrivacyAccessRows(id int64, value int64) error` | writes a new rows on one privacy-access record |
| `RemovePrivacyAccess(id int64) error` | deletes one privacy-access record and everything nested under it |
| `AddPrivacyHold(props PrivacyHoldInput) (PrivacyHoldRecord, error)` | inserts one privacy-hold record |
| `FindPrivacyHoldById(id int64) (PrivacyHoldRecord, bool)` | reads one privacy-hold record by its permanent id |
| `ListPrivacyHolds(filter PrivacyHoldFilter) ([]PrivacyHoldRecord, error)` | reads every privacy-hold record the filter keeps |
| `ListPrivacyHoldsPage(offset int, limit int) ([]PrivacyHoldRecord, error)` | reads up to limit privacy-hold records after the first offset, every one past them when limit is 0 |
| `CountPrivacyHold() (int, error)` | is how many privacy-hold records are live |
| `SetPrivacyHoldReason(id int64, value string) error` | writes a new reason on one privacy-hold record |
| `SetPrivacyHoldUntil(id int64, value int64) error` | writes a new until on one privacy-hold record |
| `SetPrivacyHoldTable(id int64, value string) error` | writes a new table on one privacy-hold record |
| `SetPrivacyHoldIds(id int64, value string) error` | writes a new ids on one privacy-hold record |
| `RemovePrivacyHold(id int64) error` | deletes one privacy-hold record and everything nested under it |
| `AddPrivacyErasure(props PrivacyErasureInput) (PrivacyErasureRecord, error)` | inserts one privacy-erasure record |
| `FindPrivacyErasureById(id int64) (PrivacyErasureRecord, bool)` | reads one privacy-erasure record by its permanent id |
| `ListPrivacyErasures(filter PrivacyErasureFilter) ([]PrivacyErasureRecord, error)` | reads every privacy-erasure record the filter keeps |
| `ListPrivacyErasuresPage(offset int, limit int) ([]PrivacyErasureRecord, error)` | reads up to limit privacy-erasure records after the first offset, every one past them when limit is 0 |
| `CountPrivacyErasure() (int, error)` | is how many privacy-erasure records are live |
| `SetPrivacyErasureAt(id int64, value int64) error` | writes a new at on one privacy-erasure record |
| `SetPrivacyErasureTable(id int64, value string) error` | writes a new table on one privacy-erasure record |
| `SetPrivacyErasureIds(id int64, value string) error` | writes a new ids on one privacy-erasure record |
| `RemovePrivacyErasure(id int64) error` | deletes one privacy-erasure record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

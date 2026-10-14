# `ContactDb`

`sandbox/internal/databases/contact_db/`, keys under `contactdb`. Build one with
`contact_db.New(sandbox)` — it touches no key, so building one is free.

## `contact-request`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `ref` | `key` | yes |  |
| `at` | `integer` | yes |  |
| `ip` | `string` | yes |  |
| `country` | `string` |  |  |
| `body` | `string` |  |  |
| `size` | `integer` |  |  |
| `archived` | `integer` |  |  |

## `contact-setting`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `enabled` | `integer` |  |  |
| `max-bytes` | `integer` |  |  |
| `rate-per-minute` | `integer` |  |  |
| `ip-daily-limit` | `integer` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddContactRequest(props ContactRequestInput) (ContactRequestRecord, error)` | inserts one contact-request record |
| `FindContactRequestById(id int64) (ContactRequestRecord, bool)` | reads one contact-request record by its permanent id |
| `FindContactRequestByRef(value string) (ContactRequestRecord, bool)` | reads one contact-request record by its indexed ref |
| `ListContactRequests(filter ContactRequestFilter) ([]ContactRequestRecord, error)` | reads every contact-request record the filter keeps |
| `ListContactRequestsPage(offset int, limit int) ([]ContactRequestRecord, error)` | reads up to limit contact-request records after the first offset, every one past them when limit is 0 |
| `CountContactRequest() (int, error)` | is how many contact-request records are live |
| `SetContactRequestRef(id int64, value string) error` | writes a new ref on one contact-request record |
| `SetContactRequestAt(id int64, value int64) error` | writes a new at on one contact-request record |
| `SetContactRequestIp(id int64, value string) error` | writes a new ip on one contact-request record |
| `SetContactRequestCountry(id int64, value string) error` | writes a new country on one contact-request record |
| `SetContactRequestBody(id int64, value string) error` | writes a new body on one contact-request record |
| `SetContactRequestSize(id int64, value int64) error` | writes a new size on one contact-request record |
| `SetContactRequestArchived(id int64, value int64) error` | writes a new archived on one contact-request record |
| `RemoveContactRequest(id int64) error` | deletes one contact-request record and everything nested under it |
| `AddContactSetting(props ContactSettingInput) (ContactSettingRecord, error)` | inserts one contact-setting record |
| `FindContactSettingById(id int64) (ContactSettingRecord, bool)` | reads one contact-setting record by its permanent id |
| `ListContactSettings(filter ContactSettingFilter) ([]ContactSettingRecord, error)` | reads every contact-setting record the filter keeps |
| `ListContactSettingsPage(offset int, limit int) ([]ContactSettingRecord, error)` | reads up to limit contact-setting records after the first offset, every one past them when limit is 0 |
| `CountContactSetting() (int, error)` | is how many contact-setting records are live |
| `SetContactSettingEnabled(id int64, value int64) error` | writes a new enabled on one contact-setting record |
| `SetContactSettingMaxBytes(id int64, value int64) error` | writes a new max-bytes on one contact-setting record |
| `SetContactSettingRatePerMinute(id int64, value int64) error` | writes a new rate-per-minute on one contact-setting record |
| `SetContactSettingIpDailyLimit(id int64, value int64) error` | writes a new ip-daily-limit on one contact-setting record |
| `RemoveContactSetting(id int64) error` | deletes one contact-setting record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

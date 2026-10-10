# `BackofficeDb`

`sandbox/internal/databases/backoffice_db/`, keys under `backofficedb`. Build one with
`backoffice_db.New(sandbox)` — it touches no key, so building one is free.

## `backoffice-user`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `username` | `string` | yes |  |
| `email` | `string` | yes |  |
| `password-hash` | `string` | yes |  |
| `role` | `integer` |  |  |
| `session` | `object` |  |  |
| `session.expires-at` | `integer` | yes |  |

## `api-token`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `token-sha256` | `key` | yes |  |
| `name` | `string` | yes |  |
| `prefix` | `string` | yes |  |
| `owner-id` | `integer` |  |  |
| `created-at` | `integer` |  |  |
| `expires-at` | `integer` |  |  |
| `ips` | `string` |  |  |
| `last-used-at` | `integer` |  |  |
| `last-used-ip` | `string` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddBackofficeUser(props BackofficeUserInput) (BackofficeUserRecord, error)` | inserts one backoffice-user record |
| `FindBackofficeUserById(id int64) (BackofficeUserRecord, bool)` | reads one backoffice-user record by its permanent id |
| `ListBackofficeUsers(filter BackofficeUserFilter) ([]BackofficeUserRecord, error)` | reads every backoffice-user record the filter keeps |
| `ListBackofficeUsersPage(offset int, limit int) ([]BackofficeUserRecord, error)` | reads up to limit backoffice-user records after the first offset, every one past them when limit is 0 |
| `CountBackofficeUser() (int, error)` | is how many backoffice-user records are live |
| `SetBackofficeUserUsername(id int64, value string) error` | writes a new username on one backoffice-user record |
| `SetBackofficeUserEmail(id int64, value string) error` | writes a new email on one backoffice-user record |
| `SetBackofficeUserPasswordHash(id int64, value string) error` | writes a new password-hash on one backoffice-user record |
| `SetBackofficeUserRole(id int64, value int64) error` | writes a new role on one backoffice-user record |
| `RemoveBackofficeUser(id int64) error` | deletes one backoffice-user record and everything nested under it |
| `AddBackofficeUserSession(parentId int64, props BackofficeUserSessionInput) (BackofficeUserSessionRecord, error)` | inserts one session record under one backoffice-user record |
| `ListBackofficeUserSessions(parentId int64) ([]BackofficeUserSessionRecord, error)` | reads every session record of one backoffice-user record |
| `AddApiToken(props ApiTokenInput) (ApiTokenRecord, error)` | inserts one api-token record |
| `FindApiTokenById(id int64) (ApiTokenRecord, bool)` | reads one api-token record by its permanent id |
| `FindApiTokenByTokenSha256(value string) (ApiTokenRecord, bool)` | reads one api-token record by its indexed token-sha256 |
| `ListApiTokens(filter ApiTokenFilter) ([]ApiTokenRecord, error)` | reads every api-token record the filter keeps |
| `ListApiTokensPage(offset int, limit int) ([]ApiTokenRecord, error)` | reads up to limit api-token records after the first offset, every one past them when limit is 0 |
| `CountApiToken() (int, error)` | is how many api-token records are live |
| `SetApiTokenTokenSha256(id int64, value string) error` | writes a new token-sha256 on one api-token record |
| `SetApiTokenName(id int64, value string) error` | writes a new name on one api-token record |
| `SetApiTokenPrefix(id int64, value string) error` | writes a new prefix on one api-token record |
| `SetApiTokenOwnerId(id int64, value int64) error` | writes a new owner-id on one api-token record |
| `SetApiTokenCreatedAt(id int64, value int64) error` | writes a new created-at on one api-token record |
| `SetApiTokenExpiresAt(id int64, value int64) error` | writes a new expires-at on one api-token record |
| `SetApiTokenIps(id int64, value string) error` | writes a new ips on one api-token record |
| `SetApiTokenLastUsedAt(id int64, value int64) error` | writes a new last-used-at on one api-token record |
| `SetApiTokenLastUsedIp(id int64, value string) error` | writes a new last-used-ip on one api-token record |
| `RemoveApiToken(id int64) error` | deletes one api-token record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

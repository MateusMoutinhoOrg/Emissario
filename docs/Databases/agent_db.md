# `AgentDb`

`sandbox/internal/databases/agent_db/`, keys under `agentdb`. Build one with
`agent_db.New(sandbox)` — it touches no key, so building one is free.

## `agent-setting`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `enabled` | `integer` |  |  |
| `base-url` | `string` |  |  |
| `api-key` | `string` |  |  |
| `model` | `string` |  |  |
| `max-tokens` | `integer` |  |  |
| `daily-limit` | `integer` |  |  |
| `ip-daily-limit` | `integer` |  |  |
| `rate-per-minute` | `integer` |  |  |
| `summaries-enabled` | `integer` |  |  |
| `summary-retention-days` | `integer` |  |  |

## `agent-url`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `key` | yes |  |
| `description` | `string` |  |  |
| `method` | `string` |  |  |
| `url` | `string` |  |  |
| `headers` | `string` |  |  |

## `agent-usage`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `at` | `integer` | yes |  |
| `ip` | `string` |  |  |
| `kind` | `string` | yes |  |
| `input-tokens` | `integer` |  |  |
| `output-tokens` | `integer` |  |  |

## `agent-summary`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `ref` | `key` | yes |  |
| `at` | `integer` | yes |  |
| `summary` | `string` |  |  |
| `messages` | `integer` |  |  |
| `input-tokens` | `integer` |  |  |
| `output-tokens` | `integer` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddAgentSetting(props AgentSettingInput) (AgentSettingRecord, error)` | inserts one agent-setting record |
| `FindAgentSettingById(id int64) (AgentSettingRecord, bool)` | reads one agent-setting record by its permanent id |
| `ListAgentSettings(filter AgentSettingFilter) ([]AgentSettingRecord, error)` | reads every agent-setting record the filter keeps |
| `ListAgentSettingsPage(offset int, limit int) ([]AgentSettingRecord, error)` | reads up to limit agent-setting records after the first offset, every one past them when limit is 0 |
| `CountAgentSetting() (int, error)` | is how many agent-setting records are live |
| `SetAgentSettingEnabled(id int64, value int64) error` | writes a new enabled on one agent-setting record |
| `SetAgentSettingBaseUrl(id int64, value string) error` | writes a new base-url on one agent-setting record |
| `SetAgentSettingApiKey(id int64, value string) error` | writes a new api-key on one agent-setting record |
| `SetAgentSettingModel(id int64, value string) error` | writes a new model on one agent-setting record |
| `SetAgentSettingMaxTokens(id int64, value int64) error` | writes a new max-tokens on one agent-setting record |
| `SetAgentSettingDailyLimit(id int64, value int64) error` | writes a new daily-limit on one agent-setting record |
| `SetAgentSettingIpDailyLimit(id int64, value int64) error` | writes a new ip-daily-limit on one agent-setting record |
| `SetAgentSettingRatePerMinute(id int64, value int64) error` | writes a new rate-per-minute on one agent-setting record |
| `SetAgentSettingSummariesEnabled(id int64, value int64) error` | writes a new summaries-enabled on one agent-setting record |
| `SetAgentSettingSummaryRetentionDays(id int64, value int64) error` | writes a new summary-retention-days on one agent-setting record |
| `RemoveAgentSetting(id int64) error` | deletes one agent-setting record and everything nested under it |
| `AddAgentUrl(props AgentUrlInput) (AgentUrlRecord, error)` | inserts one agent-url record |
| `FindAgentUrlById(id int64) (AgentUrlRecord, bool)` | reads one agent-url record by its permanent id |
| `FindAgentUrlByName(value string) (AgentUrlRecord, bool)` | reads one agent-url record by its indexed name |
| `ListAgentUrls(filter AgentUrlFilter) ([]AgentUrlRecord, error)` | reads every agent-url record the filter keeps |
| `ListAgentUrlsPage(offset int, limit int) ([]AgentUrlRecord, error)` | reads up to limit agent-url records after the first offset, every one past them when limit is 0 |
| `CountAgentUrl() (int, error)` | is how many agent-url records are live |
| `SetAgentUrlName(id int64, value string) error` | writes a new name on one agent-url record |
| `SetAgentUrlDescription(id int64, value string) error` | writes a new description on one agent-url record |
| `SetAgentUrlMethod(id int64, value string) error` | writes a new method on one agent-url record |
| `SetAgentUrlUrl(id int64, value string) error` | writes a new url on one agent-url record |
| `SetAgentUrlHeaders(id int64, value string) error` | writes a new headers on one agent-url record |
| `RemoveAgentUrl(id int64) error` | deletes one agent-url record and everything nested under it |
| `AddAgentUsage(props AgentUsageInput) (AgentUsageRecord, error)` | inserts one agent-usage record |
| `FindAgentUsageById(id int64) (AgentUsageRecord, bool)` | reads one agent-usage record by its permanent id |
| `ListAgentUsages(filter AgentUsageFilter) ([]AgentUsageRecord, error)` | reads every agent-usage record the filter keeps |
| `ListAgentUsagesPage(offset int, limit int) ([]AgentUsageRecord, error)` | reads up to limit agent-usage records after the first offset, every one past them when limit is 0 |
| `CountAgentUsage() (int, error)` | is how many agent-usage records are live |
| `SetAgentUsageAt(id int64, value int64) error` | writes a new at on one agent-usage record |
| `SetAgentUsageIp(id int64, value string) error` | writes a new ip on one agent-usage record |
| `SetAgentUsageKind(id int64, value string) error` | writes a new kind on one agent-usage record |
| `SetAgentUsageInputTokens(id int64, value int64) error` | writes a new input-tokens on one agent-usage record |
| `SetAgentUsageOutputTokens(id int64, value int64) error` | writes a new output-tokens on one agent-usage record |
| `RemoveAgentUsage(id int64) error` | deletes one agent-usage record and everything nested under it |
| `AddAgentSummary(props AgentSummaryInput) (AgentSummaryRecord, error)` | inserts one agent-summary record |
| `FindAgentSummaryById(id int64) (AgentSummaryRecord, bool)` | reads one agent-summary record by its permanent id |
| `FindAgentSummaryByRef(value string) (AgentSummaryRecord, bool)` | reads one agent-summary record by its indexed ref |
| `ListAgentSummaries(filter AgentSummaryFilter) ([]AgentSummaryRecord, error)` | reads every agent-summary record the filter keeps |
| `ListAgentSummariesPage(offset int, limit int) ([]AgentSummaryRecord, error)` | reads up to limit agent-summary records after the first offset, every one past them when limit is 0 |
| `CountAgentSummary() (int, error)` | is how many agent-summary records are live |
| `SetAgentSummaryRef(id int64, value string) error` | writes a new ref on one agent-summary record |
| `SetAgentSummaryAt(id int64, value int64) error` | writes a new at on one agent-summary record |
| `SetAgentSummarySummary(id int64, value string) error` | writes a new summary on one agent-summary record |
| `SetAgentSummaryMessages(id int64, value int64) error` | writes a new messages on one agent-summary record |
| `SetAgentSummaryInputTokens(id int64, value int64) error` | writes a new input-tokens on one agent-summary record |
| `SetAgentSummaryOutputTokens(id int64, value int64) error` | writes a new output-tokens on one agent-summary record |
| `RemoveAgentSummary(id int64) error` | deletes one agent-summary record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)

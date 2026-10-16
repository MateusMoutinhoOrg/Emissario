package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

const (
	// DefaultMaxTokens is the output tokens one provider call may take when no
	// root set any.
	DefaultMaxTokens = 1024
	// MaxMaxTokens is the most output tokens a root may let one call take.
	MaxMaxTokens = 1000000
	// DefaultRatePerMinute is the requests one ip may send a minute when no
	// root set any.
	DefaultRatePerMinute = 10
	// MaxRatePerMinute is the most requests a minute a root may let one ip
	// send.
	MaxRatePerMinute = 10000
	// MaxTokenLimit is the largest daily limit a root may set.
	MaxTokenLimit = 1000000000000
	// MaxBaseUrlLength is the longest base url accepted.
	MaxBaseUrlLength = 2048
	// MaxModelLength is the longest model id accepted.
	MaxModelLength = 200
	// MaxApiKeyLength is the longest api key accepted.
	MaxApiKeyLength = 4096
	// completionsPath is what the base url is completed with.
	completionsPath = "/chat/completions"
)

// Settings is the one agent-setting record: whether the chat answers, the
// provider it talks to, the limits it holds to, whether a visitor may ask for
// a summary and how many days one is kept. A limit of 0 is no limit; a
// retention of privacy.Forever keeps the summaries until a root deletes them.
type Settings struct {
	Enabled              bool
	BaseUrl              string
	ApiKey               string
	Model                string
	MaxTokens            int64
	DailyLimit           int64
	IpDailyLimit         int64
	RatePerMinute        int64
	SummariesEnabled     bool
	SummaryRetentionDays int64
}

// SettingsInput is what a form or a JSON body sends to change the settings,
// with the ids of the notices it acknowledged. The api key is write-only: ""
// keeps the one saved, ClearApiKey removes it.
type SettingsInput struct {
	Enabled              bool
	BaseUrl              string
	ApiKey               string
	ClearApiKey          bool
	Model                string
	MaxTokens            int64
	DailyLimit           int64
	IpDailyLimit         int64
	RatePerMinute        int64
	SummariesEnabled     bool
	SummaryRetentionDays int64
	Acknowledge          []string
}

// DefaultSettings are the settings before a root saved any: the chat and the
// summaries off — private by default, as the GDPR art. 25 and the LGPD art.
// 46 §2 ask — and the summaries kept 90 days once a root turns them on.
func DefaultSettings(sandbox *api.Sandbox) Settings {
	return Settings{MaxTokens: DefaultMaxTokens, RatePerMinute: DefaultRatePerMinute, SummaryRetentionDays: privacy.Rule(privacy.DataSummary).Default}
}

// settingsLock is held while the agent-setting record is read or written.
var settingsLock = make(chan struct{}, 1)

// ReadSettings answers the settings, the defaults while no root saved any.
func ReadSettings(sandbox *api.Sandbox) (Settings, error) {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	record, ok, err := findSetting(sandbox)
	if err != nil || !ok {
		return DefaultSettings(sandbox), err
	}
	settings := Settings{
		Enabled:              record.Enabled == 1,
		BaseUrl:              record.BaseUrl,
		ApiKey:               record.ApiKey,
		Model:                record.Model,
		MaxTokens:            record.MaxTokens,
		DailyLimit:           record.DailyLimit,
		IpDailyLimit:         record.IpDailyLimit,
		RatePerMinute:        record.RatePerMinute,
		SummariesEnabled:     record.SummariesEnabled == 1,
		SummaryRetentionDays: record.SummaryRetentionDays,
	}
	if settings.SummaryRetentionDays < 0 || settings.SummaryRetentionDays > privacy.MaxRetentionDays {
		settings.SummaryRetentionDays = privacy.Forever
	}
	if settings.MaxTokens < 1 {
		settings.MaxTokens = DefaultMaxTokens
	}
	if settings.RatePerMinute < 1 {
		settings.RatePerMinute = DefaultRatePerMinute
	}
	return settings, nil
}

// findSetting is the agent-setting record — the first, were there ever two —
// false when there is none. The caller holds settingsLock.
func findSetting(sandbox *api.Sandbox) (agent_db.AgentSettingRecord, bool, error) {
	records, err := agent_db.New(sandbox).ListAgentSettings(agent_db.AgentSettingFilter{})
	if err != nil {
		return agent_db.AgentSettingRecord{}, false, err
	}
	if len(records) == 0 {
		return agent_db.AgentSettingRecord{}, false, nil
	}
	return records[0], true, nil
}

// ApplySettings checks input and saves it, keeping the saved api key unless
// input sends a new one or clears it. A refused value comes back as the
// refusal, and nothing is written. Turning the chat on asks for the agent-v1
// notice, turning the summaries on for summaries-v1, and a summary retention
// past its warning for the retention notice: until input acknowledges them
// the refusal is a 409 listing them. Nothing records that a notice was
// acknowledged.
func ApplySettings(sandbox *api.Sandbox, input SettingsInput) (Settings, *Refusal, error) {
	current, err := ReadSettings(sandbox)
	if err != nil {
		return Settings{}, nil, err
	}
	settings, refusal := CheckSettings(sandbox, input, current)
	if refusal != nil {
		return settings, refusal, nil
	}
	if missing := privacy.NeedsAcknowledgement(RequiredNotices(current, settings), input.Acknowledge); missing != nil {
		return settings, &Refusal{Status: missing.Status, Field: missing.Field, Message: missing.Message, Notices: missing.Notices}, nil
	}
	return settings, nil, saveSettings(sandbox, settings)
}

// RequiredNotices is every notice turning before into after asks for.
func RequiredNotices(before Settings, after Settings) []string {
	notices := []string{}
	if after.Enabled && !before.Enabled {
		notices = append(notices, privacy.NoticeAgent)
	}
	if after.SummariesEnabled && !before.SummariesEnabled {
		notices = append(notices, privacy.NoticeSummaries)
	}
	return privacy.RetentionNotice(notices, privacy.DataSummary, before.SummaryRetentionDays, after.SummaryRetentionDays, 0)
}

// CheckSettings answers input as the settings to save over current, trimmed,
// or the refusal of the first value that does not hold.
func CheckSettings(sandbox *api.Sandbox, input SettingsInput, current Settings) (Settings, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	settings := Settings{
		Enabled:              input.Enabled,
		BaseUrl:              strings.TrimSpace(input.BaseUrl),
		ApiKey:               current.ApiKey,
		Model:                strings.TrimSpace(input.Model),
		MaxTokens:            input.MaxTokens,
		DailyLimit:           input.DailyLimit,
		IpDailyLimit:         input.IpDailyLimit,
		RatePerMinute:        input.RatePerMinute,
		SummariesEnabled:     input.SummariesEnabled,
		SummaryRetentionDays: input.SummaryRetentionDays,
	}
	if key := strings.TrimSpace(input.ApiKey); key != "" {
		settings.ApiKey = key
	}
	if input.ClearApiKey {
		settings.ApiKey = ""
	}

	settings.BaseUrl = strings.TrimSuffix(strings.TrimRight(settings.BaseUrl, "/"), completionsPath)
	if settings.BaseUrl != "" {
		if len(settings.BaseUrl) > MaxBaseUrlLength || hasControl(settings.BaseUrl, false) ||
			!matches(sandbox, `^https?://[^/?#@\s]+(/[^?#\s]*)?$`, settings.BaseUrl) {
			return settings, refuse("base_url", "The base URL is an http:// or https:// address with no query, such as https://api.openai.com/v1 — "+completionsPath+" is added to it.")
		}
	}
	if len(settings.Model) > MaxModelLength || hasControl(settings.Model, false) {
		return settings, refuse("model", "The model is one id with no space, such as gpt-4o-mini.")
	}
	if len(settings.ApiKey) > MaxApiKeyLength || hasControl(settings.ApiKey, false) {
		return settings, refuse("api_key", "The API key is one word with no space, at most 4096 characters.")
	}
	if settings.MaxTokens < 1 || settings.MaxTokens > MaxMaxTokens {
		return settings, refuse("max_tokens", "The max tokens is a number from 1 to 1000000.")
	}
	if settings.DailyLimit < 0 || settings.DailyLimit > MaxTokenLimit {
		return settings, refuse("daily_limit", "The daily limit is a number of tokens, or blank for no limit.")
	}
	if settings.IpDailyLimit < 0 || settings.IpDailyLimit > MaxTokenLimit {
		return settings, refuse("ip_daily_limit", "The ip daily limit is a number of tokens, or blank for no limit.")
	}
	if settings.RatePerMinute < 1 || settings.RatePerMinute > MaxRatePerMinute {
		return settings, refuse("rate_per_minute", "The rate per minute is a number of requests from 1 to 10000.")
	}
	if refusal := privacy.CheckRetention(sandbox, "summary_retention_days", settings.SummaryRetentionDays); refusal != nil {
		return settings, refuse(refusal.Field, refusal.Message)
	}
	if settings.Enabled && (settings.BaseUrl == "" || settings.Model == "") {
		return settings, refuse("enabled", "Set the base URL and the model before turning the agent on.")
	}
	return settings, nil
}

// saveSettings writes settings on the agent-setting record, adding it the
// first time.
func saveSettings(sandbox *api.Sandbox, settings Settings) error {
	settingsLock <- struct{}{}
	defer func() { <-settingsLock }()
	enabled := int64(0)
	if settings.Enabled {
		enabled = 1
	}
	summaries := int64(0)
	if settings.SummariesEnabled {
		summaries = 1
	}
	db := agent_db.New(sandbox)
	record, ok, err := findSetting(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		_, err := db.AddAgentSetting(agent_db.AgentSettingInput{
			Enabled:              enabled,
			BaseUrl:              settings.BaseUrl,
			ApiKey:               settings.ApiKey,
			Model:                settings.Model,
			MaxTokens:            settings.MaxTokens,
			DailyLimit:           settings.DailyLimit,
			IpDailyLimit:         settings.IpDailyLimit,
			RatePerMinute:        settings.RatePerMinute,
			SummariesEnabled:     summaries,
			SummaryRetentionDays: settings.SummaryRetentionDays,
		})
		return err
	}
	for _, write := range []func() error{
		func() error { return db.SetAgentSettingEnabled(record.Id, enabled) },
		func() error { return db.SetAgentSettingBaseUrl(record.Id, settings.BaseUrl) },
		func() error { return db.SetAgentSettingApiKey(record.Id, settings.ApiKey) },
		func() error { return db.SetAgentSettingModel(record.Id, settings.Model) },
		func() error { return db.SetAgentSettingMaxTokens(record.Id, settings.MaxTokens) },
		func() error { return db.SetAgentSettingDailyLimit(record.Id, settings.DailyLimit) },
		func() error { return db.SetAgentSettingIpDailyLimit(record.Id, settings.IpDailyLimit) },
		func() error { return db.SetAgentSettingRatePerMinute(record.Id, settings.RatePerMinute) },
		func() error { return db.SetAgentSettingSummariesEnabled(record.Id, summaries) },
		func() error { return db.SetAgentSettingSummaryRetentionDays(record.Id, settings.SummaryRetentionDays) },
	} {
		if err := write(); err != nil {
			return err
		}
	}
	return nil
}

// ApiKeyEnv is the environment variable whose api key wins over the one saved:
// EMISSARIO_AGENT_API_KEY, named after the project like the backoffice secret.
func ApiKeyEnv(sandbox *api.Sandbox) string {
	name := []byte(sandbox.Deps.StringsDeps.ToUpper(sandbox.Config.ProjectName))
	for i, char := range name {
		if !(char >= 'A' && char <= 'Z') && !(char >= '0' && char <= '9') {
			name[i] = '_'
		}
	}
	return string(name) + "_AGENT_API_KEY"
}

// ApiKey is the key sent to the provider — ApiKeyEnv when it is set, else the
// one saved, "" for none — and whether it came from the environment.
func ApiKey(sandbox *api.Sandbox, settings Settings) (string, bool) {
	if key := sandbox.Deps.StringsDeps.TrimSpace(sandbox.Deps.EnvDeps.Getenv(ApiKeyEnv(sandbox))); key != "" {
		return key, true
	}
	return settings.ApiKey, false
}

// KeyHint is how an api key is shown: its last four characters behind an
// ellipsis (…abcd), only the ellipsis for a key too short to show any, "" for
// none. The key itself never leaves the server.
func KeyHint(sandbox *api.Sandbox, key string) string {
	if key == "" {
		return ""
	}
	if len(key) < 12 {
		return "…"
	}
	return "…" + key[len(key)-4:]
}

// localRanges are the addresses of this machine and of private networks: a
// provider there keeps the conversations on the root's own infrastructure.
var localRanges = []string{
	"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "169.254.0.0/16",
	"::1/128", "fc00::/7", "fe80::/10",
}

// ProviderHost is the host of the base url of settings, with its port as the
// url spells it, "" while none is set.
func ProviderHost(sandbox *api.Sandbox, settings Settings) string {
	strings := sandbox.Deps.StringsDeps
	host := strings.TrimPrefix(strings.TrimPrefix(settings.BaseUrl, "https://"), "http://")
	return strings.Split(host, "/")[0]
}

// ProviderIsLocal tells whether host — as ProviderHost answers it — is this
// machine or a private network: localhost, a name under .localhost, a name
// of one label (a container or a machine of the local network, which no
// public DNS answers), or a loopback, private or link-local address. Any
// other host is a third party the conversations are transferred to.
func ProviderIsLocal(sandbox *api.Sandbox, host string) bool {
	strings := sandbox.Deps.StringsDeps
	name := strings.ToLower(host)
	if strings.HasPrefix(name, "[") {
		name = strings.TrimPrefix(strings.Split(name, "]")[0], "[")
	} else if index := strings.LastIndex(name, ":"); index >= 0 {
		name = name[:index]
	}
	if name == "" {
		return false
	}
	if name == "localhost" || strings.HasSuffix(name, ".localhost") {
		return true
	}
	if _, ok := backofficehttp.ParseIp(sandbox, name); ok {
		for _, entry := range localRanges {
			if backofficehttp.IpMatches(sandbox, entry, name) {
				return true
			}
		}
		return false
	}
	return !strings.Contains(strings.TrimSuffix(name, "."), ".")
}

// CompletionsUrl is the endpoint the provider answers on.
func CompletionsUrl(sandbox *api.Sandbox, settings Settings) string {
	return settings.BaseUrl + completionsPath
}

// Status is what the overview and the settings page say about the agent: why
// the chat would answer 503, and how its source and its writer are doing.
type Status struct {
	// Enabled is the switch of the settings.
	Enabled bool
	// Configured is true once the base url and the model are set.
	Configured bool
	// HasKey is true when an api key is saved or set in the environment.
	HasKey bool
	// KeyFromEnv is true when the key comes from ApiKeyEnv.
	KeyFromEnv bool
	// KeyHint is how the key in use is shown.
	KeyHint string
	// Readme is true when instructions/README.md is in the agent source.
	Readme bool
	// Summary is true when instructions/SUMMARY.md is.
	Summary bool
	// InstructionsDir is the folder the agent reads.
	InstructionsDir string
	// Running is true when the writer runs in this process.
	Running bool
	// Dropped is how many writes found the queue full since the start.
	Dropped int64
	// SourceSaved is true once the agent source was saved.
	SourceSaved bool
	// Source is what the sync loop of the agent source is doing.
	Source websitesource.Status
	// Answering is true when the chat would answer, Problem why not.
	Answering bool
	Problem   string
}

// CurrentStatus is the Status of the agent and the settings it read.
func CurrentStatus(sandbox *api.Sandbox) (Status, Settings, error) {
	settings, err := ReadSettings(sandbox)
	if err != nil {
		return Status{}, settings, err
	}
	_, sourceSaved, err := websitesource.Read(sandbox, websitesource.AgentName)
	if err != nil {
		return Status{}, settings, err
	}
	key, fromEnv := ApiKey(sandbox, settings)
	io := sandbox.Deps.IoDeps
	status := Status{
		Enabled:         settings.Enabled,
		Configured:      settings.BaseUrl != "" && settings.Model != "",
		HasKey:          key != "",
		KeyFromEnv:      fromEnv,
		KeyHint:         KeyHint(sandbox, key),
		Readme:          io.IsFile(io.Join(InstructionsDir(sandbox), ReadmeFile)),
		Summary:         io.IsFile(io.Join(InstructionsDir(sandbox), SummaryFile)),
		InstructionsDir: InstructionsDir(sandbox),
		Running:         Running(sandbox),
		Dropped:         Dropped(sandbox),
		SourceSaved:     sourceSaved,
		Source:          websitesource.CurrentStatus(sandbox, websitesource.AgentName),
	}
	if refusal := ready(sandbox, settings); refusal != nil {
		status.Problem = problemOf(refusal.Message)
	} else {
		status.Answering = true
	}
	return status, settings, nil
}

// problemOf is the sentence the backoffice shows for a 503 code of ready.
func problemOf(code string) string {
	switch code {
	case CodeDisabled:
		return "The agent is off: turn it on in the settings."
	case CodeNotConfigured:
		return "The base URL or the model is not set."
	case CodeNoInstructions:
		return "The agent source holds no instructions/README.md: set and sync the agent source."
	case CodeNotRunning:
		return "The agent does not run in this process: it runs in the one start-server started."
	}
	return code
}

// ready is why the chat answers 503 with settings, nil when it answers: off,
// not configured, no README.md in the agent source, or no writer in this
// process.
func ready(sandbox *api.Sandbox, settings Settings) *Refusal {
	unavailable := func(code string) *Refusal {
		return &Refusal{Status: api.StatusServiceUnavailable, Message: code}
	}
	if !settings.Enabled {
		return unavailable(CodeDisabled)
	}
	if settings.BaseUrl == "" || settings.Model == "" {
		return unavailable(CodeNotConfigured)
	}
	io := sandbox.Deps.IoDeps
	if !io.IsFile(io.Join(InstructionsDir(sandbox), ReadmeFile)) {
		return unavailable(CodeNoInstructions)
	}
	if !Running(sandbox) {
		return unavailable(CodeNotRunning)
	}
	return nil
}

// SettingsForm is the settings form as it was posted, every field text, and
// the notices it acknowledged.
type SettingsForm struct {
	Enabled              string
	BaseUrl              string
	ApiKey               string
	ClearApiKey          string
	Model                string
	MaxTokens            string
	DailyLimit           string
	IpDailyLimit         string
	RatePerMinute        string
	SummariesEnabled     string
	SummaryRetentionDays string
	Acknowledge          []string
}

// InputOfForm reads form as a SettingsInput: a checkbox is on as 1 or on; a
// blank max tokens is DefaultMaxTokens, a blank rate DefaultRatePerMinute, a
// blank limit no limit and a blank summary retention privacy.Forever. A
// number that does not read is refused.
func InputOfForm(sandbox *api.Sandbox, form SettingsForm) (SettingsInput, *Refusal) {
	checked := func(value string) bool { return value == "1" || value == "on" }
	input := SettingsInput{
		Enabled:          checked(form.Enabled),
		BaseUrl:          form.BaseUrl,
		ApiKey:           form.ApiKey,
		ClearApiKey:      checked(form.ClearApiKey),
		Model:            form.Model,
		SummariesEnabled: checked(form.SummariesEnabled),
		Acknowledge:      form.Acknowledge,
	}
	retention, refusal := privacy.ParseRetention(sandbox, "summary_retention_days", form.SummaryRetentionDays)
	if refusal != nil {
		return input, refuse(refusal.Field, refusal.Message)
	}
	input.SummaryRetentionDays = retention
	for _, number := range []struct {
		text     string
		fallback int64
		target   *int64
		field    string
		message  string
	}{
		{form.MaxTokens, DefaultMaxTokens, &input.MaxTokens, "max_tokens", "The max tokens is a number from 1 to 1000000, or blank for 1024."},
		{form.DailyLimit, 0, &input.DailyLimit, "daily_limit", "The daily limit is a number of tokens, or blank for no limit."},
		{form.IpDailyLimit, 0, &input.IpDailyLimit, "ip_daily_limit", "The ip daily limit is a number of tokens, or blank for no limit."},
		{form.RatePerMinute, DefaultRatePerMinute, &input.RatePerMinute, "rate_per_minute", "The rate per minute is a number of requests from 1 to 10000, or blank for 10."},
	} {
		text := sandbox.Deps.StringsDeps.TrimSpace(number.text)
		if text == "" {
			*number.target = number.fallback
			continue
		}
		value, err := sandbox.Deps.StringsDeps.ParseInt(text, 10, 64)
		if err != nil {
			return input, refuse(number.field, number.message)
		}
		*number.target = value
	}
	return input, nil
}

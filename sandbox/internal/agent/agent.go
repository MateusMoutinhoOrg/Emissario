package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// The agent is a public chat a root configures in the backoffice: anonymous
// visitors of the website talk to a model behind an OpenAI-compatible Chat
// Completions endpoint, through POST /api/agent/messages, and may hand the
// conversation over as an anonymous summary through POST
// /api/agent/summaries. Emissario holds no conversation: the history lives in
// the visitor's browser, signed by the server (signature.go) so it cannot be
// forged. What the model is told comes from the agent source, a git
// repository synced into <FontsDir>/agent (websitesource): instructions/README.md
// is its system prompt, and every file under instructions/ is one it can list
// and read (tools.go), beside the urls a root declared (urls.go).
//
// Every call to the provider is limited — per ip per minute, per day and per
// ip per day, in tokens (limits.go) — and recorded as an agent-usage; usage
// and summaries reach agent-db through a queue (queue.go), held while a
// backup is taken, so the chat keeps answering meanwhile.

const (
	// DashboardPath is the overview page of the agent, the top-bar link.
	DashboardPath = "/admin/agent"
	// SummariesPath lists the summaries.
	SummariesPath = "/admin/list-agent-summaries"
	// SummaryPath shows one summary, ?id=<ref>.
	SummaryPath = "/admin/get-agent-summary"
	// UsagePath is the token usage page.
	UsagePath = "/admin/agent-usage"
	// RemoveSummariesPath removes one summary, or every one before a date.
	RemoveSummariesPath = "/admin/root/remove-agent-summaries"
	// SettingsPath is the settings page, and where its form posts.
	SettingsPath = "/admin/root/set-agent-settings"
	// UrlsPath lists the urls the agent can call.
	UrlsPath = "/admin/root/list-agent-urls"
	// AddUrlPath is the form that adds a url, and where it posts.
	AddUrlPath = "/admin/root/add-agent-url"
	// SetUrlPath is the form that changes a url, ?id=, and where it posts.
	SetUrlPath = "/admin/root/set-agent-url"
	// RemoveUrlPath removes a url.
	RemoveUrlPath = "/admin/root/remove-agent-url"
	// SourcePath is the page that sets the agent source.
	SourcePath = websitesource.AgentPagePath
	// SyncSourcePath syncs the agent source now.
	SyncSourcePath = "/admin/root/sync-agent-source"
)

const (
	// MaxMessages is the most messages one conversation holds.
	MaxMessages = 50
	// MaxContentLength is the most characters one message holds; a reply of
	// the model is cut to it, so the history the visitor sends back holds.
	MaxContentLength = 4000
	// MaxBodyBytes is the largest body /api/agent/* reads.
	MaxBodyBytes = 64 * 1024
	// MaxToolRounds is how many rounds of tool calls one message may take;
	// the call after the last answers with no tool.
	MaxToolRounds = 5
	// DeadlineMilliseconds is the time one request to /api/agent/* has, every
	// provider call and tool call included.
	DeadlineMilliseconds = 60 * 1000
	// WriteDeadlineMilliseconds is the write deadline those requests extend
	// themselves to: the deadline, and room to write the answer.
	WriteDeadlineMilliseconds = DeadlineMilliseconds + 5*1000
	// UrlTimeoutMilliseconds bounds one call_url.
	UrlTimeoutMilliseconds = 10 * 1000
	// MaxUrlResponseBytes is how much of what a url answers the model reads.
	MaxUrlResponseBytes = 32 * 1024
	// MaxInstructionBytes is how much of one instruction file is read: the
	// README.md and SUMMARY.md included.
	MaxInstructionBytes = 64 * 1024
	// MaxInstructionFiles is how many paths list_instructions answers.
	MaxInstructionFiles = 1000
	// MaxProviderResponseBytes is how much of a provider answer is read.
	MaxProviderResponseBytes = 4 * 1024 * 1024
	// MaxSummaryLength is the most characters a stored summary holds.
	MaxSummaryLength = 8000
	// InstructionsFolder is the folder of the agent source the agent reads.
	InstructionsFolder = "instructions"
	// ReadmeFile is the system instructions of the agent, under
	// InstructionsFolder: without it the chat answers 503.
	ReadmeFile = "README.md"
	// SummaryFile is the optional extra instructions of the summaries.
	SummaryFile = "SUMMARY.md"
	// QueueSize is how many writes wait for the writer at most.
	QueueSize = 10000
	// UsageRetentionDays is how long an agent-usage is kept: it holds an ip.
	UsageRetentionDays = 90
	// DefaultPeriodDays is the period a summaries or usage page shows by
	// default.
	DefaultPeriodDays = 30
	// MaxPeriodDays is the longest period one query covers.
	MaxPeriodDays = 366
	// MaxIpRows is how many ips the usage report lists.
	MaxIpRows = 100
)

const (
	// StatusBadGateway answers a provider that failed.
	StatusBadGateway = 502
	// StatusGatewayTimeout answers a request past DeadlineMilliseconds.
	StatusGatewayTimeout = 504
)

// The codes /api/agent/* answers in the error of {"error","field"}.
const (
	CodeInvalidSignature = "invalid-signature"
	CodeRateLimit        = "rate-limit"
	CodeIpDailyLimit     = "ip-daily-limit"
	CodeDailyLimit       = "daily-limit"
	CodeDisabled         = "disabled"
	CodeNotConfigured    = "not-configured"
	CodeNoInstructions   = "no-instructions"
	CodeNotRunning       = "not-running"
	CodeRestoring        = "restoring"
	CodeProviderFailed   = "provider-failed"
	CodeDeadline         = "deadline"
	CodeOrigin           = "origin"
)

const (
	// KindMessage is an agent-usage of a message.
	KindMessage = "message"
	// KindSummary is an agent-usage of a summary; it holds no ip.
	KindSummary = "summary"
)

// The notices a page shows after a form, from its ?notice=.
const (
	NoticeSaved   = "saved"
	NoticeAdded   = "added"
	NoticeRemoved = "removed"
)

// Refusal is why a request is not answered as asked: its status, the field it
// blames ("" for none), the message — a code of /api/agent/* or a sentence of
// the backoffice — and, for a 429 or a 503, the Retry-After in seconds.
type Refusal struct {
	Status     int
	Field      string
	Message    string
	RetryAfter int64
}

// refuse is a 400 blaming field.
func refuse(field string, message string) *Refusal {
	return &Refusal{Status: api.StatusBadRequest, Field: field, Message: message}
}

// Fail answers refusal: its Retry-After, then the failure the route returns.
func Fail(sandbox *api.Sandbox, setHeader func(key string, value string), refusal *Refusal) error {
	if refusal.RetryAfter > 0 {
		setHeader("Retry-After", sandbox.Deps.StringsDeps.FormatInt(refusal.RetryAfter, 10))
	}
	return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
}

// NowSeconds is the current instant, in Unix seconds.
func NowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000000
}

// nowMilliseconds is the current instant, in Unix milliseconds.
func nowMilliseconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000
}

// InstructionsDir is the folder the agent reads: the instructions/ of the
// checkout of the agent source.
func InstructionsDir(sandbox *api.Sandbox) string {
	return sandbox.Deps.IoDeps.Join(websitesource.RepoDir(sandbox, websitesource.AgentName), InstructionsFolder)
}

// cut is text cut to at most length bytes, on a character boundary.
func cut(text string, length int) string {
	if len(text) <= length {
		return text
	}
	end := length
	for end > 0 && text[end]&0xC0 == 0x80 {
		end--
	}
	return text[:end]
}

// cutRunes is text cut to at most length characters.
func cutRunes(text string, length int) string {
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length])
}

// isText tells whether content is valid UTF-8 holding no NUL byte: what a file
// or an answer must be for the model to read it.
func isText(content []byte) bool {
	text := string(content)
	if string([]rune(text)) != text {
		return false
	}
	for _, char := range content {
		if char == 0 {
			return false
		}
	}
	return true
}

// hasControl tells whether text holds a space-less control character: a tab,
// a line break or anything below the space but the ones allowed.
func hasControl(text string, allowSpace bool) bool {
	for _, letter := range text {
		if letter < ' ' || letter == 0x7f || (!allowSpace && letter == ' ') {
			return true
		}
	}
	return false
}

// matches tells whether text matches the regular expression pattern; a
// pattern that does not compile matches nothing.
func matches(sandbox *api.Sandbox, pattern string, text string) bool {
	ok, err := sandbox.Deps.StringsDeps.MatchPattern(pattern, text)
	return err == nil && ok
}

// Admit is why a request to /api/agent/* from ip is refused before its
// conversation is looked at, nil when it goes on: 503 while a restore writes
// the databases, 403 for a request another site's page sent — there is no
// CORS — and 429 past the rate per minute of ip, which every request counts in.
func Admit(sandbox *api.Sandbox, ip string, origin string, host string) *Refusal {
	if snapshots.Restoring(sandbox) {
		return &Refusal{Status: api.StatusServiceUnavailable, Message: CodeRestoring, RetryAfter: 5}
	}
	if !backofficehttp.SameOrigin(sandbox, origin, host) {
		return &Refusal{Status: api.StatusForbidden, Field: "origin", Message: CodeOrigin}
	}
	settings, err := ReadSettings(sandbox)
	if err != nil {
		return failed(sandbox, "the agent settings could not be read", err)
	}
	return CheckRate(sandbox, settings, ip)
}

// ReplyJSON is {"message": {"role": "assistant", "content": reply},
// "signature"}: what /api/agent/messages answers.
func ReplyJSON(sandbox *api.Sandbox, reply string, signature string) *serializabledeps.SerializableObject {
	message := sandbox.Deps.SerializableDeps.CreateObject()
	message.AddItemToObject("role", RoleAssistant)
	message.AddItemToObject("content", reply)
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("message", message)
	document.AddItemToObject("signature", signature)
	return document
}

// SummaryCreatedJSON is {"id": ref}: what /api/agent/summaries answers.
func SummaryCreatedJSON(sandbox *api.Sandbox, ref string) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("id", ref)
	return document
}

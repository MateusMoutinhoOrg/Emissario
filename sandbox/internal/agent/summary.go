package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
)

// A summary is what is kept of a conversation, at the visitor's request: the
// configured model writes it under the mandatory summary instructions —
// remove names, dates, places, ids, CPF, phones, emails, addresses — and
// SUMMARY.md, then Redact takes out what still looks like a CPF, a CNPJ, an
// email, a phone or a card number. The summary is anonymous: it holds no ip,
// and its agent-usage holds none either. The conversation itself is never
// stored.

// summaryPromptAsset is the mandatory part of the summary instructions.
const summaryPromptAsset = "agent/summary.md"

// refBytes is how many random bytes the id of a summary holds, in hex.
const refBytes = 8

// Summarize writes the summary of messages — checked by CheckConversation —
// for the visitor at ip, queues it as an agent-summary, and answers its id.
func Summarize(sandbox *api.Sandbox, ip string, messages []Message) (string, *Refusal) {
	deadline := nowMilliseconds(sandbox) + DeadlineMilliseconds
	settings, err := ReadSettings(sandbox)
	if err != nil {
		return "", failed(sandbox, "the agent settings could not be read", err)
	}
	if refusal := ready(sandbox, settings); refusal != nil {
		return "", refusal
	}
	if !settings.SummariesEnabled {
		return "", &Refusal{Status: api.StatusServiceUnavailable, Message: CodeSummariesOff}
	}
	system, refusal := systemPrompt(sandbox, summaryPromptAsset, SummaryFile, false)
	if refusal != nil {
		return "", refusal
	}
	completion, refusal := call(sandbox, settings, ip, KindSummary, completionRequest{
		Settings: settings,
		Messages: []ChatMessage{
			{Role: "system", Content: system},
			{Role: RoleUser, Content: transcript(sandbox, messages)},
		},
		Deadline: deadline,
	})
	if refusal != nil {
		return "", refusal
	}
	summary := cutRunes(Redact(sandbox, sandbox.Deps.StringsDeps.TrimSpace(completion.Content)), MaxSummaryLength)
	if summary == "" {
		sandbox.Deps.StdDeps.Eprintf("warning: the agent provider at %s answered an empty summary\n", settings.BaseUrl)
		return "", providerFailed()
	}
	ref, err := sandbox.Deps.RandDeps.Hex(refBytes)
	if err != nil {
		return "", failed(sandbox, "the id of a summary could not be drawn", err)
	}
	enqueue(sandbox, pending{Summary: &agent_db.AgentSummaryInput{
		Ref:          ref,
		At:           NowSeconds(sandbox),
		Summary:      summary,
		Messages:     int64(len(messages)),
		InputTokens:  completion.InputTokens,
		OutputTokens: completion.OutputTokens,
	}})
	return ref, nil
}

// transcript is messages as the one text the summary is made of.
func transcript(sandbox *api.Sandbox, messages []Message) string {
	text := "The conversation to summarize:\n"
	for _, message := range messages {
		speaker := "Visitor"
		if message.Role == RoleAssistant {
			speaker = "Agent"
		}
		text += "\n" + speaker + ": " + message.Content + "\n"
	}
	return text
}

// redactions are the patterns Redact takes out, in order — an email first, as
// it may hold digits, a CNPJ before a CPF, which it holds the length of, and an
// international phone before a card, which its digits could pass for.
var redactions = []struct {
	pattern     string
	replacement string
}{
	{`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`, "[email]"},
	{`\b\d{2}\.?\d{3}\.?\d{3}/?\d{4}-?\d{2}\b`, "[cnpj]"},
	{`\b\d{3}\.?\d{3}\.?\d{3}-?\d{2}\b`, "[cpf]"},
	{`\+\d{1,3}[\s.-]?(?:\(\d{2,3}\)\s?|\d{2,3}[\s.-]?)?\d{4,5}[\s.-]?\d{4}\b`, "[phone]"},
	{`\b(?:\d[ -]?){12,18}\d\b`, "[card]"},
	{`(?:\+\d{1,3}[\s.-]?)?(?:\(\d{2,3}\)\s?|\b\d{2,3}[\s.-])?\b\d{4,5}[\s.-]?\d{4}\b`, "[phone]"},
}

// Redact is text with every CPF, CNPJ, email, phone and card number it still
// holds replaced by [cpf], [cnpj], [email], [phone] and [card]: what the model
// left in a summary in spite of its instructions.
func Redact(sandbox *api.Sandbox, text string) string {
	for _, redaction := range redactions {
		replaced, err := sandbox.Deps.StringsDeps.ReplacePattern(redaction.pattern, text, redaction.replacement)
		if err == nil {
			text = replaced
		}
	}
	return text
}

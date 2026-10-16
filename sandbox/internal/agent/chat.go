package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// SystemPromptAsset is the mandatory part of the system prompt, embedded:
// what every agent holds to, whatever its README.md says.
const SystemPromptAsset = "agent/system.md"

// Reply answers the next message of the agent in the conversation messages —
// checked by CheckMessages — for the visitor at ip: the system prompt is the
// embedded mandatory instructions and README.md, and the model may call its
// tools for up to MaxToolRounds rounds, after which it answers with none. Each
// provider call is limited and recorded first (call). The reply is cut to
// MaxContentLength characters, so the history the visitor sends back holds.
func Reply(sandbox *api.Sandbox, ip string, messages []Message) (string, *Refusal) {
	deadline := nowMilliseconds(sandbox) + DeadlineMilliseconds
	settings, err := ReadSettings(sandbox)
	if err != nil {
		return "", failed(sandbox, "the agent settings could not be read", err)
	}
	if refusal := ready(sandbox, settings); refusal != nil {
		return "", refusal
	}
	system, refusal := systemPrompt(sandbox, SystemPromptAsset, ReadmeFile, true)
	if refusal != nil {
		return "", refusal
	}
	urls, err := ListUrls(sandbox)
	if err != nil {
		return "", failed(sandbox, "the agent urls could not be read", err)
	}
	tools := toolsJSON(sandbox, urls)

	conversation := []ChatMessage{{Role: "system", Content: system}}
	for _, message := range messages {
		conversation = append(conversation, ChatMessage{Role: message.Role, Content: message.Content})
	}
	for round := 0; ; round++ {
		choice := "auto"
		if round >= MaxToolRounds {
			choice = "none"
		}
		completion, refusal := call(sandbox, settings, ip, KindMessage, completionRequest{
			Settings:   settings,
			Messages:   conversation,
			Tools:      tools,
			ToolChoice: choice,
			Deadline:   deadline,
		})
		if refusal != nil {
			return "", refusal
		}
		if len(completion.ToolCalls) == 0 || round >= MaxToolRounds {
			reply := sandbox.Deps.StringsDeps.TrimSpace(completion.Content)
			if reply == "" {
				sandbox.Deps.StdDeps.Eprintf("warning: the agent provider at %s answered no text\n", settings.BaseUrl)
				return "", providerFailed()
			}
			return cutRunes(reply, MaxContentLength), nil
		}
		conversation = append(conversation, ChatMessage{Role: RoleAssistant, Content: completion.Content, ToolCalls: completion.ToolCalls})
		for _, toolCall := range completion.ToolCalls {
			conversation = append(conversation, ChatMessage{
				Role:       "tool",
				ToolCallId: toolCall.Id,
				Content:    runTool(sandbox, toolCall, urls, deadline),
			})
		}
	}
}

// call is one provider call of the visitor at ip, limited: refused when it
// could pass a daily limit, its max-tokens reserved while it runs, and its
// tokens counted and queued as an agent-usage of kind once it answered,
// keeping ip as the IP mode says — hashed by default. A summary is recorded
// with no ip — its ip only counts in memory, toward the limit of today. The
// tally counts a visitor under the key a stored row reads back as
// (privacy.TallyKey), so the limits hold across a restart.
func call(sandbox *api.Sandbox, settings Settings, ip string, kind string, request completionRequest) (Completion, *Refusal) {
	key := privacy.TallyKey(sandbox, ip)
	if refusal := reserve(sandbox, settings, key); refusal != nil {
		return Completion{}, refusal
	}
	completion, refusal := complete(sandbox, request)
	if refusal != nil {
		release(sandbox, settings, key, 0)
		return Completion{}, refusal
	}
	release(sandbox, settings, key, completion.InputTokens+completion.OutputTokens)
	recorded, err := privacy.Store(sandbox, ip)
	if err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the ip of an agent call could not be hashed: %s\n", err.Error())
		recorded = ""
	}
	if kind == KindSummary {
		recorded = ""
	}
	recordUsage(sandbox, recorded, kind, completion)
	return completion, nil
}

// systemPrompt is the embedded asset followed by the file of the agent
// source, at most MaxInstructionBytes of it; a file required and missing is
// a 503 no-instructions.
func systemPrompt(sandbox *api.Sandbox, asset string, file string, required bool) (string, *Refusal) {
	mandatory, err := sandbox.Deps.EmbedDeps.ReadFile(asset)
	if err != nil {
		return "", failed(sandbox, "the embedded "+asset+" could not be read", err)
	}
	prompt := sandbox.Deps.StringsDeps.TrimSpace(string(mandatory))
	content, _, message := ReadInstruction(sandbox, file)
	if message != "" {
		if required {
			return "", &Refusal{Status: api.StatusServiceUnavailable, Message: CodeNoInstructions}
		}
		return prompt, nil
	}
	return prompt + "\n\n---\n\n" + sandbox.Deps.StringsDeps.TrimSpace(content), nil
}

// failed says on stderr why the agent could not answer, and answers a 500.
func failed(sandbox *api.Sandbox, what string, err error) *Refusal {
	sandbox.Deps.StdDeps.Eprintf("warning: %s: %s\n", what, err.Error())
	return &Refusal{Status: api.StatusInternalServerError, Message: "the agent could not answer"}
}

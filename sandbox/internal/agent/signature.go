package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// The history of a conversation lives in the visitor's browser, so the server
// signs it: every answer carries the HMAC-SHA256 of every message so far, and
// the next request must send it back. A visitor cannot then put words in the
// mouth of the agent, nor ask for a summary of a conversation the agent never
// had. The key is derived from the backoffice secret — a secret generated for
// the run, when none is set, so a restart starts every conversation over.

// signatureContext is what the key is derived under, so it signs nothing but
// a history of the agent.
const signatureContext = "emissario-agent-history-v1"

// Message is one message of a conversation, as the visitor sends it.
type Message struct {
	// Role is user or assistant.
	Role string
	// Content is the text of the message.
	Content string
}

// Sign is the signature of messages, lower-case hexadecimal.
func Sign(sandbox *api.Sandbox, messages []Message) string {
	return sandbox.Deps.HashDeps.HmacSha256Hex(signingKey(sandbox), canonical(sandbox, messages))
}

// Verify tells whether signature is the one Sign answers for messages, in a
// time that does not depend on where they differ.
func Verify(sandbox *api.Sandbox, messages []Message, signature string) bool {
	if sandbox.Config.SessionSecret == "" || signature == "" {
		return false
	}
	expected := Sign(sandbox, messages)
	return sandbox.Deps.HashDeps.Equal([]byte(expected), []byte(sandbox.Deps.StringsDeps.ToLower(signature)))
}

// signingKey is the key of the signatures, derived from the backoffice secret.
func signingKey(sandbox *api.Sandbox) []byte {
	return []byte(sandbox.Deps.HashDeps.HmacSha256Hex([]byte(sandbox.Config.SessionSecret), []byte(signatureContext)))
}

// canonical is the one spelling of messages that is signed: each message its
// role, the length of its content and the content, so no two histories spell
// the same.
func canonical(sandbox *api.Sandbox, messages []Message) []byte {
	text := signatureContext
	for _, message := range messages {
		text += "\n" + message.Role + "\n" + sandbox.Deps.StringsDeps.FormatInt(int64(len(message.Content)), 10) + "\n" + message.Content
	}
	return []byte(text)
}

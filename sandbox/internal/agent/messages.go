package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

const (
	// RoleUser is a message of the visitor.
	RoleUser = "user"
	// RoleAssistant is a message of the agent.
	RoleAssistant = "assistant"
)

// CheckMessages is why messages is not a conversation that may ask for the
// next reply, nil when it is: one to MaxMessages messages, user and assistant
// in turns, the first and the last of the visitor, none blank nor over
// MaxContentLength characters. With more than one message, signature must be
// the one the last answer carried; the first message carries none.
func CheckMessages(sandbox *api.Sandbox, messages []Message, signature string) *Refusal {
	if refusal := checkShape(sandbox, messages, RoleUser); refusal != nil {
		return refusal
	}
	if len(messages) == 1 {
		if signature != "" {
			return invalidSignature()
		}
		return nil
	}
	if !Verify(sandbox, messages[:len(messages)-1], signature) {
		return invalidSignature()
	}
	return nil
}

// CheckConversation is why messages is not a whole conversation a summary may
// be made of, nil when it is: the shape CheckMessages asks for, ending with
// the answer of the agent, and signature the one that answer carried.
func CheckConversation(sandbox *api.Sandbox, messages []Message, signature string) *Refusal {
	if refusal := checkShape(sandbox, messages, RoleAssistant); refusal != nil {
		return refusal
	}
	if !Verify(sandbox, messages, signature) {
		return invalidSignature()
	}
	return nil
}

// invalidSignature is the 401 of a history the server did not sign.
func invalidSignature() *Refusal {
	return &Refusal{Status: api.StatusUnauthorized, Field: "signature", Message: CodeInvalidSignature}
}

// checkShape is why messages is not a conversation that begins with the
// visitor, alternates, and ends with last.
func checkShape(sandbox *api.Sandbox, messages []Message, last string) *Refusal {
	if len(messages) == 0 {
		return refuse("messages", "messages holds at least one message.")
	}
	if len(messages) > MaxMessages {
		return refuse("messages", sandbox.Deps.StdDeps.Sprintf("messages holds at most %d messages: start a new conversation.", MaxMessages))
	}
	for index, message := range messages {
		field := sandbox.Deps.StdDeps.Sprintf("messages[%d]", index)
		expected := RoleUser
		if index%2 == 1 {
			expected = RoleAssistant
		}
		if message.Role != expected {
			return refuse(field+".role", "messages alternate user and assistant, starting with user.")
		}
		if sandbox.Deps.StringsDeps.TrimSpace(message.Content) == "" {
			return refuse(field+".content", "a message is never blank.")
		}
		if len([]rune(message.Content)) > MaxContentLength {
			return refuse(field+".content", sandbox.Deps.StdDeps.Sprintf("a message holds at most %d characters.", MaxContentLength))
		}
	}
	if messages[len(messages)-1].Role != last {
		return refuse("messages", "the last message is of the "+last+".")
	}
	return nil
}

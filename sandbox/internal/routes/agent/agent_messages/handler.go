package agent_messages

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /api/agent/messages, the public chat: the next message
// of the agent in the conversation the body holds, and the signature that
// covers the conversation and it. The request gets agent.DeadlineMilliseconds
// in all, so it moves its write deadline past the server's first. Before the
// model is asked, it is refused while a restore runs, when another site sent
// it, past the rate per minute of its ip, when the conversation is not one or
// its signature not the server's, and while the agent is off or has no
// instructions — each with the code agent.Refusal carries.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	response.SetWriteDeadline(agent.WriteDeadlineMilliseconds)
	access := privacy.Access{AtMs: privacy.NowMilliseconds(sandbox), Ip: props.ClientIp, Port: props.ClientPort, Kind: privacy.AccessAgent}
	defer func() { privacy.RecordAccess(sandbox, access) }()
	if refusal := agent.Admit(sandbox, props.ClientIp, input.Origin, input.Host); refusal != nil {
		return agent.Fail(sandbox, response.SetHeader, refusal)
	}
	messages := []agent.Message{}
	for _, message := range input.Body.Messages {
		messages = append(messages, agent.Message{Role: message.Role, Content: message.Content})
	}
	if refusal := agent.CheckMessages(sandbox, messages, input.Body.Signature); refusal != nil {
		return agent.Fail(sandbox, response.SetHeader, refusal)
	}
	reply, refusal := agent.Reply(sandbox, props.ClientIp, messages)
	if refusal != nil {
		return agent.Fail(sandbox, response.SetHeader, refusal)
	}
	signature := agent.Sign(sandbox, append(messages, agent.Message{Role: agent.RoleAssistant, Content: reply}))
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, agent.ReplyJSON(sandbox, reply, signature))
}

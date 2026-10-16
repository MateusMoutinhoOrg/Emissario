package agent_summaries

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /api/agent/summaries: the whole conversation the body
// holds — ending with the last answer of the agent, under the signature that
// answer carried — is summarized by the model, the summary kept with no ip,
// and its id answered 201; the text of the summary is never answered. It is
// refused like POST /api/agent/messages, and has the same deadline.
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
	if refusal := agent.CheckConversation(sandbox, messages, input.Body.Signature); refusal != nil {
		return agent.Fail(sandbox, response.SetHeader, refusal)
	}
	ref, refusal := agent.Summarize(sandbox, props.ClientIp, messages)
	access.Ref = ref
	if refusal != nil {
		return agent.Fail(sandbox, response.SetHeader, refusal)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusCreated, agent.SummaryCreatedJSON(sandbox, ref))
}

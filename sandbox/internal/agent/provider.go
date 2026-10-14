package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
)

// The provider is any endpoint that speaks the OpenAI Chat Completions api —
// OpenAI, OpenRouter, Groq, Ollama… — at <base-url>/chat/completions, with a
// model that supports tool calling. Everything sent there leaves the server:
// the conversation, the instructions and what every tool answered.

// ChatMessage is one message of what is sent to the provider: the system
// prompt, the conversation, and the tool calls of this request with what each
// answered.
type ChatMessage struct {
	// Role is system, user, assistant or tool.
	Role string
	// Content is the text of the message.
	Content string
	// ToolCalls are the calls an assistant message asked for.
	ToolCalls []ToolCall
	// ToolCallId is the call a tool message answers.
	ToolCallId string
}

// ToolCall is one call of a tool the model asked for.
type ToolCall struct {
	// Id is the id the provider gave the call.
	Id string
	// Name is the tool called.
	Name string
	// Arguments is the JSON object of its arguments, as the model wrote it.
	Arguments string
}

// Completion is what one provider call answered.
type Completion struct {
	// Content is the text of the answer, "" when it only calls tools.
	Content string
	// ToolCalls are the tools it asked for.
	ToolCalls []ToolCall
	// InputTokens and OutputTokens are the usage the provider reported, or an
	// estimate when it reported none.
	InputTokens  int64
	OutputTokens int64
}

// completionRequest is one provider call.
type completionRequest struct {
	Settings Settings
	Messages []ChatMessage
	// Tools is the tools array, nil to offer none.
	Tools *serializabledeps.SerializableObject
	// ToolChoice is auto or none, "" to send none.
	ToolChoice string
	// Deadline is when the request must have answered, Unix milliseconds.
	Deadline int64
}

// complete sends request to the provider and reads its answer: 504 when the
// deadline passed, 502 when the provider could not be reached or answered no
// completion. The api key and the body sent are never logged.
func complete(sandbox *api.Sandbox, request completionRequest) (Completion, *Refusal) {
	remaining := request.Deadline - nowMilliseconds(sandbox)
	if remaining <= 0 {
		return Completion{}, deadlinePassed()
	}
	body := sandbox.Deps.SerializableDeps.SerializeToJson(requestJSON(sandbox, request))
	call := sandbox.Deps.RequestDeps.NewRequest(CompletionsUrl(sandbox, request.Settings))
	call.SetMethod("POST")
	call.SetTimeout(remaining)
	call.AddHeader("Content-Type", "application/json")
	call.AddHeader("Accept", "application/json")
	if key, _ := ApiKey(sandbox, request.Settings); key != "" {
		call.AddHeader("Authorization", "Bearer "+key)
	}
	call.SetBody([]byte(body))
	response, err := call.Fetch()
	if err != nil {
		if nowMilliseconds(sandbox) >= request.Deadline {
			return Completion{}, deadlinePassed()
		}
		sandbox.Deps.StdDeps.Eprintf("warning: the agent provider at %s could not be reached: %s\n", request.Settings.BaseUrl, err.Error())
		return Completion{}, providerFailed()
	}
	defer response.Close()
	answer, err := response.ReadBody(MaxProviderResponseBytes)
	if err != nil {
		if nowMilliseconds(sandbox) >= request.Deadline {
			return Completion{}, deadlinePassed()
		}
		sandbox.Deps.StdDeps.Eprintf("warning: the answer of the agent provider at %s could not be read: %s\n", request.Settings.BaseUrl, err.Error())
		return Completion{}, providerFailed()
	}
	status := response.GetStatusCode()
	if status < 200 || status > 299 {
		sandbox.Deps.StdDeps.Eprintf("warning: the agent provider at %s answered %d: %s\n", request.Settings.BaseUrl, status, oneLine(sandbox, cut(string(answer), 300)))
		return Completion{}, providerFailed()
	}
	completion, ok := parseCompletion(sandbox, string(answer))
	if !ok {
		sandbox.Deps.StdDeps.Eprintf("warning: the agent provider at %s answered no completion: %s\n", request.Settings.BaseUrl, oneLine(sandbox, cut(string(answer), 300)))
		return Completion{}, providerFailed()
	}
	if completion.InputTokens == 0 && completion.OutputTokens == 0 {
		completion.InputTokens = estimateTokens(len(body))
		output := len(completion.Content)
		for _, toolCall := range completion.ToolCalls {
			output += len(toolCall.Name) + len(toolCall.Arguments)
		}
		completion.OutputTokens = estimateTokens(output)
	}
	return completion, nil
}

// estimateTokens is how many tokens length bytes count as when a provider
// reports no usage: one per four bytes, rounded up — so the limits still hold.
func estimateTokens(length int) int64 {
	return int64((length + 3) / 4)
}

// deadlinePassed is the 504 of a request past DeadlineMilliseconds.
func deadlinePassed() *Refusal {
	return &Refusal{Status: StatusGatewayTimeout, Message: CodeDeadline}
}

// providerFailed is the 502 of a provider that failed.
func providerFailed() *Refusal {
	return &Refusal{Status: StatusBadGateway, Message: CodeProviderFailed}
}

// oneLine is text with its line breaks as spaces, for one line of stderr.
func oneLine(sandbox *api.Sandbox, text string) string {
	strings := sandbox.Deps.StringsDeps
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r", " "), "\n", " ")
}

// requestJSON is the body of one Chat Completions call.
func requestJSON(sandbox *api.Sandbox, request completionRequest) *serializabledeps.SerializableObject {
	json := sandbox.Deps.SerializableDeps
	messages := json.CreateArray()
	for _, message := range request.Messages {
		messages.AddItemToArray(messageJSON(sandbox, message))
	}
	document := json.CreateObject()
	document.AddItemToObject("model", request.Settings.Model)
	document.AddItemToObject("max_tokens", request.Settings.MaxTokens)
	document.AddItemToObject("messages", messages)
	if request.Tools != nil {
		document.AddItemToObject("tools", request.Tools)
		if request.ToolChoice != "" {
			document.AddItemToObject("tool_choice", request.ToolChoice)
		}
	}
	return document
}

// messageJSON is one message of the messages of a call.
func messageJSON(sandbox *api.Sandbox, message ChatMessage) *serializabledeps.SerializableObject {
	json := sandbox.Deps.SerializableDeps
	item := json.CreateObject()
	item.AddItemToObject("role", message.Role)
	if message.Role == "tool" {
		item.AddItemToObject("tool_call_id", message.ToolCallId)
	}
	if len(message.ToolCalls) == 0 {
		item.AddItemToObject("content", message.Content)
		return item
	}
	if message.Content == "" {
		item.AddItemToObject("content", json.CreateNull())
	} else {
		item.AddItemToObject("content", message.Content)
	}
	calls := json.CreateArray()
	for _, toolCall := range message.ToolCalls {
		function := json.CreateObject()
		function.AddItemToObject("name", toolCall.Name)
		function.AddItemToObject("arguments", toolCall.Arguments)
		call := json.CreateObject()
		call.AddItemToObject("id", toolCall.Id)
		call.AddItemToObject("type", "function")
		call.AddItemToObject("function", function)
		calls.AddItemToArray(call)
	}
	item.AddItemToObject("tool_calls", calls)
	return item
}

// parseCompletion reads the first choice and the usage of a Chat Completions
// answer, false when it holds no message.
func parseCompletion(sandbox *api.Sandbox, answer string) (Completion, bool) {
	document, err := sandbox.Deps.SerializableDeps.ParseJson(answer)
	if err != nil || !document.IsObject() {
		return Completion{}, false
	}
	choices := item(document, "choices")
	if choices == nil || !choices.IsArray() {
		return Completion{}, false
	}
	size, _ := choices.GetArraySize()
	if size == 0 {
		return Completion{}, false
	}
	message := item(choices.GetArrayItem(0), "message")
	if message == nil || !message.IsObject() {
		return Completion{}, false
	}
	completion := Completion{Content: text(message, "content")}
	if calls := item(message, "tool_calls"); calls != nil && calls.IsArray() {
		count, _ := calls.GetArraySize()
		for index := 0; index < count; index++ {
			call := calls.GetArrayItem(index)
			function := item(call, "function")
			if function == nil {
				continue
			}
			arguments := text(function, "arguments")
			if raw := item(function, "arguments"); raw != nil && raw.IsObject() {
				arguments = sandbox.Deps.SerializableDeps.SerializeToJson(raw)
			}
			completion.ToolCalls = append(completion.ToolCalls, ToolCall{
				Id:        text(call, "id"),
				Name:      text(function, "name"),
				Arguments: arguments,
			})
		}
	}
	if usage := item(document, "usage"); usage != nil && usage.IsObject() {
		completion.InputTokens = number(usage, "prompt_tokens")
		completion.OutputTokens = number(usage, "completion_tokens")
	}
	return completion, true
}

// item is the value of key in object, nil when object is no object or has no
// such key.
func item(object *serializabledeps.SerializableObject, key string) *serializabledeps.SerializableObject {
	if object == nil || !object.IsObject() || !object.HasKey(key) {
		return nil
	}
	found, err := object.GetObjectItem(key)
	if err != nil {
		return nil
	}
	return found
}

// text is the string at key of object, "" when it is none.
func text(object *serializabledeps.SerializableObject, key string) string {
	found := item(object, key)
	if found == nil || !found.IsString() {
		return ""
	}
	value, _ := found.GetString()
	return value
}

// number is the whole number at key of object, 0 when it is none.
func number(object *serializabledeps.SerializableObject, key string) int64 {
	found := item(object, key)
	if found == nil || !(found.IsInt() || found.IsFloat()) {
		return 0
	}
	value, err := found.GetInt()
	if err != nil || value < 0 {
		return 0
	}
	return value
}

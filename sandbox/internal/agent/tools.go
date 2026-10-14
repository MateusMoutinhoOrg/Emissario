package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The model is given three tools: list_instructions and read_instruction read
// the files under instructions/ of the agent source, and call_url calls one
// of the urls a root declared, by its name. What a tool answers is a JSON
// object handed back to the model as data — an error included — and never
// fails the request.

const (
	// ToolListInstructions lists the files under instructions/.
	ToolListInstructions = "list_instructions"
	// ToolReadInstruction reads one of them.
	ToolReadInstruction = "read_instruction"
	// ToolCallUrl calls one of the urls, by name.
	ToolCallUrl = "call_url"
)

// toolsJSON is the tools array offered to the model: call_url only when a url
// is declared, its names an enum and each description in its own.
func toolsJSON(sandbox *api.Sandbox, urls []agent_db.AgentUrlRecord) *serializabledeps.SerializableObject {
	json := sandbox.Deps.SerializableDeps
	tools := json.CreateArray()

	tools.AddItemToArray(toolJSON(sandbox, ToolListInstructions,
		"Lists the path of every file of the instructions you can read with read_instruction.",
		json.CreateObject(), nil))

	path := json.CreateObject()
	path.AddItemToObject("type", "string")
	path.AddItemToObject("description", "the path of the file, as list_instructions answers it, such as faq/prices.md")
	readProperties := json.CreateObject()
	readProperties.AddItemToObject("path", path)
	tools.AddItemToArray(toolJSON(sandbox, ToolReadInstruction,
		"Reads one file of the instructions, at most 64 KiB of it.",
		readProperties, []string{"path"}))

	if len(urls) == 0 {
		return tools
	}
	names := json.CreateArray()
	description := "Calls one of the endpoints below by its name and answers its status and body (at most 32 KiB). Send query for a GET endpoint, body for a POST one.\n"
	for _, url := range urls {
		names.AddItemToArray(url.Name)
		description += "- " + url.Name + " (" + url.Method + "): " + url.Description + "\n"
	}
	name := json.CreateObject()
	name.AddItemToObject("type", "string")
	name.AddItemToObject("enum", names)
	name.AddItemToObject("description", "the name of the endpoint")
	queryValues := json.CreateObject()
	queryValues.AddItemToObject("type", "string")
	query := json.CreateObject()
	query.AddItemToObject("type", "object")
	query.AddItemToObject("additionalProperties", queryValues)
	query.AddItemToObject("description", "the query string parameters of a GET endpoint, name to value")
	body := json.CreateObject()
	body.AddItemToObject("type", "object")
	body.AddItemToObject("description", "the JSON body of a POST endpoint")
	callProperties := json.CreateObject()
	callProperties.AddItemToObject("name", name)
	callProperties.AddItemToObject("query", query)
	callProperties.AddItemToObject("body", body)
	tools.AddItemToArray(toolJSON(sandbox, ToolCallUrl, sandbox.Deps.StringsDeps.TrimSpace(description), callProperties, []string{"name"}))
	return tools
}

// toolJSON is one function tool.
func toolJSON(sandbox *api.Sandbox, name string, description string, properties *serializabledeps.SerializableObject, required []string) *serializabledeps.SerializableObject {
	json := sandbox.Deps.SerializableDeps
	parameters := json.CreateObject()
	parameters.AddItemToObject("type", "object")
	parameters.AddItemToObject("properties", properties)
	if len(required) > 0 {
		list := json.CreateArray()
		for _, key := range required {
			list.AddItemToArray(key)
		}
		parameters.AddItemToObject("required", list)
	}
	function := json.CreateObject()
	function.AddItemToObject("name", name)
	function.AddItemToObject("description", description)
	function.AddItemToObject("parameters", parameters)
	tool := json.CreateObject()
	tool.AddItemToObject("type", "function")
	tool.AddItemToObject("function", function)
	return tool
}

// runTool runs call and answers what it gives back to the model, a JSON
// object; deadline bounds a call_url.
func runTool(sandbox *api.Sandbox, call ToolCall, urls []agent_db.AgentUrlRecord, deadline int64) string {
	json := sandbox.Deps.SerializableDeps
	arguments := json.CreateObject()
	if trimmed := sandbox.Deps.StringsDeps.TrimSpace(call.Arguments); trimmed != "" {
		parsed, err := json.ParseJson(trimmed)
		if err != nil || !parsed.IsObject() {
			return toolError(sandbox, "the arguments are not a JSON object")
		}
		arguments = parsed
	}
	var answer *serializabledeps.SerializableObject
	switch call.Name {
	case ToolListInstructions:
		answer = listInstructions(sandbox)
	case ToolReadInstruction:
		answer = readInstruction(sandbox, text(arguments, "path"))
	case ToolCallUrl:
		answer = callUrl(sandbox, arguments, urls, deadline)
	default:
		return toolError(sandbox, "there is no tool named "+cut(call.Name, 64))
	}
	return json.SerializeToJson(answer)
}

// toolError is the answer of a tool that could not do what it was asked.
func toolError(sandbox *api.Sandbox, message string) string {
	return sandbox.Deps.SerializableDeps.SerializeToJson(errorJSON(sandbox, message))
}

// errorJSON is {"error": message}.
func errorJSON(sandbox *api.Sandbox, message string) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("error", message)
	return document
}

// InstructionFiles is the path of every file under instructions/, relative to
// it and slash-separated, sorted — dotfiles and the files of dot-folders left
// out — at most MaxInstructionFiles of them, and whether there were more.
func InstructionFiles(sandbox *api.Sandbox) ([]string, bool) {
	io := sandbox.Deps.IoDeps
	strings := sandbox.Deps.StringsDeps
	dir := InstructionsDir(sandbox)
	if !io.IsDir(dir) {
		return []string{}, false
	}
	files := []string{}
	for _, path := range io.ListFilesRecursively(dir) {
		relative := strings.TrimPrefix(strings.ReplaceAll(strings.TrimPrefix(path, dir), "\\", "/"), "/")
		safe, ok := sandbox.Deps.OpinionatedAgnosFront.SafePath(relative)
		if !ok || safe == "" || safe != relative {
			continue
		}
		files = append(files, relative)
	}
	sandbox.Deps.SortDeps.Strings(files)
	if len(files) > MaxInstructionFiles {
		return files[:MaxInstructionFiles], true
	}
	return files, false
}

// listInstructions answers {"files": [...]}, and "truncated" when there are
// more than MaxInstructionFiles.
func listInstructions(sandbox *api.Sandbox) *serializabledeps.SerializableObject {
	files, truncated := InstructionFiles(sandbox)
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, file := range files {
		list.AddItemToArray(file)
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("files", list)
	if truncated {
		document.AddItemToObject("truncated", true)
	}
	return document
}

// readInstruction answers {"path", "content", "truncated"?} for the file at
// path under instructions/, or {"error"}: a path that leaves the folder, names
// a dotfile, is no file or is not text.
func readInstruction(sandbox *api.Sandbox, path string) *serializabledeps.SerializableObject {
	content, truncated, message := ReadInstruction(sandbox, path)
	if message != "" {
		return errorJSON(sandbox, message)
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("path", path)
	document.AddItemToObject("content", content)
	if truncated {
		document.AddItemToObject("truncated", true)
	}
	return document
}

// ReadInstruction is the text of the file at path under instructions/ —
// at most MaxInstructionBytes of it, and whether it was cut — or why it
// cannot be read. The path goes through SafePath first, so nothing outside the
// folder is read: no .. segment, no dotfile.
func ReadInstruction(sandbox *api.Sandbox, path string) (string, bool, string) {
	io := sandbox.Deps.IoDeps
	safe, ok := sandbox.Deps.OpinionatedAgnosFront.SafePath(path)
	if !ok || safe == "" {
		return "", false, "no such file: use a path list_instructions answers"
	}
	file := io.Join(append([]string{InstructionsDir(sandbox)}, sandbox.Deps.StringsDeps.Split(safe, "/")...)...)
	if !io.IsFile(file) {
		return "", false, "no such file: use a path list_instructions answers"
	}
	content, err := io.ReadFile(file)
	if err != nil {
		return "", false, "the file could not be read"
	}
	truncated := len(content) > MaxInstructionBytes
	text := cut(string(content), MaxInstructionBytes)
	if !isText([]byte(text)) {
		return "", false, "the file is not text"
	}
	return text, truncated, ""
}

// callUrl calls the url arguments names and answers {"status", "body",
// "truncated"?}, or {"error"}: no url of that name, a query or a body that is
// not one, a call that failed, or an answer that is not text. The headers of
// the url are added here, never shown to the model; no redirect is followed.
func callUrl(sandbox *api.Sandbox, arguments *serializabledeps.SerializableObject, urls []agent_db.AgentUrlRecord, deadline int64) *serializabledeps.SerializableObject {
	name := text(arguments, "name")
	var found *agent_db.AgentUrlRecord
	for index := range urls {
		if urls[index].Name == name {
			found = &urls[index]
		}
	}
	if found == nil {
		return errorJSON(sandbox, "there is no endpoint named "+cut(name, 64))
	}
	address := found.Url
	var body []byte
	if found.Method == "GET" {
		query, message := queryOf(sandbox, item(arguments, "query"))
		if message != "" {
			return errorJSON(sandbox, message)
		}
		if query != "" {
			separator := "?"
			if sandbox.Deps.StringsDeps.Contains(address, "?") {
				separator = "&"
			}
			address += separator + query
		}
	} else if sent := item(arguments, "body"); sent != nil {
		if !sent.IsObject() && !sent.IsArray() {
			return errorJSON(sandbox, "body is a JSON object")
		}
		body = []byte(sandbox.Deps.SerializableDeps.SerializeToJson(sent))
	}

	timeout := int64(UrlTimeoutMilliseconds)
	if remaining := deadline - nowMilliseconds(sandbox); remaining < timeout {
		timeout = remaining
	}
	if timeout <= 0 {
		return errorJSON(sandbox, "no time is left to call the endpoint")
	}
	request := sandbox.Deps.RequestDeps.NewRequest(address)
	request.SetMethod(found.Method)
	request.SetTimeout(timeout)
	request.AddHeader("Accept", "application/json, text/*")
	headers, _ := ParseHeaders(sandbox, found.Headers)
	for _, header := range headers {
		request.AddHeader(header.Name, header.Value)
	}
	if found.Method == "POST" {
		request.AddHeader("Content-Type", "application/json")
		if body == nil {
			body = []byte("{}")
		}
		request.SetBody(body)
	}
	response, err := request.Fetch()
	if err != nil {
		return errorJSON(sandbox, "the endpoint could not be reached or did not answer in time")
	}
	defer response.Close()
	contentType := sandbox.Deps.StringsDeps.ToLower(response.GetHeader("Content-Type"))
	if contentType != "" && !textual(sandbox, contentType) {
		return errorJSON(sandbox, "the endpoint answered "+cut(contentType, 64)+", which is not text")
	}
	content, err := response.ReadBody(MaxUrlResponseBytes + 1)
	if err != nil {
		return errorJSON(sandbox, "the answer of the endpoint could not be read")
	}
	truncated := len(content) > MaxUrlResponseBytes
	answer := cut(string(content), MaxUrlResponseBytes)
	if !isText([]byte(answer)) {
		return errorJSON(sandbox, "the endpoint answered something that is not text")
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("status", response.GetStatusCode())
	document.AddItemToObject("body", answer)
	if truncated {
		document.AddItemToObject("truncated", true)
	}
	return document
}

// textual tells whether contentType is text or JSON.
func textual(sandbox *api.Sandbox, contentType string) bool {
	strings := sandbox.Deps.StringsDeps
	media := strings.TrimSpace(strings.Split(contentType, ";")[0])
	return strings.HasPrefix(media, "text/") || media == "application/json" || strings.HasSuffix(media, "+json") ||
		media == "application/xml" || strings.HasSuffix(media, "+xml")
}

// queryOf is the query string of the object query — name=value pairs sorted
// by name, each escaped — "" for none, or why it is not one: a value that is
// an object or an array.
func queryOf(sandbox *api.Sandbox, query *serializabledeps.SerializableObject) (string, string) {
	if query == nil || query.IsNull() {
		return "", ""
	}
	if !query.IsObject() {
		return "", "query is an object of names to values"
	}
	keys, _ := query.GetKeys()
	sandbox.Deps.SortDeps.Strings(keys)
	pairs := []string{}
	for _, key := range keys {
		value, _ := query.GetObjectItem(key)
		spelled := ""
		switch {
		case value.IsString():
			spelled, _ = value.GetString()
		case value.IsInt():
			number, _ := value.GetInt()
			spelled = sandbox.Deps.StringsDeps.FormatInt(number, 10)
		case value.IsFloat():
			number, _ := value.GetFloat()
			spelled = sandbox.Deps.StringsDeps.FormatFloat(number, 'f', -1, 64)
		case value.IsBool():
			flag, _ := value.GetBool()
			spelled = "false"
			if flag {
				spelled = "true"
			}
		case value.IsNull():
		default:
			return "", "the query value of " + cut(key, 64) + " is not a text, a number or a boolean"
		}
		pairs = append(pairs, visualization.QueryEscape(sandbox, key)+"="+visualization.QueryEscape(sandbox, spelled))
	}
	return sandbox.Deps.StringsDeps.Join(pairs, "&"), ""
}

package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// NoticeJSON is one notice: its id, its title and its text.
func NoticeJSON(sandbox *api.Sandbox, notice privacy.Notice) *serializabledeps.SerializableObject {
	object := sandbox.Deps.SerializableDeps.CreateObject()
	object.AddItemToObject("id", notice.Id)
	object.AddItemToObject("title", notice.Title)
	object.AddItemToObject("text", notice.Text)
	return object
}

// AcknowledgementRequiredJSON is the body of the 409 a request answers while
// it has not acknowledged every notice it needs:
// {"error":"acknowledgement-required","field":"acknowledge","notice":{…},
// "notices":[…]}, notice the first of notices. The request is sent again
// with every id in "acknowledge".
func AcknowledgementRequiredJSON(sandbox *api.Sandbox, notices []privacy.Notice) *serializabledeps.SerializableObject {
	listed := sandbox.Deps.SerializableDeps.CreateArray()
	for _, notice := range notices {
		listed.AddItemToArray(NoticeJSON(sandbox, notice))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("error", privacy.AcknowledgementRequired)
	document.AddItemToObject("field", privacy.AcknowledgeField)
	if len(notices) > 0 {
		document.AddItemToObject("notice", NoticeJSON(sandbox, notices[0]))
	}
	document.AddItemToObject("notices", listed)
	return document
}

// FailPrivacy answers refusal: the 409 of AcknowledgementRequiredJSON when it
// lists notices — loaded with context — and {"error","field"} under its
// status otherwise.
func FailPrivacy(sandbox *api.Sandbox, response *serverdeps.Response, refusal *privacy.Refusal, context privacy.NoticeContext) error {
	if len(refusal.Notices) == 0 {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	notices, err := privacy.LoadNotices(sandbox, refusal.Notices, context)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusConflict, AcknowledgementRequiredJSON(sandbox, notices))
}

// RetentionJSON is a retention as JSON spells it: the number of days, or null
// to keep the data until a root deletes it.
func RetentionJSON(sandbox *api.Sandbox, days int64) any {
	if days == privacy.Forever {
		return sandbox.Deps.SerializableDeps.CreateNull()
	}
	return days
}

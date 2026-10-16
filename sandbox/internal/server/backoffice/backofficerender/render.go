package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficetranslation"
)

// RenderHTML renders the template at path with vars and answers it under status.
// Templates run on html/template, so every value is escaped for the context
// it lands in — text, an attribute, a url — whether or not the template
// spells {{html ...}} too; a field added without it is still safe.
//
// The page is in the language the Content-Language header of response
// names, which the backoffice-language middleware set from the request —
// backofficetranslation.DefaultLang when none did — and its text is the
// template's {{t}}, {{th}} and {{tr}} (embeddeps.Translator).
func RenderHTML(sandbox *api.Sandbox, response *serverdeps.Response, status int, path string, vars any) error {
	translator, err := backofficetranslation.Translator(sandbox, response.GetHeader("Content-Language"))
	if err != nil {
		return err
	}
	content, err := sandbox.Deps.EmbedDeps.RenderTranslatedHTMLTemplate(path, vars, translator)
	if err != nil {
		return err
	}
	response.SetHeader("Content-Type", "text/html; charset=utf-8")
	response.SetHeader("Content-Language", translator.Lang)
	response.SetHeader("Cache-Control", "no-store")
	response.SetStatus(status)
	return response.Write(content)
}

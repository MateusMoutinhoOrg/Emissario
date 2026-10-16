package backoffice_language

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficetranslation"
)

// Handle runs in front of every ANY /admin/*, on the lowest rung of the
// chain but backoffice-maintenance's: it is a middleware, and answers
// nothing.
//
// It picks the language of the page (backofficetranslation.Pick): ?lang= when
// it names a language of the backoffice — the cookie then keeps it for the
// next pages — else the language the cookie kept, else English. It sets it as
// the Content-Language of the response, which every page renders in
// (backofficerender.RenderHTML), and declines, so the route after it runs. A
// ?lang= naming no language is ignored.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	lang, chosen, err := backofficetranslation.Pick(sandbox, input.Lang, input.LangCookie)
	if err != nil {
		return err
	}
	if chosen && lang != input.LangCookie {
		response.AddHeader("Set-Cookie", backofficetranslation.Cookie(sandbox, lang))
	}
	response.SetHeader("Content-Language", lang)
	return nil
}

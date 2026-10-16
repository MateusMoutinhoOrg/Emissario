package embeddeps

// This package is the sandbox's *copy* of the api an embedded-asset library
// exposes — the same mechanic as argvdeps and dbdeps, for the same reason:
// reading a file is an OS-bound effect, and compiling one into a binary needs
// the `embed` directive, so neither may appear inside the sandbox. The
// contract is restated here, and the adapter — which lives outside the
// sandbox — is what fills it.
//
// The library asks for an asset the way it would ask a filesystem, by path;
// where those bytes come from is the adapter's decision. The standard adapter
// serves them out of the assets compiled into the binary, so an installed
// `agnos` carries them wherever it runs; another adapter could read a
// directory on disk, an archive, or a network store without the sandbox
// noticing.
//
// The tracker in sandbox/ never reads an asset — its display text is small
// and fixed, so it lives in sandbox/config as compile-time constants. This
// contract is filled anyway, as a standing capability of the template: see
// the Deps.EmbedDeps field for why.

// Contract is the embedded-asset library injected whole as the Deps.EmbedDeps
// field. It is read-only by design: assets ship with the program, and nothing
// in the library ever writes one back.
//
// Every path is slash-separated and relative to the root of the asset tree
// the adapter serves — "report.tmpl", "templates/invoice.tmpl" — never an
// absolute path and never a path reaching outside that root, so the same call
// means the same asset whatever the adapter is backed by.
type Contract struct {
	// ReadFile returns the whole content of one asset. The error reports an
	// asset that does not exist or could not be read; callers inside the
	// sandbox report it rather than assuming the bytes are there, because a
	// missing asset is a packaging mistake, not a user mistake.
	ReadFile func(path string) ([]byte, error)

	// ListFiles returns the names of the assets directly inside the given
	// directory, in lexical order, relative to that directory. Nested
	// directories are not descended into and are not reported. The root
	// itself is addressed as ".".
	ListFiles func(path string) ([]string, error)

	// ListFilesRecursively returns every asset at or below the given
	// directory, in lexical order, as slash-separated paths relative to that
	// directory — "templates/invoice.tmpl" and not just "invoice.tmpl".
	// Directories are never reported, only the files inside them.
	ListFilesRecursively func(path string) ([]string, error)

	// RenderTemplate reads the template from path, renders it using the given variables,
	// and returns the resulting byte slice.
	RenderTemplate func(path string, vars interface{}) ([]byte, error)

	// RenderHTMLTemplate is RenderTemplate for an HTML page: every value is
	// escaped for the context it lands in — text, an attribute, a url, a
	// script — so a value nobody escaped by hand can never become markup.
	RenderHTMLTemplate func(path string, vars interface{}) ([]byte, error)

	// RenderTranslatedHTMLTemplate is RenderHTMLTemplate for a page whose
	// text is translated: the template reaches translator through the
	// functions Translator lists, every value still escaped for its context.
	RenderTranslatedHTMLTemplate func(path string, vars interface{}, translator Translator) ([]byte, error)
}

// Translator is the language a translated template is rendered in, and the
// functions it calls for its text. The adapter turns every argument a
// template passes into its text form before calling one, and registers them
// under these names:
//
//	{{lang}}                  Lang, for <html lang>
//	{{languages}}             Languages, for a language picker
//	{{t "key" .A .B}}         Text: the text of key, its {0}, {1}… filled, escaped like any value
//	{{th "key" .A .B}}        Markup: the same, its markup kept — only in text, never in an attribute
//	{{tr .Message}}           Message: the translation of a text built in English, or the text itself
type Translator struct {
	// Lang is the code of the language, "en" or "pt-BR".
	Lang string
	// Languages are every language a page can be shown in, in the order a
	// picker lists them.
	Languages []Language
	// Text returns the text of key with {0}, {1}… replaced by args.
	Text func(key string, args []string) string
	// Markup returns the markup of key with {0}, {1}… replaced by args,
	// each one escaped: the markup comes from the translation file, never
	// from a value.
	Markup func(key string, args []string) string
	// Message returns the translation of text, a message written in
	// English, or text itself when none is known.
	Message func(text string) string
}

// Language is one language a translated page can be shown in.
type Language struct {
	// Code is its code, "en" or "pt-BR".
	Code string
	// Name is its name in itself, "English" or "Português (Brasil)".
	Name string
	// Current tells whether it is the language of the page.
	Current bool
}

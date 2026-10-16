# `deps.EmbedDeps`

`sandbox/deps/embeddeps`

## `Contract`

Contract is the embedded-asset library injected whole as the Deps.EmbedDeps field. It is read-only by design: assets ship with the program, and nothing in the library ever writes one back. Every path is slash-separated and relative to the root of the asset tree the adapter serves — "report.tmpl", "templates/invoice.tmpl" — never an absolute path and never a path reaching outside that root, so the same call means the same asset whatever the adapter is backed by.

| Field | Type | Description |
| --- | --- | --- |
| `ReadFile` | `func(path string) ([]byte, error)` | ReadFile returns the whole content of one asset. The error reports an asset that does not exist or could not be read; callers inside the sandbox report it rather than assuming the bytes are there, because a missing asset is a packaging mistake, not a user mistake. |
| `ListFiles` | `func(path string) ([]string, error)` | ListFiles returns the names of the assets directly inside the given directory, in lexical order, relative to that directory. Nested directories are not descended into and are not reported. The root itself is addressed as ".". |
| `ListFilesRecursively` | `func(path string) ([]string, error)` | ListFilesRecursively returns every asset at or below the given directory, in lexical order, as slash-separated paths relative to that directory — "templates/invoice.tmpl" and not just "invoice.tmpl". Directories are never reported, only the files inside them. |
| `RenderTemplate` | `func(path string, vars interface{}) ([]byte, error)` | RenderTemplate reads the template from path, renders it using the given variables, and returns the resulting byte slice. |
| `RenderHTMLTemplate` | `func(path string, vars interface{}) ([]byte, error)` | RenderHTMLTemplate is RenderTemplate for an HTML page: every value is escaped for the context it lands in — text, an attribute, a url, a script — so a value nobody escaped by hand can never become markup. |
| `RenderTranslatedHTMLTemplate` | `func(path string, vars interface{}, translator Translator) ([]byte, error)` | RenderTranslatedHTMLTemplate is RenderHTMLTemplate for a page whose text is translated: the template reaches translator through the functions Translator lists, every value still escaped for its context. |

## `Translator`

Translator is the language a translated template is rendered in, and the functions it calls for its text. The adapter turns every argument a template passes into its text form before calling one, and registers them under these names: {{lang}} Lang, for <html lang> {{languages}} Languages, for a language picker {{t "key" .A .B}} Text: the text of key, its {0}, {1}… filled, escaped like any value {{th "key" .A .B}} Markup: the same, its markup kept — only in text, never in an attribute {{tr .Message}} Message: the translation of a text built in English, or the text itself

| Field | Type | Description |
| --- | --- | --- |
| `Lang` | `string` | Lang is the code of the language, "en" or "pt-BR". |
| `Languages` | `[]Language` | Languages are every language a page can be shown in, in the order a picker lists them. |
| `Text` | `func(key string, args []string) string` | Text returns the text of key with {0}, {1}… replaced by args. |
| `Markup` | `func(key string, args []string) string` | Markup returns the markup of key with {0}, {1}… replaced by args, each one escaped: the markup comes from the translation file, never from a value. |
| `Message` | `func(text string) string` | Message returns the translation of text, a message written in English, or text itself when none is known. |

## `Language`

Language is one language a translated page can be shown in.

| Field | Type | Description |
| --- | --- | --- |
| `Code` | `string` | Code is its code, "en" or "pt-BR". |
| `Name` | `string` | Name is its name in itself, "English" or "Português (Brasil)". |
| `Current` | `bool` | Current tells whether it is the language of the page. |

[every contract](doc.md)

package backofficetranslation

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/embeddeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
)

// The languages of the backoffice pages. Each one is a file of the assets,
// translations/<code>.json:
//
//	{
//	  "language": "Português (Brasil)",
//	  "text": { "common": { "nav": { "home": "Início" } } },
//	  "messages": { "Invalid username or password.": "Usuário ou senha inválidos." }
//	}
//
// language is its name in itself, shown by the picker. text holds the text
// of the templates, nested, read as dotted keys: {{t "common.nav.home"}}. A
// key missing from a language is taken from DefaultLang, and a key missing
// from both shows as the key itself. messages translates the texts the
// server builds in English — an error, a notice — by their English text,
// {{tr .Error}}; a text it does not list stays in English. Adding a language
// is adding its file: nothing else changes.
//
// A message whose English text holds {0}, {1}… is a pattern: each one stands
// for any text the server put there — a count, a name — and the translation
// puts it back where its own {0}, {1}… are:
//
//	"{0} contact requests deleted.": "{0} pedidos de contato excluídos."
//
// A text the messages list as it is wins over a pattern, and of the patterns
// it matches the one with the most text of its own wins. What a placeholder
// matched is translated in turn, so "{0} and {1}." with "{0} views": "{0}
// visualizações" reads "3 views and 1 IP." whole, and "Preset {0} applied."
// shows the translated name of a built-in preset.
//
// Only the backoffice pages are translated: the JSON of /api/admin, the
// command line and the docs stay in English.

// Dir is the directory of the assets the translation files live in.
const Dir = "translations"

// DefaultLang is the language of a page when none was chosen, and the one a
// key missing from the chosen language is taken from.
const DefaultLang = "en"

// QueryKey is the query parameter a page is asked in another language
// with, ?lang=pt-BR, which the picker of every page sets.
const QueryKey = "lang"

// CookieName is the cookie the language chosen is kept in, so the next
// pages are shown in it.
const CookieName = "backoffice_language"

// CookiePath is the path the language cookie is sent to: the pages alone.
const CookiePath = "/admin"

// CookieSeconds is how long the browser keeps the language chosen: a year.
const CookieSeconds = 365 * 24 * 60 * 60

// Catalog is one language: its code, its name and its texts.
type Catalog struct {
	// Code is the name of its file without .json, "pt-BR".
	Code string
	// Name is its name in itself, "Português (Brasil)".
	Name string
	// Text is the text of every key, "common.nav.home" → "Início".
	Text map[string]string
	// Messages is the translation of every server message it lists, by its
	// English text.
	Messages map[string]string
	// Patterns are the messages holding {0}, {1}…, the one with the most
	// text of its own first.
	Patterns []Pattern
}

// Pattern is a message whose English text holds {0}, {1}…, each standing for
// any text.
type Pattern struct {
	// Parts is the English text cut at its placeholders: the text before the
	// first, between each two and after the last.
	Parts []string
	// Slots is the number of each placeholder, in the order they appear.
	Slots []int
	// Fixed is how long the English text is without its placeholders.
	Fixed int
	// Translation is the translation, holding the same placeholders.
	Translation string
}

// catalogs is every language, DefaultLang first, read once from the assets
// and never written again; catalogsLock guards it and catalogsRead.
var catalogs []Catalog

var catalogsRead bool

var catalogsLock = make(chan struct{}, 1)

// Catalogs is every language of the backoffice, DefaultLang first and then
// in the order of their codes.
func Catalogs(sandbox *api.Sandbox) ([]Catalog, error) {
	catalogsLock <- struct{}{}
	defer func() { <-catalogsLock }()
	if catalogsRead {
		return catalogs, nil
	}
	read, err := readCatalogs(sandbox)
	if err != nil {
		return nil, err
	}
	catalogs, catalogsRead = read, true
	return catalogs, nil
}

// readCatalogs reads every translations/<code>.json of the assets.
func readCatalogs(sandbox *api.Sandbox) ([]Catalog, error) {
	names, err := sandbox.Deps.EmbedDeps.ListFiles(Dir)
	if err != nil {
		return nil, err
	}
	first := []Catalog{}
	rest := []Catalog{}
	for _, name := range names {
		if !sandbox.Deps.StringsDeps.HasSuffix(name, ".json") {
			continue
		}
		catalog, err := readCatalog(sandbox, sandbox.Deps.StringsDeps.TrimSuffix(name, ".json"))
		if err != nil {
			return nil, err
		}
		if catalog.Code == DefaultLang {
			first = append(first, catalog)
			continue
		}
		rest = append(rest, catalog)
	}
	if len(first) == 0 {
		return nil, sandbox.Deps.StdDeps.Errorf("%s/%s.json, the default language of the backoffice, is missing", Dir, DefaultLang)
	}
	return append(first, rest...), nil
}

// readCatalog reads translations/<code>.json.
func readCatalog(sandbox *api.Sandbox, code string) (Catalog, error) {
	path := Dir + "/" + code + ".json"
	content, err := sandbox.Deps.EmbedDeps.ReadFile(path)
	if err != nil {
		return Catalog{}, err
	}
	root, err := sandbox.Deps.SerializableDeps.ParseJson(string(content))
	if err != nil {
		return Catalog{}, sandbox.Deps.StdDeps.Errorf("%s: %s", path, err.Error())
	}
	if !root.IsObject() {
		return Catalog{}, sandbox.Deps.StdDeps.Errorf("%s: not an object", path)
	}
	catalog := Catalog{Code: code, Name: code, Text: map[string]string{}, Messages: map[string]string{}}
	if root.HasKey("language") {
		name, err := stringOf(sandbox, root, "language")
		if err != nil {
			return Catalog{}, sandbox.Deps.StdDeps.Errorf("%s: language: %s", path, err.Error())
		}
		catalog.Name = name
	}
	if root.HasKey("text") {
		text, err := root.GetObjectItem("text")
		if err != nil {
			return Catalog{}, err
		}
		if err := flatten(sandbox, text, "", catalog.Text); err != nil {
			return Catalog{}, sandbox.Deps.StdDeps.Errorf("%s: text%s", path, err.Error())
		}
	}
	if root.HasKey("messages") {
		messages, err := root.GetObjectItem("messages")
		if err != nil {
			return Catalog{}, err
		}
		if !messages.IsObject() {
			return Catalog{}, sandbox.Deps.StdDeps.Errorf("%s: messages: not an object", path)
		}
		keys, err := messages.GetKeys()
		if err != nil {
			return Catalog{}, err
		}
		for _, key := range keys {
			value, err := stringOf(sandbox, messages, key)
			if err != nil {
				return Catalog{}, sandbox.Deps.StdDeps.Errorf("%s: messages: %s: %s", path, key, err.Error())
			}
			catalog.Messages[key] = value
			if pattern, ok := parsePattern(sandbox, key, value); ok {
				catalog.Patterns = insertPattern(catalog.Patterns, pattern)
			}
		}
	}
	return catalog, nil
}

// parsePattern is the pattern of the message text translated as translation,
// false when text holds no {0}, {1}… or nothing else: a pattern that is only
// placeholders would match every text.
func parsePattern(sandbox *api.Sandbox, text string, translation string) (Pattern, bool) {
	pattern := Pattern{Translation: translation}
	part := ""
	for i := 0; i < len(text); i++ {
		if text[i] == '{' {
			end := i + 1
			for end < len(text) && text[end] >= '0' && text[end] <= '9' {
				end++
			}
			if end > i+1 && end < len(text) && text[end] == '}' {
				slot, err := sandbox.Deps.StringsDeps.Atoi(text[i+1 : end])
				if err == nil {
					pattern.Parts = append(pattern.Parts, part)
					pattern.Slots = append(pattern.Slots, slot)
					pattern.Fixed += len(part)
					part = ""
					i = end
					continue
				}
			}
		}
		part += text[i : i+1]
	}
	if len(pattern.Slots) == 0 {
		return Pattern{}, false
	}
	pattern.Parts = append(pattern.Parts, part)
	pattern.Fixed += len(part)
	if pattern.Fixed == 0 {
		return Pattern{}, false
	}
	return pattern, true
}

// insertPattern is patterns with pattern put after every one with as much
// text of its own or more, so the most specific is tried first.
func insertPattern(patterns []Pattern, pattern Pattern) []Pattern {
	at := len(patterns)
	for at > 0 && patterns[at-1].Fixed < pattern.Fixed {
		at--
	}
	patterns = append(patterns, Pattern{})
	copy(patterns[at+1:], patterns[at:])
	patterns[at] = pattern
	return patterns
}

// translateMessage is the translation of message in catalog: the one it
// lists for message as it is, else that of the first pattern message matches,
// what each placeholder matched translated the same way, else message itself.
// A placeholder matches less than the whole message, so this ends.
func translateMessage(sandbox *api.Sandbox, catalog Catalog, message string) string {
	if value, ok := catalog.Messages[message]; ok {
		return value
	}
	for _, pattern := range catalog.Patterns {
		captured, ok := matchPattern(sandbox, pattern.Parts, message)
		if !ok {
			continue
		}
		args := []string{}
		for i, slot := range pattern.Slots {
			for len(args) <= slot {
				args = append(args, "{"+sandbox.Deps.StringsDeps.FormatInt(int64(len(args)), 10)+"}")
			}
			args[slot] = translateMessage(sandbox, catalog, captured[i])
		}
		return fill(sandbox, pattern.Translation, args)
	}
	return message
}

// matchPattern is what each placeholder between parts matched in text, in
// the order they appear, each at least one byte long; false when text does
// not read as parts with something between each two.
func matchPattern(sandbox *api.Sandbox, parts []string, text string) ([]string, bool) {
	if !sandbox.Deps.StringsDeps.HasPrefix(text, parts[0]) {
		return nil, false
	}
	return matchPlaceholders(sandbox, parts[1:], text[len(parts[0]):])
}

// matchPlaceholders matches text against a placeholder followed by parts[0],
// another followed by parts[1]…, trying every length of each placeholder.
func matchPlaceholders(sandbox *api.Sandbox, parts []string, text string) ([]string, bool) {
	if len(parts) == 1 {
		end := len(text) - len(parts[0])
		if end < 1 || !sandbox.Deps.StringsDeps.HasSuffix(text, parts[0]) {
			return nil, false
		}
		return []string{text[:end]}, true
	}
	for end := 1; end <= len(text); end++ {
		if !sandbox.Deps.StringsDeps.HasPrefix(text[end:], parts[0]) {
			continue
		}
		rest, ok := matchPlaceholders(sandbox, parts[1:], text[end+len(parts[0]):])
		if ok {
			return append([]string{text[:end]}, rest...), true
		}
	}
	return nil, false
}

// flatten puts every string under node into into, each under its dotted key.
func flatten(sandbox *api.Sandbox, node *serializabledeps.SerializableObject, prefix string, into map[string]string) error {
	if !node.IsObject() {
		return sandbox.Deps.StdDeps.Errorf("%s: not an object", prefix)
	}
	keys, err := node.GetKeys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		child, err := node.GetObjectItem(key)
		if err != nil {
			return err
		}
		dotted := key
		if prefix != "" {
			dotted = prefix + "." + key
		}
		if child.IsObject() {
			if err := flatten(sandbox, child, dotted, into); err != nil {
				return err
			}
			continue
		}
		value, err := child.GetString()
		if err != nil || !child.IsString() {
			return sandbox.Deps.StdDeps.Errorf(".%s: not a string", dotted)
		}
		into[dotted] = value
	}
	return nil
}

// stringOf is the string under key of object.
func stringOf(sandbox *api.Sandbox, object *serializabledeps.SerializableObject, key string) (string, error) {
	item, err := object.GetObjectItem(key)
	if err != nil {
		return "", err
	}
	if !item.IsString() {
		return "", sandbox.Deps.StdDeps.Errorf("not a string")
	}
	return item.GetString()
}

// Supported tells whether code names a language of the backoffice, exactly
// as its file is named.
func Supported(sandbox *api.Sandbox, code string) (bool, error) {
	all, err := Catalogs(sandbox)
	if err != nil {
		return false, err
	}
	for _, catalog := range all {
		if catalog.Code == code {
			return true, nil
		}
	}
	return false, nil
}

// Pick is the language of a page asked with ?lang=query, the language
// cookie holding cookie: the query when it names a language — chosen tells
// so, and the cookie is then to be set — else the cookie when it names one,
// else DefaultLang.
func Pick(sandbox *api.Sandbox, query string, cookie string) (lang string, chosen bool, err error) {
	if query != "" {
		ok, err := Supported(sandbox, query)
		if err != nil {
			return "", false, err
		}
		if ok {
			return query, true, nil
		}
	}
	if cookie != "" {
		ok, err := Supported(sandbox, cookie)
		if err != nil {
			return "", false, err
		}
		if ok {
			return cookie, false, nil
		}
	}
	return DefaultLang, false, nil
}

// Cookie is the Set-Cookie value that keeps lang as the language of the
// next pages. lang is a language Supported, so it holds no character a
// cookie may not.
func Cookie(sandbox *api.Sandbox, lang string) string {
	secure := "; Secure"
	if sandbox.Config.InsecureHttp {
		secure = ""
	}
	return sandbox.Deps.StdDeps.Sprintf("%s=%s; Path=%s; Max-Age=%d; HttpOnly; SameSite=Lax%s", CookieName, lang, CookiePath, CookieSeconds, secure)
}

// Translator is what a page in lang is rendered with: its texts, the default
// language's for a key it lacks. A lang that is no language of the
// backoffice is DefaultLang.
func Translator(sandbox *api.Sandbox, lang string) (embeddeps.Translator, error) {
	all, err := Catalogs(sandbox)
	if err != nil {
		return embeddeps.Translator{}, err
	}
	fallback := all[0]
	current := fallback
	for _, catalog := range all {
		if catalog.Code == lang {
			current = catalog
		}
	}
	languages := make([]embeddeps.Language, len(all))
	for i, catalog := range all {
		languages[i] = embeddeps.Language{Code: catalog.Code, Name: catalog.Name, Current: catalog.Code == current.Code}
	}
	text := func(key string) string {
		if value, ok := current.Text[key]; ok {
			return value
		}
		if value, ok := fallback.Text[key]; ok {
			return value
		}
		return key
	}
	return embeddeps.Translator{
		Lang:      current.Code,
		Languages: languages,
		Text: func(key string, args []string) string {
			return fill(sandbox, text(key), args)
		},
		Markup: func(key string, args []string) string {
			escaped := make([]string, len(args))
			for i, arg := range args {
				escaped[i] = Escape(sandbox, arg)
			}
			return fill(sandbox, text(key), escaped)
		},
		Message: func(message string) string {
			return translateMessage(sandbox, current, message)
		},
	}, nil
}

// fill replaces {0}, {1}… in text by args, in one pass: an arg holding
// {1} is kept as it is. A {n} with no arg n stays in the text.
func fill(sandbox *api.Sandbox, text string, args []string) string {
	if len(args) == 0 {
		return text
	}
	filled := ""
	for i := 0; i < len(text); i++ {
		if text[i] == '{' {
			end := i + 1
			for end < len(text) && text[end] >= '0' && text[end] <= '9' {
				end++
			}
			if end > i+1 && end < len(text) && text[end] == '}' {
				index, err := sandbox.Deps.StringsDeps.Atoi(text[i+1 : end])
				if err == nil && index < len(args) {
					filled += args[index]
					i = end
					continue
				}
			}
		}
		filled += text[i : i+1]
	}
	return filled
}

// Escape is text written as html text or an attribute value: & < > " and '
// replaced by their entities.
func Escape(sandbox *api.Sandbox, text string) string {
	text = sandbox.Deps.StringsDeps.ReplaceAll(text, "&", "&amp;")
	text = sandbox.Deps.StringsDeps.ReplaceAll(text, "<", "&lt;")
	text = sandbox.Deps.StringsDeps.ReplaceAll(text, ">", "&gt;")
	text = sandbox.Deps.StringsDeps.ReplaceAll(text, "\"", "&#34;")
	return sandbox.Deps.StringsDeps.ReplaceAll(text, "'", "&#39;")
}

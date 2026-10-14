package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// The body of a contact request is kept as it was sent, so it is shown from
// its own text: laid out again here, never parsed and serialized back, which
// would sort its keys and round its large numbers. Every body stored was
// checked to be JSON when it arrived.

// Indent is body laid out two spaces per level, one value per line; an empty
// object or array stays {} or []. A text that does not nest as JSON comes back
// as it is.
func Indent(sandbox *api.Sandbox, body string) string {
	return layout(body, true)
}

// Preview is body on one line — ", " and ": " between its values — cut to
// length characters with an ellipsis.
func Preview(sandbox *api.Sandbox, body string, length int) string {
	line := []rune(layout(body, false))
	if len(line) <= length {
		return string(line)
	}
	return string(line[:length]) + "…"
}

// Matches tells whether body holds needle — already lower case — in any case:
// in its text as sent, or, when that spells a character as a \u escape, in
// its text with every escape read.
func Matches(sandbox *api.Sandbox, body string, needle string) bool {
	strings := sandbox.Deps.StringsDeps
	if strings.Contains(strings.ToLower(body), needle) {
		return true
	}
	if !strings.Contains(body, `\u`) {
		return false
	}
	parsed, err := sandbox.Deps.SerializableDeps.ParseJson(body)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(sandbox.Deps.SerializableDeps.SerializeToJson(parsed)), needle)
}

// layout is body without the blanks between its values, laid out on lines of
// one value each when indent is true, on one line otherwise. Strings are
// copied whole, escapes and all; the structural characters are all ASCII, so
// no byte of a multi-byte character is mistaken for one.
func layout(body string, indent bool) string {
	out := make([]byte, 0, len(body)+len(body)/4)
	depth := 0
	inString, escaped := false, false
	newline := func() {
		if !indent {
			return
		}
		out = append(out, '\n')
		for level := 0; level < depth; level++ {
			out = append(out, ' ', ' ')
		}
	}
	for index := 0; index < len(body); index++ {
		char := body[index]
		if inString {
			out = append(out, char)
			switch {
			case escaped:
				escaped = false
			case char == '\\':
				escaped = true
			case char == '"':
				inString = false
			}
			continue
		}
		switch char {
		case ' ', '\t', '\n', '\r':
		case '"':
			inString = true
			out = append(out, char)
		case '{', '[':
			out = append(out, char)
			next := index + 1
			for next < len(body) && (body[next] == ' ' || body[next] == '\t' || body[next] == '\n' || body[next] == '\r') {
				next++
			}
			if next < len(body) && (body[next] == '}' || body[next] == ']') {
				out = append(out, body[next])
				index = next
				continue
			}
			depth++
			newline()
		case '}', ']':
			depth--
			if depth < 0 {
				return body
			}
			newline()
			out = append(out, char)
		case ',':
			out = append(out, char)
			if indent {
				newline()
			} else {
				out = append(out, ' ')
			}
		case ':':
			out = append(out, ':', ' ')
		default:
			out = append(out, char)
		}
	}
	if depth != 0 || inString {
		return body
	}
	return string(out)
}

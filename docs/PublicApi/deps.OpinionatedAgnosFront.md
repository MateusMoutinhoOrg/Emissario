# `deps.OpinionatedAgnosFront`

`sandbox/deps/OpinionatedAgnosFront`

| Constant | Value | Description |
| --- | --- | --- |
| `Root` | `"front"` | Root is the directory of the embedded asset tree every file is read from, as embeddeps spells a path: slash-separated and relative to the root of the assets package, so "assets/front" on disk. |
| `Index` | `"index.html"` | Index is the file a directory is answered with, "/" included. |
| `NotFound` | `"404.html"` | NotFound is the file a path naming no file is answered with, under a 404. front-init writes it once; it is the project's to restyle. |
| `RevalidateCache` | `"no-cache"` | RevalidateCache is the cache header every file is served with: the browser may store the answer but must ask before reusing it, so a new build is picked up on the next load whatever the file is named. |

## `Contract`

Contract is the front lib injected whole as the Deps.OpinionatedAgnosFront field.

| Field | Type | Description |
| --- | --- | --- |
| `Resolve` | `func(embedded embeddeps.Contract, requested string) (string, []byte, bool)` | Resolve reads, through embedded, the file one request path names under Root, trying in order the path itself, the path plus ".html", and the path as a directory holding Index — so /about answers about.html or about/index.html, and / answers index.html. It returns the path relative to Root it read, for ContentTypeOf, and false when the path is unsafe or names nothing, which a caller reads as "not mine". |
| `SafePath` | `func(requested string) (string, bool)` | SafePath turns a request path into the path relative to Root it names, and reports false on anything that could climb out of it: "." and ".." segments, an empty segment, and a backslash or a NUL inside one. A segment starting with "." — a dotfile, .env or .git — is refused too, but for .well-known at the root. A leading and a trailing slash are dropped; "" names Root itself. |
| `ExtensionOf` | `func(path string) string` | ExtensionOf returns the extension of the last segment of a path, dot included, or "" when it has none. |
| `ContentTypeOf` | `func(relative string) string` | ContentTypeOf reads the media type off a file's extension, matched in lower case. An extension nothing claims is served as application/octet-stream rather than guessed at, which also keeps an unknown file from being rendered as html by the browser. |

[every contract](doc.md)

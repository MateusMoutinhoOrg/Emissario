# `ANY /admin/*`

Picks the language of the backoffice pages, from ?lang= or the cookie that keeps it

## Try it

Only what is required:

```bash
curl localhost:3000/admin
```

With every value it reads:

```bash
curl 'localhost:3000/admin?lang=my-lang' \
  -b 'backoffice_language=my-backoffice_language'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `lang` | query | text | no | `my-lang` | the language a page is asked in, pt-BR — kept in the backoffice_language cookie for the next pages |
| `backoffice_language` | cookie | text | no | `my-backoffice_language` | the language chosen before, set by ?lang= |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

---

For developers: `sandbox/internal/routes/backoffice_language/` · Middleware · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

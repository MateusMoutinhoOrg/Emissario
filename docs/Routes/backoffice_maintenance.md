# `ANY /*`

Answers 503 to every request while a restore writes the databases, and to every write while a snapshot reads them

## Try it

Only what is required:

```bash
curl localhost:3000/
```

With every value it reads:

```bash
curl localhost:3000/ \
  -H 'x-request-method: my-x-request-method'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

---

For developers: `sandbox/internal/routes/backoffice_maintenance/` · Middleware · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

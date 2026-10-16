# `POST /admin/root/add-ip-block`

Blocks an ip, a CIDR range, a hashed ip or a country

Root only. value is one ip (203.0.113.7), a CIDR range (203.0.113.0/24), a hashed ip (h: and 32 hex digits, as the records keep it) or a country code (BR); an IPv6 address is blocked by its /64. reason is at most 200 characters; days from 1 to 36500, blank to block until a root removes it. Adding answers 303 to the blocks with ?notice=added; a refused value answers the page under a 400 with the reason, a value blocked already or one matching the ip you are connected from under a 409.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/add-ip-block \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST 'localhost:3000/admin/root/add-ip-block?lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'days=text&reason=text&value=text'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" -X POST localhost:3000/admin/root/add-ip-block -H 'Origin: http://localhost:3000' --data 'value=203.0.113.7&reason=spam&days=30'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `lang` | query string | text | no | `my-lang` | the language a page is asked in, pt-BR — kept in the backoffice_language cookie for the next pages — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `backoffice_language` | cookie | text | no | `my-backoffice-language` | the language chosen before, set by ?lang= — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## Body

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 4 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `days` | text | no | at most 16 characters |
| `reason` | text | no | at most 200 characters |
| `value` | text | no | at most 64 characters |

Example:

```text
days=text&reason=text&value=text
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 4 KB. |
| `415` | The body was not sent with `Content-Type: application/x-www-form-urlencoded`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-language`](backoffice_language.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`ip-blocker`](ip_blocker.md) | always |

---

For developers: `sandbox/internal/routes/add_ip_block_form/` · IP Blocks · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

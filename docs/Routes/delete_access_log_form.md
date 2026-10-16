# `POST /admin/root/delete-access-log`

Deletes access log rows, to its readers

Root readers only. ids are the access log rows to delete. A row younger than the access log retention is kept under a legal obligation in Brazil (Marco Civil art. 15): deleting one asks for the erase-kept-v1 notice at every use, sent back in acknowledge. A row on a legal hold is never deleted. Nothing records the deletion. ip, from, to and kind redraw the page the form came from.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/delete-access-log \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST 'localhost:3000/admin/root/delete-access-log?lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'acknowledge=text&from=text&ids=1&ip=text&kind=text&to=text'
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

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 256 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `acknowledge` | list of texts | no |  |
| `from` | text | no | at most 128 characters |
| `ids` | list of whole numbers | no |  |
| `ip` | text | no | at most 128 characters |
| `kind` | text | no | at most 128 characters |
| `to` | text | no | at most 128 characters |

Example:

```text
acknowledge=text&from=text&ids=1&ip=text&kind=text&to=text
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 256 KB. |
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

---

For developers: `sandbox/internal/routes/delete_access_log_form/` · Privacy · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

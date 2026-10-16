# `POST /admin/root/erase-privacy-subject`

Deletes every record of a person but those the law or a hold keeps

Root only. Deletes every record of the person of ip and text but the held ones and, unless also-access-log=1, the access log rows; deleting access log rows answers the page under a 409 with the erase-kept-v1 notice every time, until acknowledge=erase-kept-v1. The ids deleted of contact-request, visualization-view, agent-usage and agent-summary are kept in privacy-erasure — no personal data, no user — so a restore deletes them again; deleting access log rows keeps nothing. Answers 303 back to the search with notice=erased&count=<deleted>&kept=<kept>.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/erase-privacy-subject \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST localhost:3000/admin/root/erase-privacy-subject \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'acknowledge=text&also-access-log=text&ip=text&text=text'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## Body

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 16 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `acknowledge` | list of texts | no |  |
| `also-access-log` | text | no | at most 8 characters |
| `ip` | text | no | at most 128 characters |
| `text` | text | no | at most 200 characters |

Example:

```text
acknowledge=text&also-access-log=text&ip=text&text=text
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 16 KB. |
| `415` | The body was not sent with `Content-Type: application/x-www-form-urlencoded`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/erase_privacy_subject_form/` · Privacy · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

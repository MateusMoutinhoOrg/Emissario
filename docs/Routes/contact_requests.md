# `POST /api/contact-requests`

Receives one contact request of the website: any JSON value, stored as sent with its time, ip and country

Public, no authentication, same origin only (no CORS). The body is any JSON value — an object, an array, a string, a number — at most the largest body set at /admin/root/set-contact-settings (16384 bytes by default, 1048576 at most); it is stored exactly as sent, with the time, the ip and its country. Answers 201 {"id"}. Errors {"error","field"}: 400 invalid-json, 403 origin, 413 too-large, 415 a Content-Type other than application/json, 429 rate-limit or ip-daily-limit with Retry-After, 503 disabled, not-running, restoring or busy.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/contact-requests \
  -H 'Content-Type: application/json' \
  --data-binary @file.bin
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/contact-requests \
  -H 'Origin: my-origin' \
  -H 'Host: my-host' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  --data-binary @file.bin
```

More examples:

```bash
curl -X POST localhost:3000/api/contact-requests -H 'Content-Type: application/json' -d '{"name":"Ana","email":"ana@example.com","message":"Hello"}'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `Origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser |
| `Host` | header | text | no | `my-host` | the host the request was sent to |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## Body

Send any data — a file, for example with the header `Content-Type: application/json`, up to 1 MB. The body is required.

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 1 MB. |
| `415` | The body was not sent with `Content-Type: application/json`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`ip-blocker`](ip_blocker.md) | always |

---

For developers: `sandbox/internal/routes/contact_requests/` · Contact Requests · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

# `POST /admin/login`

Signs a backoffice user in and sets the session cookie

The username field takes a username or an email. A match sets the session cookie and redirects to /admin/home; anything else answers the login page under a 401. Every attempt is counted before its password is checked. After 20 failed sign-ins from one client ip, or 10 on one account (its username and its email together) from the ips it was not signed in from in the last 30 days, in 15 minutes, every attempt is answered the login page under a 429 with Retry-After, its password unchecked, until the 15 minutes pass; an ip the account signed in from keeps 10 of its own. A body over 4 KiB, a username over 254 bytes or a password over 1024 is refused unchecked.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/login \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'password=text&username=text'
```

With every value it reads:

```bash
curl -X POST localhost:3000/admin/login \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'password=text&username=text'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## Body

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 4 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `password` | text | yes | at most 1024 characters |
| `username` | text | yes | at most 254 characters |

Example:

```text
password=text&username=text
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
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/backoffice_login/` · Backoffice · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

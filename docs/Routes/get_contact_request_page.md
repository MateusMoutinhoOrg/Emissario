# `GET /admin/get-contact-request`

Shows one contact request whole: its JSON laid out as sent, its time, ip and country

Root and viewer; a root also sees the buttons that archive, move back to the inbox and delete it. An id no contact request holds answers the list under a 404.

## Try it

Only what is required:

```bash
curl localhost:3000/admin/get-contact-request
```

With every value it reads:

```bash
curl 'localhost:3000/admin/get-contact-request?id=my-id&notice=my-notice&count=0' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" "localhost:3000/admin/get-contact-request?id=3f9a0c1b2d4e5f60"
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `id` | query string | text | no | `my-id` | the id of the contact request, as /api/contact-requests answered it |
| `notice` | query string | text | no | `my-notice` | the outcome code of the last archive |
| `count` | query string | whole number | no — `0` when left out | `0` | how many contact requests the last archive touched |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |

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

For developers: `sandbox/internal/routes/get_contact_request_page/` · Contact Requests · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

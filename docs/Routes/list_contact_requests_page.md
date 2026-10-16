# `GET /admin/list-contact-requests`

Lists the contact requests, newest first: inbox, archived or all, filtered by period, country and a text of the JSON

Root and viewer. A bad date, from after to, a country that is not two letters or an unknown status answers the page under a 400 with the reason. A root also sees the checkboxes and the buttons that archive, move back to the inbox and delete the contact requests selected.

## Try it

Only what is required:

```bash
curl localhost:3000/admin/list-contact-requests
```

With every value it reads:

```bash
curl 'localhost:3000/admin/list-contact-requests?from=my-from&to=my-to&country=my-country&contains=my-contains&status=my-status&page=1&limit=20&notice=my-notice&count=0&lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" "localhost:3000/admin/list-contact-requests?from=2026-10-01&to=2026-10-14&country=BR&contains=ana"
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `from` | query string | text | no | `my-from` | the first day, YYYY-MM-DD; no bound by default |
| `to` | query string | text | no | `my-to` | the last day, YYYY-MM-DD, included; no bound by default |
| `country` | query string | text | no | `my-country` | an ISO 3166-1 alpha-2 code, such as BR, or XX for an unknown country |
| `contains` | query string | text | no | `my-contains` | a text the JSON holds, any case |
| `status` | query string | text | no | `my-status` | inbox (the default), archived or all |
| `page` | query string | whole number | no — `1` when left out | `1` | the page, from 1 |
| `limit` | query string | whole number | no — `20` when left out | `20` | the contact requests per page, 1 to 100 |
| `notice` | query string | text | no | `my-notice` | the outcome code of the last archive or removal |
| `count` | query string | whole number | no — `0` when left out | `0` | how many contact requests the last archive or removal touched |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `lang` | query string | text | no | `my-lang` | the language a page is asked in, pt-BR — kept in the backoffice_language cookie for the next pages — read by [`backoffice-language`](backoffice_language.md), which runs first |
| `backoffice_language` | cookie | text | no | `my-backoffice-language` | the language chosen before, set by ?lang= — read by [`backoffice-language`](backoffice_language.md), which runs first |
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
| [`backoffice-language`](backoffice_language.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/list_contact_requests_page/` · Contact Requests · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

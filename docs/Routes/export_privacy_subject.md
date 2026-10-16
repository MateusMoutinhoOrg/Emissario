# `GET /admin/root/export-privacy-subject`

Downloads every record of a person as JSON

Root only. Every record of the person of ip and text as a JSON download, as find-privacy-subject answers it: what an access or a portability request asks for (LGPD art. 18 II, V; GDPR arts. 15, 20).

## Try it

Only what is required:

```bash
curl localhost:3000/admin/root/export-privacy-subject
```

With every value it reads:

```bash
curl 'localhost:3000/admin/root/export-privacy-subject?ip=my-ip&text=my-text&lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `ip` | query string | text | no | `my-ip` | the ip of the person: one address, or a hash h:… |
| `text` | query string | text | no | `my-text` | a text the person sent: an email, a phone, a name |
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
| `200` | It worked. The answer comes as `application/json`. |

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

For developers: `sandbox/internal/routes/export_privacy_subject/` · Privacy · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

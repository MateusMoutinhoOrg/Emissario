# `POST /admin/root/set-agent-settings`

Saves the provider and the limits of the public agent

Root only. enabled=1 turns the chat on, its absence off; a blank api-key keeps the one saved, clear-api-key=1 removes it; a blank max-tokens is 1024, a blank rate-per-minute 10, a blank daily limit no limit; summaries-enabled=1 lets visitors ask for summaries, kept summary-retention-days days, blank until a root deletes them. A refused value answers the form again under a 400; a save that needs notices (agent-v1, summaries-v1, retention-summary-v1) answers it under a 409 with them inside, until each id is sent back as acknowledge; a save answers 303 back to the page with ?notice=saved.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/set-agent-settings \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST 'localhost:3000/admin/root/set-agent-settings?lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -b 'backoffice_language=my-backoffice-language' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'acknowledge=text&api-key=text&base-url=text&clear-api-key=text&daily-limit=text&enabled=text&ip-daily-limit=text&max-tokens=text&model=text&rate-per-minute=text&summaries-enabled=text&summary-retention-days=text'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" -X POST localhost:3000/admin/root/set-agent-settings -H 'Origin: http://localhost:3000' --data 'enabled=1&base-url=https://api.openai.com/v1&api-key=sk-...&model=gpt-4o-mini&max-tokens=1024&daily-limit=200000&ip-daily-limit=20000&rate-per-minute=10'
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

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 16 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `acknowledge` | list of texts | no |  |
| `api-key` | text | no | at most 4096 characters |
| `base-url` | text | no | at most 2048 characters |
| `clear-api-key` | text | no | at most 8 characters |
| `daily-limit` | text | no | at most 16 characters |
| `enabled` | text | no | at most 8 characters |
| `ip-daily-limit` | text | no | at most 16 characters |
| `max-tokens` | text | no | at most 16 characters |
| `model` | text | no | at most 200 characters |
| `rate-per-minute` | text | no | at most 16 characters |
| `summaries-enabled` | text | no | at most 8 characters |
| `summary-retention-days` | text | no | at most 16 characters |

Example:

```text
acknowledge=text&api-key=text&base-url=text&clear-api-key=text&daily-limit=text&enabled=text&ip-daily-limit=text&max-tokens=text&model=text&rate-per-minute=text&summaries-enabled=text&summary-retention-days=text
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
| [`backoffice-language`](backoffice_language.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/set_agent_settings_form/` · Agent · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

# `POST /admin/root/set-agent-source`

Sets the git repository, branch and sync time the agent reads its instructions from

Root only. The repository is cloned into <fonts>/agent and its branch pulled every sync time by the goroutine start-server runs; a blank sync time stops the automatic syncing. It may be the repository of the website. A refused value answers the form again under a 400.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/root/set-agent-source \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST localhost:3000/admin/root/set-agent-source \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'branch=text&repo-url=text&sync-time=text'
```

More examples:

```bash
curl -b "backoffice_session=$SESSION" -X POST localhost:3000/admin/root/set-agent-source -H 'Origin: http://localhost:3000' --data 'repo-url=https://github.com/org/site.git&branch=main&sync-time=10m'
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

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 16 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `branch` | text | no | at most 255 characters |
| `repo-url` | text | no | at most 2048 characters |
| `sync-time` | text | no | at most 32 characters |

Example:

```text
branch=text&repo-url=text&sync-time=text
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

For developers: `sandbox/internal/routes/set_agent_source_form/` · Agent · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

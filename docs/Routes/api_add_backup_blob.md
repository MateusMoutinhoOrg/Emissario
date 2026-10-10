# `POST /api/admin/root/add-backup-blob`

Stores the body as one content of the backups, once whatever the number of snapshots that will name it, and answers its sha to reference it by

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/admin/root/add-backup-blob \
  -H 'Content-Type: application/octet-stream' \
  --data-binary @file.bin
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/admin/root/add-backup-blob \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/octet-stream' \
  --data-binary @file.bin
```

More examples:

```bash
curl -X POST localhost:3000/api/admin/root/add-backup-blob -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/octet-stream' --data-binary @data/backofficedb/backoffice-user/1/values/username
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `authorization` | header | text | no | `my-authorization` | an API token created on /admin/list-backoffice-api-tokens, as Bearer <token> — read by [`backoffice-api-token-auth`](backoffice_api_token_auth.md), which runs first |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## Body

Send any data — a file, for example with the header `Content-Type: application/octet-stream`, up to 256 MB. The body is required.

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 256 MB. |
| `415` | The body was not sent with `Content-Type: application/octet-stream`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-api-root-guard`](backoffice_api_root_guard.md) | always |
| [`backoffice-api-token-auth`](backoffice_api_token_auth.md) | always |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |

---

For developers: `sandbox/internal/routes/api_add_backup_blob/` · Backoffice Backups API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

# `GET /api/admin/root/list-backup-files/{Id:integer}`

Lists every file of a snapshot, or the ones whose path starts with a prefix, as its path below the --database folder and the sha of its content

## Try it

Only what is required:

```bash
curl localhost:3000/api/admin/root/list-backup-files/1
```

With every value it reads:

```bash
curl 'localhost:3000/api/admin/root/list-backup-files/1?prefix=my-prefix' \
  -H 'authorization: my-authorization' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method'
```

More examples:

```bash
curl localhost:3000/api/admin/root/list-backup-files/1 -H "Authorization: Bearer $TOKEN"
curl "localhost:3000/api/admin/root/list-backup-files/1?prefix=backofficedb/" -H "Authorization: Bearer $TOKEN"
```

## In the address

| Part | What goes there | Example | Description |
| --- | --- | --- | --- |
| `{Id:integer}` | whole number | `1` |  |

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `prefix` | query string | text | no | `my-prefix` | keeps only the files whose path below the --database folder starts with it |
| `authorization` | header | text | no | `my-authorization` | an API token created on /admin/list-backoffice-api-tokens, as Bearer <token> — read by [`backoffice-api-token-auth`](backoffice_api_token_auth.md), which runs first |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `404` | `{Id:integer}` is not a whole number, so this route does not answer the address. |

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

For developers: `sandbox/internal/routes/api_list_backup_files/` · Backoffice Backups API · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

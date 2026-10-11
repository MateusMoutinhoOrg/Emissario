# `GET /{*Rest}`

Serves the public website: the frontend/ folder of the website source

## Try it

Only what is required:

```bash
curl localhost:3000/my-rest
```

With every value it reads:

```bash
curl localhost:3000/my-rest \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method'
```

## In the address

| Part | What goes there | Example | Description |
| --- | --- | --- | --- |
| `{*Rest}` | the rest of the address — one part or more, like `a/b.png` | `my-rest` | the file under the frontend/ folder of the website source, else that path plus .html; none is its index.html |

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html; charset=utf-8`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`api-get-current-backoffice-user`](api_get_current_backoffice_user.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-api-root-guard`](backoffice_api_root_guard.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-api-token-auth`](backoffice_api_token_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-download-backup`](api_download_backup.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-download-backup-file`](api_download_backup_file.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-list-backup-files`](api_list_backup_files.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-list-backups`](api_list_backups.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-backoffice-api-token-page`](add_backoffice_api_token_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-home`](backoffice_home.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-index`](backoffice_index.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-login-page`](backoffice_login_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-backoffice-api-tokens-page`](list_backoffice_api_tokens_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-backoffice-users-page`](list_backoffice_users_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-same-origin`](backoffice_same_origin.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-backoffice-user-page`](add_backoffice_user_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`download-backup`](download_backup.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-backups-page`](list_backups_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-backoffice-user-page`](set_backoffice_user_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-website-source-page`](set_website_source_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`health`](health.md) | depends on the address — `explain-route` gives the exact answer |
| [`openapi`](openapi.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/front/` · Assets · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

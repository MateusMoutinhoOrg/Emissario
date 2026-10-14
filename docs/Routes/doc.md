# Routes

Every address this server answers. Open one to see what to send, a request you can run as it is,
and what comes back.

The requests call `localhost:3000`, where `emissario start-server` listens when that port is free —
it prints the address it took. Change it to wherever your server runs.

## How to read an address

| In the address | Means | For example |
| --- | --- | --- |
| `GET`, `POST`, … | the method to send it with; `ANY` takes every one | `curl -X POST …` |
| `/users` | exactly that text | `/users` |
| `{name}` | a value you choose | `/users/{tenant}` -> `/users/acme` |
| `{name:integer}` | a value of that type: `integer`, `number` or `uuid` | `/articles/{id:integer}` -> `/articles/42` |
| `{*name}` | the rest of the address, one part or more | `/files/{*file}` -> `/files/a/b.png` |
| `*` | anything else, or nothing | `/admin/*` -> `/admin`, `/admin/users` |
| `(a\|b)` | one of these words | `/(en\|pt)` -> `/en` |

## Agent

| Route | What it does |
| --- | --- |
| [`POST /api/agent/messages`](agent_messages.md) | Answers the next message of the public agent in a conversation the server signed |
| [`POST /api/agent/summaries`](agent_summaries.md) | Keeps an anonymous summary of a conversation the server signed, and answers its id |
| [`GET /admin/agent`](agent_dashboard_page.md) | Shows whether the public agent answers and why not, the tokens of today against the limits, and the latest summaries |
| [`GET /admin/agent-usage`](agent_usage_page.md) | Shows the tokens of the agent: today against the limits, by day, by ip and the ips at their limit |
| [`GET /admin/get-agent-summary`](get_agent_summary_page.md) | Shows one summary whole |
| [`GET /admin/list-agent-summaries`](list_agent_summaries_page.md) | Lists the summaries of a period, newest first, searched by text |
| [`POST /admin/root/add-agent-url`](add_agent_url_form.md) | Adds a url the public agent can call |
| [`GET /admin/root/add-agent-url`](add_agent_url_page.md) | Shows the form that adds a url the public agent can call |
| [`GET /admin/root/list-agent-urls`](list_agent_urls_page.md) | Lists the urls the public agent can call, by name |
| [`POST /admin/root/remove-agent-summaries`](remove_agent_summaries_form.md) | Deletes one summary, or every summary before a date |
| [`POST /admin/root/remove-agent-url`](remove_agent_url_form.md) | Removes a url the public agent can call |
| [`POST /admin/root/set-agent-settings`](set_agent_settings_form.md) | Saves the provider and the limits of the public agent |
| [`GET /admin/root/set-agent-settings`](set_agent_settings_page.md) | Shows the form that sets the provider and the limits of the public agent, beside its state |
| [`POST /admin/root/set-agent-source`](set_agent_source_form.md) | Sets the git repository, branch and sync time the agent reads its instructions from |
| [`GET /admin/root/set-agent-source`](set_agent_source_page.md) | Shows the form that sets the git repository the agent reads its instructions from, and what the last sync did |
| [`POST /admin/root/set-agent-url`](set_agent_url_form.md) | Changes a url the public agent can call |
| [`GET /admin/root/set-agent-url`](set_agent_url_page.md) | Shows the form that changes a url the public agent can call |
| [`POST /admin/root/sync-agent-source`](sync_agent_source_form.md) | Syncs the agent source now, even when it is up to date, and answers at once: the sync runs in the background |

## Agent API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/agent-usage`](api_agent_usage.md) | Answers the tokens of the agent: today against the limits, by day, by ip and the ips at their limit |
| [`POST /api/admin/get-agent-summary`](api_get_agent_summary.md) | Answers one summary whole |
| [`POST /api/admin/list-agent-summaries`](api_list_agent_summaries.md) | Answers one page of the summaries of a period, newest first, searched by text |
| [`POST /api/admin/root/add-agent-url`](api_add_agent_url.md) | Adds a url the public agent can call |
| [`POST /api/admin/root/get-agent-settings`](api_get_agent_settings.md) | Answers the settings of the public agent and its state |
| [`POST /api/admin/root/list-agent-urls`](api_list_agent_urls.md) | Answers the urls the public agent can call, by name |
| [`POST /api/admin/root/remove-agent-summaries`](api_remove_agent_summaries.md) | Deletes one summary, or every summary before a date |
| [`POST /api/admin/root/remove-agent-url`](api_remove_agent_url.md) | Removes a url the public agent can call |
| [`POST /api/admin/root/set-agent-settings`](api_set_agent_settings.md) | Saves the provider and the limits of the public agent |
| [`POST /api/admin/root/set-agent-source`](api_set_agent_source.md) | Sets the git repository, branch and sync time the agent reads its instructions from |
| [`POST /api/admin/root/set-agent-url`](api_set_agent_url.md) | Changes a url the public agent can call |
| [`POST /api/admin/root/sync-agent-source`](api_sync_agent_source.md) | Syncs the agent source now, even when it is up to date, and answers 202 at once: the sync runs in the background |

## Backoffice Users API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/get-backoffice-user`](api_get_backoffice_user.md) | Answers one backoffice user by id |
| [`POST /api/admin/list-backoffice-users`](api_list_backoffice_users.md) | Lists backoffice users, filtered and paginated |
| [`POST /api/admin/root/add-backoffice-user`](api_add_backoffice_user.md) | Adds a backoffice user |
| [`POST /api/admin/root/remove-backoffice-user`](api_remove_backoffice_user.md) | Removes a backoffice user and every session of it |
| [`POST /api/admin/root/set-backoffice-user`](api_set_backoffice_user.md) | Edits a backoffice user; a missing or blank password keeps the current one |

## Backoffice API

| Route | What it does |
| --- | --- |
| [`GET /api/admin/me`](api_get_current_backoffice_user.md) | Answers the backoffice user of the Bearer token |

## Visualization API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/list-visualization-pages`](api_list_visualization_pages.md) | Lists every canonical page of the public website viewed in a period, paginated |
| [`POST /api/admin/list-visualization-views`](api_list_visualization_views.md) | Lists the page views of the public website, newest first, paginated, the ip masked to a viewer |
| [`POST /api/admin/list-visualization-visitors`](api_list_visualization_visitors.md) | Lists every ip that viewed the public website in a period, paginated, masked to a viewer |
| [`POST /api/admin/visualization-metrics`](api_visualization_metrics.md) | Answers the metrics of the page views of a period: the totals, the views over time and the top pages, countries, devices, browsers, operating systems, referrers, not-found paths and visitors |
| [`POST /api/admin/root/get-visualization-settings`](api_get_visualization_settings.md) | Answers the settings of the visualization, the country database in use, how many views are stored, the oldest one and how many were dropped |
| [`POST /api/admin/root/purge-visualization-views`](api_purge_visualization_views.md) | Deletes every page view before a date, or every one, and answers how many were removed |
| [`POST /api/admin/root/set-visualization-settings`](api_set_visualization_settings.md) | Sets the settings of the visualization: tracking on or off, the retention in days and the ignored ips |

## Middleware

| Route | What it does |
| --- | --- |
| [`ANY /api/admin/root/*`](backoffice_api_root_guard.md) | Lets only root users reach /api/admin/root |
| [`ANY /api/admin/*`](backoffice_api_token_auth.md) | Requires a valid API token on every /api/admin route |
| [`ANY /*`](backoffice_client_ip.md) | Works out the client ip every route after it reads, from the connection or the reverse proxy in front |
| [`ANY /*`](backoffice_maintenance.md) | Answers 503 to every request while a restore writes the databases, and to every write while a snapshot reads them |
| [`ANY /admin/root/*`](backoffice_root_guard.md) | Lets only root users reach /admin/root |
| [`ANY /admin/*`](backoffice_same_origin.md) | Refuses a request to /admin another site's page sent |
| [`ANY /*`](backoffice_security_headers.md) | Sends the security headers on every /admin and /api/admin response, and the basic ones on every other |
| [`ANY /admin/* !(/admin/login)`](backoffice_session_auth.md) | Requires a valid backoffice session on /admin, except /admin/login |

## Backoffice Backups API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/root/add-backup-blob`](api_add_backup_blob.md) | Stores the body as one content of the backups, once whatever the number of snapshots that will name it, and answers its sha to reference it by |
| [`POST /api/admin/root/add-backup-file/{Id:integer}/{*Path}`](api_add_backup_file.md) | Stores the body as one file of an open snapshot, at the path below the --database folder the address ends with; a path it already holds is replaced |
| [`POST /api/admin/root/add-backup-reference`](api_add_backup_reference.md) | Adds to an open snapshot one file whose content add-backup-blob already stored, by its sha; a path it already holds is replaced |
| [`POST /api/admin/root/close-backup`](api_close_backup.md) | Closes an open snapshot once every content it names is stored: it turns ready, and can be downloaded and restored |
| [`POST /api/admin/root/create-backup`](api_create_backup.md) | Starts a snapshot of every database, under the name given or one after the current instant, and answers 202 at once: it is taken in the background |
| [`POST /api/admin/root/create-empty-backup`](api_create_empty_backup.md) | Records an empty open snapshot, under the name given or one after the current instant, that files are then added to one by one until it is closed |
| [`GET /api/admin/root/download-backup/{Id:integer}`](api_download_backup.md) | Downloads a snapshot as a zip archive |
| [`GET /api/admin/root/download-backup-file/{Id:integer}/{*Path}`](api_download_backup_file.md) | Downloads one file of a snapshot, at the path below the --database folder the address ends with |
| [`GET /api/admin/root/list-backup-files/{Id:integer}`](api_list_backup_files.md) | Lists every file of a snapshot, or the ones whose path starts with a prefix, as its path below the --database folder and the sha of its content |
| [`GET /api/admin/root/list-backups`](api_list_backups.md) | Lists every snapshot, or the ones whose name starts with a prefix, newest first, and whether a backup job is running |
| [`POST /api/admin/root/optimize-backup-storage`](api_optimize_backup_storage.md) | Starts removing every stored content no snapshot holds anymore, and answers 202 at once |
| [`POST /api/admin/root/remove-backup`](api_remove_backup.md) | Deletes a snapshot; the contents only it held stay stored until the backups size is optimized |
| [`POST /api/admin/root/restore-backup`](api_restore_backup.md) | Starts putting a snapshot back over every database, after a pre-restore snapshot of them, and answers 202 at once |
| [`POST /api/admin/root/upload-backup`](api_upload_backup.md) | Stores a zip archive a download built as a new snapshot |

## Website Source API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/root/sync-website-source`](api_sync_website_source.md) | Syncs the website source now, even when it is up to date, and answers 202 at once: the sync runs in the background |

## Backoffice API Tokens

| Route | What it does |
| --- | --- |
| [`POST /admin/add-backoffice-api-token`](add_backoffice_api_token_form.md) | Creates an API token and shows it once |
| [`GET /admin/add-backoffice-api-token`](add_backoffice_api_token_page.md) | Shows the form that creates an API token |
| [`GET /admin/list-backoffice-api-tokens`](list_backoffice_api_tokens_page.md) | Lists your API tokens, or every user's for a root |
| [`POST /admin/revoke-backoffice-api-token/{Id:integer}`](revoke_backoffice_api_token_form.md) | Revokes an API token: your own, or anyone's for a root |

## Backoffice

| Route | What it does |
| --- | --- |
| [`GET /admin/home`](backoffice_home.md) | Shows the backoffice home page to the signed-in user |
| [`GET /admin`](backoffice_index.md) | Sends the signed-in user to the backoffice home page |
| [`POST /admin/login`](backoffice_login.md) | Signs a backoffice user in and sets the session cookie |
| [`GET /admin/login`](backoffice_login_page.md) | Shows the sign-in form |
| [`POST /admin/logout`](backoffice_logout.md) | Ends the current session |

## Backoffice Users

| Route | What it does |
| --- | --- |
| [`GET /admin/list-backoffice-users`](list_backoffice_users_page.md) | Lists backoffice users, filtered and paginated |
| [`POST /admin/root/add-backoffice-user`](add_backoffice_user_form.md) | Adds a backoffice user |
| [`GET /admin/root/add-backoffice-user`](add_backoffice_user_page.md) | Shows the form that adds a backoffice user |
| [`POST /admin/root/remove-backoffice-user/{Id:integer}`](remove_backoffice_user_form.md) | Removes a backoffice user and every session of it |
| [`POST /admin/root/set-backoffice-user/{Id:integer}`](set_backoffice_user_form.md) | Edits a backoffice user; a blank password keeps the current one |
| [`GET /admin/root/set-backoffice-user/{Id:integer}`](set_backoffice_user_page.md) | Shows the form that edits a backoffice user |

## Visualization

| Route | What it does |
| --- | --- |
| [`GET /admin/list-visualization-pages`](list_visualization_pages_page.md) | Lists every canonical page of the public website viewed in a period: its views, unique visitors and last view |
| [`GET /admin/list-visualization-views`](list_visualization_views_page.md) | Lists the page views of the public website, newest first, and the summary of one visitor when filtered by one ip |
| [`GET /admin/list-visualization-visitors`](list_visualization_visitors_page.md) | Lists every ip that viewed the public website in a period: its country, device, browser, os, views and pages, masked to a viewer |
| [`GET /admin/root/export-visualization-views`](export_visualization_views.md) | Downloads the page views of a period as CSV, newest first, 100 000 at most |
| [`POST /admin/root/purge-visualization-views`](purge_visualization_views_form.md) | Deletes every page view before a date, or every one |
| [`POST /admin/root/set-visualization-settings`](set_visualization_settings_form.md) | Saves the settings of the visualization: tracking on or off, the retention in days and the ignored ips |
| [`GET /admin/root/set-visualization-settings`](set_visualization_settings_page.md) | Shows the settings of the visualization: tracking, retention, ignored ips, the country database in use and the Delete views form |
| [`GET /admin/visualization`](visualization_dashboard_page.md) | Shows the dashboard of the page views of the public website: cards, views over time and the top pages, countries, devices, browsers, operating systems, referrers, not-found paths and visitors |

## Backoffice Backups

| Route | What it does |
| --- | --- |
| [`POST /admin/root/create-backup`](create_backup_form.md) | Starts a snapshot of every database, under the name the form gives or one after the current instant, and answers at once: it is taken in the background |
| [`GET /admin/root/download-backup/{Id:integer}`](download_backup.md) | Downloads a snapshot as a zip archive |
| [`GET /admin/root/list-backups`](list_backups_page.md) | Lists every snapshot of the databases, or the ones whose name starts with a prefix, with the controls that take, download, upload and restore one |
| [`POST /admin/root/optimize-backup-storage`](optimize_backup_storage_form.md) | Starts removing every stored content no snapshot holds anymore, and answers at once |
| [`POST /admin/root/remove-backup/{Id:integer}`](remove_backup_form.md) | Deletes a snapshot; the contents only it held stay stored until the backups size is optimized |
| [`POST /admin/root/restore-backup/{Id:integer}`](restore_backup_form.md) | Starts putting a snapshot back over every database, after a pre-restore snapshot of them, and answers at once |
| [`POST /admin/root/upload-backup`](upload_backup.md) | Stores a zip archive a download built as a new snapshot |

## Website Source

| Route | What it does |
| --- | --- |
| [`POST /admin/root/set-website-source`](set_website_source_form.md) | Sets the git repository, branch and sync time the public website is served from |
| [`GET /admin/root/set-website-source`](set_website_source_page.md) | Shows the form that sets the git repository the public website is served from, and what the last sync did |
| [`POST /admin/root/sync-website-source`](sync_website_source_form.md) | Syncs the website source now, even when it is up to date, and answers at once: the sync runs in the background |

## Assets

| Route | What it does |
| --- | --- |
| [`GET /{*Rest}`](front.md) | Serves the public website: the frontend/ folder of the website source |

## Server

| Route | What it does |
| --- | --- |
| [`GET /health`](health.md) | Reports that the server is up |
| [`GET /openapi.json`](openapi.md) | Answers the OpenAPI document of every route, to import into Postman or open in Swagger UI |

## Postman, Swagger and other tools

[openapi.json](openapi.json) holds every route below as an OpenAPI 3.0 document, and the running
server answers the same file at `/openapi.json`:

- **Postman**: *Import*, then pick the file or paste `http://localhost:3000/openapi.json`. Every
  route lands in a folder named after its section.
- **Swagger UI**, or any tool that reads OpenAPI: open `http://localhost:3000/openapi.json`.

A route that runs in front of others for every method — a check of a token, for example — is
not listed on its own there: what it reads is listed on each route it guards, and a token sent
as `Authorization` is the document's sign-in.

## When something goes wrong

| Status | Means |
| --- | --- |
| `400` | Something you sent is missing or has the wrong type or format |
| `401` | You have to identify yourself first — a token, for example |
| `403` | You are identified, but not allowed to do this |
| `404` | No route answers this address |
| `405` | The address exists, but not for this method — a `GET` where it takes a `POST`, for example |
| `413` | The body is too large |
| `415` | The body is not in the format the route reads — check `Content-Type` |
| `500` | The server failed while answering |

Unless the project changed it, the answer to an error is JSON naming what went wrong and, when
it is one value, which one:

```json
{"error": "required parameter 'authorization' is missing", "field": "authorization"}
```

For developers: each page, and `openapi.json`, is generated on every build from
`sandbox/internal/routes/<name>/route.yaml` ([RouteYaml](../RouteYaml/doc.md#openapi)); hidden
routes are left out. `agnos list-routes` prints the routes in the order they run, and
`agnos explain-route <METHOD> <path>` which ones a request reaches. The error answers are
the eight files of `sandbox/internal/server/errors/` ([RouteYaml](../RouteYaml/doc.md#failures)).

# `POST /api/agent/messages`

Answers the next message of the public agent in a conversation the server signed

No authentication: any visitor of the website may call it, from the same origin (no CORS; a request another site sent is answered 403). The history lives in the visitor and is signed by the server: send every message of the conversation, user and assistant in turns, the first and the last of the user, with the signature the last answer carried (empty on the first message); the answer is the next message of the agent and the signature covering the messages and it. The model reads instructions/README.md of the agent source as its system prompt and may list and read the files of instructions/ and call the urls a root declared, for up to 5 rounds. 400 invalid, 401 invalid-signature (the backoffice secret changed at a restart: start over), 403 origin, 413 over 64 KiB, 429 rate-limit, ip-daily-limit or daily-limit with Retry-After, 502 the provider failed, 503 disabled, not configured, no README.md or a restore running, 504 past the 60 s deadline. The conversation is sent to the provider configured, a third party, and is never stored.

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/api/agent/messages \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"content":"text","role":"user"}]}'
```

With every value it reads:

```bash
curl -X POST localhost:3000/api/agent/messages \
  -H 'Origin: my-origin' \
  -H 'Host: my-host' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-client-port: my-x-client-port' \
  -H 'x-request-method: my-x-request-method' \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"content":"text","role":"user"}],"signature":"text"}'
```

More examples:

```bash
curl -X POST localhost:3000/api/agent/messages -H 'Content-Type: application/json' -d '{"messages":[{"role":"user","content":"Hello"}],"signature":""}'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `Origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser |
| `Host` | header | text | no | `my-host` | the host the request was sent to |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-client-port` | header | text | no | `my-x-client-port` | the source port of the visitor's connection, set by the server and never by the client — or by the reverse proxy start-server trusts with --client-port-header — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |

## Body

Send JSON with the header `Content-Type: application/json`, up to 64 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `messages` | list of objects | yes | 1 to 50 items |
| `messages[].content` | text | yes | at most 4000 characters |
| `messages[].role` | text | yes | one of `user`, `assistant` |
| `signature` | text | no | at most 128 characters |

Example:

```json
{
  "messages": [
    {
      "content": "text",
      "role": "user"
    }
  ],
  "signature": "text"
}
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `413` | The body is larger than 64 KB. |
| `415` | The body was not sent with `Content-Type: application/json`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |

---

For developers: `sandbox/internal/routes/agent_messages/` · Agent · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)

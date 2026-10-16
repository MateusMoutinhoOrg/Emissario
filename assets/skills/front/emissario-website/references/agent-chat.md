# AI chat: `POST /api/agent/messages`

Emissario provides the chat API; the site builds the chat UI. A visitor's message goes to an AI model
that the people who run the site configure. They also write its instructions, in another repository.
Emissario stores no conversation: **the history lives in the page** (memory or `localStorage`), and
the server signs it so that it cannot be forged.

## Tell the visitor it is an AI, before the chat starts

The law (EU AI Act art. 50(1)) requires people to be told they are talking to an AI. The model is told
to say so, but a model can fail; the page cannot. Put a visible line above the chat box, shown before
the first message:

```html
<p class="chat-disclosure">You are chatting with an AI assistant.</p>
```

Never present the agent as a person (a human name and photo with no disclosure, "typing…" from
"Ana").

## The protocol

Each request sends **the whole conversation** and the signature of the last answer:

```json
{"messages": [{"role": "user", "content": "Hello"}], "signature": ""}
```

- `messages`: `user` and `assistant` in turns, the first and the last from the `user`. At most
  {{.AgentMaxMessages}} messages, {{.AgentMaxContentLength}} characters each, a {{.AgentMaxBodyKiB}} KiB body.
- `signature`: empty on the first message; afterwards, the one the last answer carried. It covers every
  message but the last.

`200`:

```json
{"message": {"role": "assistant", "content": "Hi! I am an AI assistant…"}, "signature": "<covers the messages and this one>"}
```

The next request sends every message, this answer included, and this new signature. A reply is at most
{{.AgentMaxContentLength}} characters, so the history sent back always passes. Never edit a message
already answered: the signature would no longer match.

- A conversation reaching {{.AgentMaxMessages}} messages cannot take another: offer to start a new one.
- A request may take up to {{.AgentDeadlineSeconds}} s (the model may read files and call tools first):
  show that it is working, and keep the send button off until it answers.
- Answers are plain text. Render them as text (`textContent`), never as HTML.

## Errors: `{"error": "<code>", "field": "..."}`

| Status | `error` | What the page does |
|---|---|---|
| `400` | why the body is not a conversation; `field` names the part (`messages[2].role`) | a bug of the page |
| `401` | `invalid-signature` | the history no longer matches (the server restarted with a new secret): **start a new conversation**, then send the visitor's message again |
| `403` | `origin`, `blocked` | the chat is not available here |
| `413` | — | the body is over {{.AgentMaxBodyKiB}} KiB: start a new conversation |
| `429` | `rate-limit` | too many messages: wait a minute (`Retry-After: 60`) |
| `429` | `ip-daily-limit`, `daily-limit` | the chat is closed until tomorrow (`Retry-After`: the seconds until midnight) |
| `502` | `provider-failed` | the AI could not answer: try again later |
| `503` | `disabled`, `not-configured`, `no-instructions`, `not-running` | the chat is unavailable: hide it or say so |
| `503` | a backup is being restored | try again in a few seconds (`Retry-After: 5`) |
| `504` | `deadline` | it took too long: try again |

The chat is **off until the people who run the site turn it on**, so `503 disabled` is a normal
answer. Each IP may send {{.AgentRatePerMinute}} messages a minute by default.

## A client

```js
const KEY = "chat";
let {history, signature} = load();

function load() {
  try { return JSON.parse(localStorage.getItem(KEY)) ?? {history: [], signature: ""}; }
  catch { return {history: [], signature: ""}; }
}
function save() {
  try { localStorage.setItem(KEY, JSON.stringify({history, signature})); } catch {}
}
function reset() { history = []; signature = ""; save(); }

async function send(text, retried = false) {
  const messages = [...history, {role: "user", content: text}];
  const response = await fetch("/api/agent/messages", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({messages, signature}),
  });
  if (response.status === 401 && !retried) { reset(); return send(text, true); }
  const answer = await response.json().catch(() => ({}));
  if (!response.ok) throw Object.assign(new Error(answer.error ?? "failed"), {status: response.status});
  history = [...messages, answer.message];
  signature = answer.signature;
  save();
  return answer.message.content;
}
```

Keeping the history in `localStorage` is what lets the conversation survive a reload; it is strictly
necessary for the chat the visitor asked for, so it needs no consent banner. Offer a "New conversation"
button that calls `reset()`.

## Summaries (optional): `POST /api/agent/summaries`

A page may hand the conversation over as an anonymous summary for the people who run the site, for
example behind a "Send this conversation to the team" button. The visitor should choose it.

```json
{"messages": [/* the whole conversation, ending with the last answer of the agent */], "signature": "<the one that answer carried>"}
```

- `201 {"id": "<16 hex characters>"}`: show the id to the visitor as a reference. The summary text is
  never answered.
- Off by default: until the owners turn summaries on, it answers `503 summaries-disabled`. Hide the
  button then.
- Otherwise the same errors as `/api/agent/messages`.

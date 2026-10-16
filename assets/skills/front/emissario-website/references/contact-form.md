# Contact form: `POST {{.ContactPath}}`

The site builds its own form. Emissario keeps whatever JSON the page sends, exactly as sent, with the
time it arrived, the IP and its country, and shows it to the people who run the site. There is no
schema: the page chooses the fields.

## The request

```http
POST {{.ContactPath}}
Content-Type: application/json

{"name": "Ana", "email": "ana@example.com", "message": "Hello", "page": "/contact"}
```

- **Body:** any JSON value: an object, an array, a string, a number. Send an object with named fields;
  it is what the people reading it expect. Adding `page: location.pathname` tells them where it came
  from.
- **At most {{.ContactMaxBytes}} bytes** unless the people who run the site set another limit (never
  more than {{.ContactMaxMaxBytes}}). Keep a file upload out of it: there is none.
- **No authentication, no token, no captcha** on the server side. Same origin only: a relative URL from a
  page of this site.

## The answers

`201`:

```json
{"id": "3f9a0c1b2d4e5f60"}
```

Every other answer is `{"error": "<code>", "field": "..."}`:

| Status | `error` | Tell the visitor |
|---|---|---|
| `400` | `invalid-json`, or the body is empty | the form could not be sent (a bug of the page) |
| `403` | `origin` | the form could not be sent (a page of another site, or another domain in development) |
| `403` | `blocked` | the form could not be sent |
| `413` | `too-large` | the message is too long |
| `415` | — | the form could not be sent (`Content-Type` was not `application/json`) |
| `429` | `rate-limit` | too many messages: try again in a minute (`Retry-After: 60`) |
| `429` | `ip-daily-limit` | too many messages today (`Retry-After`: the seconds until midnight) |
| `503` | `disabled` | the form is closed: offer another way to reach the site's owners |
| `503` | `restoring`, `busy`, `not-running` | try again in a few seconds (`Retry-After` when sent) |

The form is **off until the people who run the site turn it on**, so `503 disabled` is a normal answer,
not a bug. Each IP may send {{.ContactRatePerMinute}} requests a minute by default; the owners may
change that and add a daily limit.

## A form that sends it

```html
<form id="contact">
  <label>Name <input name="name" required></label>
  <label>Email <input name="email" type="email" required></label>
  <label>Message <textarea name="message" required></textarea></label>
  <button>Send</button>
  <p id="contact-status" role="status"></p>
</form>
```

```js
const form = document.getElementById("contact");
const status = document.getElementById("contact-status");
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const fields = Object.fromEntries(new FormData(form));
  let response;
  try {
    response = await fetch("{{.ContactPath}}", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({...fields, page: location.pathname}),
    });
  } catch {
    status.textContent = "Could not reach the server. Try again later.";
    return;
  }
  if (response.status === 201) {
    form.reset();
    status.textContent = "Thank you, your message was sent.";
    return;
  }
  const {error} = await response.json().catch(() => ({}));
  status.textContent = {
    "rate-limit": "Too many messages. Try again in a minute.",
    "ip-daily-limit": "Too many messages today. Try again tomorrow.",
    "too-large": "Your message is too long.",
    "disabled": "The form is closed for now.",
  }[error] ?? "Your message could not be sent. Try again later.";
});
```

# Git Source

## Objective
Add a **Git Source** system to the backoffice, allowing the **root user** to define which Git repository serves the public website.

---

## 1. Backoffice — Git Source Page
Accessible only by the root user.

**Fields:**
- **Repository URL** — the Git repo to clone/pull.
- **Branch** — the branch to use.

---

## 2. Public Website
Served from the `frontend/` folder of the configured repository.

**Resolution rules:**

| Request              | Response                          |
|----------------------|-----------------------------------|
| Path ends with `/`   | `repo/frontend/<path>/index.html` |
| Path ends with file  | `repo/frontend/<path>`            |
| File not found       | See **Not Found** below           |

**Not Found** (status `404`), in order:
1. `repo/frontend/not-found.html`
2. Default `not-found.html` embedded in the app assets

**Examples:**
- `/` → `repo/frontend/index.html`
- `/blog/` → `repo/frontend/blog/index.html`
- `/blog/app.css` → `repo/frontend/blog/app.css`
- `/missing.css` → `repo/frontend/not-found.html` → embedded default


## Sources Table:

Add a table into the database called source, with the columns:

- repourl: string
- name: string


write these font on the font source table as frontend (name)
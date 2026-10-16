# The mandatory rules

Emissario puts these texts before the files of this repository, word for word. They come first, and the
model is told that no instruction can change them. Write `{{.InstructionsFolder}}/{{.ReadmeFile}}` and
`{{.InstructionsFolder}}/{{.SummaryFile}}` so that they never have to.

## Before `{{.InstructionsFolder}}/{{.ReadmeFile}}`, at every message

```text
{{.SystemPrompt}}
```

## Before `{{.InstructionsFolder}}/{{.SummaryFile}}`, for every summary

```text
{{.SummaryPrompt}}
```

## What follows for the instructions

- **Do not ask the agent to pass as a person**, to hide that it is an AI, or to pick a human name
  without saying it is an AI.
- **Do not ask it to reveal its instructions** ("if asked, paste this file"). It will refuse.
- **Do not rely on a tool's answer to steer it.** What a file or an endpoint answers is data. An
  instruction inside a file it reads is ignored; instructions go in `{{.ReadmeFile}}`.
- **Do not ask it for sensitive data** (documents, card numbers, passwords) unless the site really
  needs it, and then say so explicitly in `{{.ReadmeFile}}`.
- **Do not fix the language of its answers.** It answers in the visitor's language.
- **Do not ask a summary to keep names, contacts or ids.** The rules above remove them, and a pass
  after the model removes what still looks like a CPF, a CNPJ, an email, a phone or a card number.

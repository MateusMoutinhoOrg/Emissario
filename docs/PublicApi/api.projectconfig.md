# `sandbox/api/projectconfig.go`

| Constant | Value | Description |
| --- | --- | --- |
| `DefaultFontsDir` | `"EmissarioFonts"` | DefaultFontsDir is the folder the git sources are cloned into when no command line says otherwise: `--fonts` defaults to it, and a project used as a lib keeps it. |

## `ProjectConfig`

ProjectConfig is the part of the Config the project declares itself: api.Config embeds it, so every field typed here is read as sandbox.Config.<Field>, beside the ProjectName and Version the build renders from project.yaml: type ProjectConfig struct { Port int } Fill it in sandbox/constructors/config/constructor.go, after NewConfig — that file is written once and is yours too. A field is written in the builtin types or in a type of this package, like every other of sandbox/api/. Written once by `agnos start` and then yours: no build rewrites this file.

| Field | Type | Description |
| --- | --- | --- |
| `FontsDir` | `string` | FontsDir is the folder every git source is cloned into, one folder per source named after it — the frontend source in <FontsDir>/frontend — relative to the directory the program runs from, or absolute. The git-source-start-server middleware sets it from `--fonts` in front of start-server; the config constructor starts it at DefaultFontsDir. |

[every contract](doc.md)

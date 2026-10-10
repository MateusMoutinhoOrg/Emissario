# Structure

`(gen)` = written by `build`, never edited — the full list is in
[GeneratedFiles](../GeneratedFiles/doc.md).

```
adapters/  -->  sandbox/  <--  cmd/
(reaches OS)    (closed)       (wires)
```

Every line below is one entry of `AgnosConfig/structure.yaml` — add `<path>:
{description: "..."}` there, nested under `children:` of its parent, with `dir: true` on a
directory, `gen: true` on a file `build` rewrites, and `order:` to place it among its siblings
(unordered siblings follow, alphabetically).

```
AgnosConfig/        written once by `start`, read by every `build`
sandbox/            closed: imports nothing outside sandbox/, no OS packages
  generated.new.go  (gen) New(deps) *api.Sandbox, one <x>.New<X> per api/ file
  api/              contracts only; imports nothing at all
  internal/         the logic; unreachable from outside the sandbox
docs/               one dir per doc, holding doc.md + doc.yaml. README.md indexes them all
go.mod              written by `start`; add-dep and remove-dep edit its require block
README.md           (gen) `render AgnosConfig/docs/ReadmeHeader.md` + the documentation index
```

Every rule this shape has to hold to — layers, naming, generated files, docs — is in
[Rules](../Rules/doc.md); the command that makes each change is in
[Workflow](../Workflow/doc.md).

# LibUsage

`emissario` is a Go module before it is anything else: every feature lives in `sandbox/`
and is reachable from any Go program that imports it.

```bash
go get github.com/MateusMoutinhoOrg/Emissario@latest
```

## Wiring

`sandbox.New` takes no arguments and returns the API object.

```go
package main

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox"
)

func main() {
	lib := sandbox.New() // *api.Sandbox

	_ = lib
}
```

## What the sandbox exposes

`*api.Sandbox` is a flat struct, one field per contract declared in `sandbox/api/`.
Everything callable from Go is behind one of them.

| Field | Type |
| --- | --- |
| `lib.Config` | `api.Config` |

[PublicApi](../PublicApi/doc.md) lists every one of them — signatures, props structs — generated from `sandbox/api/` itself on every build.

`sandbox/api` is pure contract and `sandbox/` never touches the OS, so both are safe to import
anywhere; the rest of the rules a caller can count on are in [Rules](../Rules/doc.md#layers),
and [DepList](../DepList/doc.md) lists every contract that can be added.

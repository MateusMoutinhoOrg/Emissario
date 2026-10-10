# emissario

[![Go Reference](https://pkg.go.dev/badge/github.com/MateusMoutinhoOrg/Emissario.svg)](https://pkg.go.dev/github.com/MateusMoutinhoOrg/Emissario)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue)](go.mod)

Place your description in AgnosConfig/docs/ReadmeHeader.md



## Documentation

### LibUsage

Using the project as a Go module - wiring the deps, calling the sandbox

| Doc | Description |
| --- | --- |
| [LibUsage](docs/LibUsage/doc.md) | Use emissario as a Go module: wire the deps, build the sandbox, call its API |
| [PublicApi](docs/PublicApi/doc.md) | Every exported symbol of emissario, generated from the contract sources and their doc comments |
| [LibExamples](docs/LibExamples/doc.md) | Index of every runnable example of emissario as a Go module |

### Architecture

How the project is put together - layers, boundaries, data flow

| Doc | Description |
| --- | --- |
| [Adapters](docs/Adapters/doc.md) | Contract, adapter and binding: three units, one field of Deps, and who fills it |

### Development

Changing this repository - schema, build mechanics, recipes

| Doc | Description |
| --- | --- |
| [Requirements](docs/Requirements/doc.md) | The two tools this project needs — Go and agnos — installed per platform |
| [Workflow](docs/Workflow/doc.md) | Every change this project takes and the agnos command that makes it |
| [Rules](docs/Rules/doc.md) | Every rule the generators, `verify` and the hand-written files must hold to |
| [Structure](docs/Structure/doc.md) | The project schema: what lives where, what is generated, what verify enforces |

### Reference

Lookup tables - schemas, file formats, generated file listings

| Doc | Description |
| --- | --- |
| [Extensions](docs/Extensions/doc.md) | The generation mechanics this project turns on, and what each one writes |
| [DepList](docs/DepList/doc.md) | Every dep `agnos add-dep` can add, the adapters that fill it, and what backs each one |
| [GeneratedFiles](docs/GeneratedFiles/doc.md) | Every file agnos writes into this project and whether build overwrites it |

## License

Place your license in this file. It is pasted verbatim into the License section of README.md.


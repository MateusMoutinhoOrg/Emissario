# `deps.OpinionatedAgnosDatabase`

`sandbox/deps/OpinionatedAgnosDatabase`

## `Contract`

Contract is the database lib injected whole as the Deps.OpinionatedAgnosDatabase field.

| Field | Type | Description |
| --- | --- | --- |
| `Fail` | `func(failure *databasedeps.Error) error` | Fail turns one failure the database reported into an error the sandbox carries. A nil *databasedeps.Error is success and answers nil. |
| `Collection` | `func(database databasedeps.Database, name string) (databasedeps.Collection, error)` | Collection resolves one collection of a database by name. A name the database does not declare, or a database Databases.New refused, is an error rather than a zero collection. |
| `ReadString` | `func(record databasedeps.Record, field string) (string, error)` | ReadString reads one Key or String field of a record. |
| `ReadInt` | `func(record databasedeps.Record, field string) (int64, error)` | ReadInt reads one Int or Link field of a record. |
| `ReadFloat` | `func(record databasedeps.Record, field string) (float64, error)` | ReadFloat reads one Float field of a record. |
| `ReadBytes` | `func(record databasedeps.Record, field string) ([]byte, error)` | ReadBytes reads one Bytes field of a record, byte for byte. |
| `TextMatches` | `func(value string, startsWith string, equals string) bool` | TextMatches is the filter a generated <T>Filter applies to one text field: an empty needle passes everything, so a zero value turns the filter off. |
| `BytesMatches` | `func(value []byte, startsWith []byte, equals []byte) bool` | BytesMatches is TextMatches for a binary field, compared byte for byte: an empty needle passes everything. |
| `IntInRange` | `func(value int64, min int64, max int64) bool` | IntInRange is TextMatches for a whole-number field: a zero bound is no bound. |
| `FloatInRange` | `func(value float64, min float64, max float64) bool` | FloatInRange is TextMatches for a floating-point field. |

[every contract](doc.md)

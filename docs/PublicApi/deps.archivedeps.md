# `deps.ArchiveDeps`

`sandbox/deps/archivedeps`

## `File`

File is one entry of an archive: where it sits in the archive and what it holds.

| Field | Type | Description |
| --- | --- | --- |
| `Path` | `string` | Path is the entry's name inside the archive, slash-separated and relative to its root ("data/backofficedb/backoffice-user/size"). It is handed back exactly as the archive spells it: whether it is safe to write anywhere is the caller's to check. |
| `Content` | `[]byte` | Content is the entry's whole uncompressed content. |

## `Contract`

Contract is the archive library injected whole as the Deps.ArchiveDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `Zip` | `func(files []File) ([]byte, error)` | Zip packs files, in the order given, into one zip archive and returns its bytes. Every entry is compressed and carries no modification time, so the same files always pack into the same bytes. |
| `Unzip` | `func(archive []byte, maxBytes int64) ([]File, error)` | Unzip reads every file of the zip archive in archive, in the order the archive lists them. Directory entries are not reported. The error reports an archive that is not a zip or is damaged, and one whose files add up to more than maxBytes once uncompressed — which is what keeps a small archive from unpacking into all of the memory. |

[every contract](doc.md)

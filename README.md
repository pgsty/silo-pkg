# silo-pkg

Pigsty's maintained fork of [minio/pkg](https://github.com/minio/pkg), the
collection of common packages used in MinIO projects. It exists so the
community MinIO fork has somewhere to take fixes that upstream, now driven by a
closed product, will not carry.

## Using it

The module path is deliberately **unchanged**. This repository still declares
`module github.com/minio/pkg/v3`, so every `import "github.com/minio/pkg/v3/..."`
keeps working and the fork stays a drop-in replacement. Only the right-hand side
of each `replace` directive names the maintained repository:

```go
replace (
	github.com/minio/pkg/v3 => github.com/pgsty/silo-pkg/v3 v3.12.2
	github.com/minio/minio-go/v7 => github.com/pgsty/silo-go/v7 v7.3.1
)
```

The second replacement selects the Silo Go SDK used by this package. Go does
not inherit `replace` directives from dependency modules, so top-level
consumers that want the complete maintained stack must declare both
replacements in their own `go.mod`.

The `/v3` suffix is required — it is the module's major version, not a directory.

The repository was renamed from `pgsty/minio-pkg` on 2026-08-02. GitHub redirects
the old path, but pin the new one.

## Versioning

Tags follow upstream's numbering so it is obvious which release a version is
based on. They do not promise identical contents: this fork skips upstream work
that only serves the closed product, and carries fixes upstream has not taken.
Read the tag annotation before upgrading — some releases have to be taken
together with a matching MinIO server change, and say so.

## License

Use of this package is governed by the GNU AGPLv3 license that can be found in
the [LICENSE](./LICENSE) file.

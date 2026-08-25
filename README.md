# silo-pkg

Pigsty's maintained fork of [minio/pkg](https://github.com/minio/pkg), the
collection of common packages used in MinIO projects. It exists so the
community MinIO fork has somewhere to take fixes that upstream, now driven by a
closed product, will not carry.

## Using it

The module path is deliberately **unchanged**. This repository still declares
`module github.com/minio/pkg/v3`, so every `import "github.com/minio/pkg/v3/..."`
keeps working and the fork stays a drop-in replacement. Only the right-hand side
of the `replace` directive names this repository:

```
replace github.com/minio/pkg/v3 => github.com/pgsty/silo-pkg/v3 v3.12.1
```

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

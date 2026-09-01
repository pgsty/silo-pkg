# silo-pkg

Pigsty's maintained fork of [minio/pkg](https://github.com/minio/pkg), the
collection of common packages used in MinIO projects. It exists so the
community MinIO fork has somewhere to take fixes that upstream, now driven by a
closed product, will not carry.

## Using it

Import it directly. This repository declares `module github.com/pgsty/silo-pkg/v3`,
so a consumer requires it by name and needs no `replace` directive:

```go
require github.com/pgsty/silo-pkg/v3 v3.13.1
```

```go
import "github.com/pgsty/silo-pkg/v3/policy"
```

The `/v3` suffix is required — it is the module's major version, not a directory.

Through v3.12.2 the module kept upstream's `github.com/minio/pkg/v3` path and was
selected with a `replace` directive. That worked, but Go does not inherit `replace`
directives from dependency modules, so every consumer had to repeat the redirect,
and the `require` line had to name an upstream version this fork's source no longer
matched. v3.13.0 and later own the path instead. Consumers still on the old arrangement keep
building against the versions they already pinned; to move, drop the `replace`,
require this path, and rewrite the imports.

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

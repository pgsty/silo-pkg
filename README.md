# silo-pkg

Pigsty's maintained fork of [minio/pkg](https://github.com/minio/pkg), the
collection of common packages used in MinIO projects. It exists so the
community MinIO fork has somewhere to take fixes that upstream, now driven by a
closed product, will not carry.

## Using it

Import it directly. This repository declares `module github.com/pgsty/silo-pkg/v3`,
so a consumer requires it by name and needs no `replace` directive:

```go
require github.com/pgsty/silo-pkg/v3 v3.14.0
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

## Go and TLS compatibility

The library retains its Go 1.26 floor and is also tested with Go 1.27. Runtime
defaults depend on the consuming application's Go version and `GODEBUG`, not
just this library's `go.mod`. The web-environment client leaves TLS key exchange
at Go defaults; LDAP and OIDC helpers also preserve caller-supplied TLS settings.
For default-configured TLS, `GODEBUG=tlsmlkem=0` disables hybrid key exchanges;
`GODEBUG=tlssecpmlkem=0` disables only the SecP hybrids and retains X25519MLKEM768.
These settings preserve certificate verification.

On macOS, applications targeting Go 1.27 replace Keychain trust with on-disk
roots and Go's verifier when either `SSL_CERT_FILE` or `SSL_CERT_DIR` is set.
Stale or incomplete CA paths can break previously trusted connections; unset
inherited values to restore Keychain trust. Explicit CAs supplied to
`certs.GetRootCAs` remain additive to the selected root pool.
An application still targeting Go 1.26 retains the
old platform default unless it opts in with
`GODEBUG=x509sslcertoverrideplatform=1`. The Windows loader in this package reads
the Windows ROOT store directly and is unchanged. See the
[Go release notes](https://go.dev/doc/go1.27).

## Versioning

SILO versions are released independently of upstream. Read the release notes
before upgrading: changes to shared policy behavior must be adopted together
with the matching SILO server, Console and mcli changes.

Version 3.14.0 separates self-service password changes from user administration.
Review the [password authorization migration](UPSTREAM.md#breaking-authorization-compatibility)
when upgrading policies that deny `admin:CreateUser` or `admin:ChangeMyPassword`.

## Contributing

Submit issues and pull requests to [pgsty/silo-pkg](https://github.com/pgsty/silo-pkg).
Code contributions use AGPL-3.0-or-later, the same license as this package.
Contributors retain their copyright; no CLA or separate Apache-2.0 license grant
is required. Sign commits with `git commit -s`; see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Use of this package is governed by the GNU AGPL v3.0 or later license in
the [LICENSE](./LICENSE) file.

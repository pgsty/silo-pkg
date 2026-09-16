# Changelog

## v3.14.1 — 2026-09-16

[GitHub release](https://github.com/pgsty/silo-pkg/releases/tag/v3.14.1) ·
[Changes since v3.14.0](https://github.com/pgsty/silo-pkg/compare/v3.14.0...v3.14.1)

- Pin upstream minio-go to `v7.3.1-0.20260915093545-32e1f32cb176`, fixing
  CopyObject responses that report an S3 error inside an HTTP 200 response.
  The SDK now retries and reports the error instead of returning success.
- Update JWX from v3.2.0 to v3.3.0, which JSON-escapes custom claim, header
  and JWK field names on output (GHSA-4cf7-xm37-g63h). Its test dependency
  advances testify to v1.12.1.
- Migrate the lint configuration to `gomodguard_v2` while retaining
  golangci-lint v2.13.1.
- Retain the Go 1.26 compatibility floor, Go 1.27.1 toolchain and
  go-systemd v22.6.0 NetBSD compatibility replacement. Public Go signatures
  and the password-policy semantics introduced in v3.14.0 are unchanged.

## v3.14.0 — 2026-09-13

Published from `827f8109ff11bf6239a35d8d6d137cb5738539c3`.
[GitHub release](https://github.com/pgsty/silo-pkg/releases/tag/v3.14.0) ·
[Changes since v3.13.3](https://github.com/pgsty/silo-pkg/compare/v3.13.3...v3.14.0)

- **Breaking policy semantics:** `Policy.IsAllowedActions` exposes
  `admin:ChangeMyPassword` unless explicitly denied; `admin:CreateUser` requires
  an explicit Allow. The built-in `readonly` drops its CreateUser deny; the new
  `consolereadonly` also grants bucket listing and follows the permission split.
  Neither policy independently grants user administration. See
  [the migration notes](UPSTREAM.md#breaking-authorization-compatibility).
- Pin upstream minio-go to `v7.3.1-0.20260910142817-60bd07042d49`, incorporating
  upload-limit, streaming Content-Type signing, RDMA TLS trust, listing checksum
  and restore-status fixes. Refresh the Go x/* dependencies and govulncheck 1.8.0.
- Retain Go 1.26 as the library floor, toolchain Go 1.27.1, unchanged public Go
  signatures, and the go-systemd v22.6.0 NetBSD compatibility replacement.
- Validation includes full race suites with Go 1.26.8 and 1.27.1, lint, LDAP
  configuration validation and vulnerability scanning. No reachable or imported
  vulnerable package was reported. GO-2026-5932 remains a module-only advisory
  in unused OpenPGP code; this is not a claim that the module graph has no CVEs.

As of 2026-09-13 the matching Server and Console changes are on `main`, but
**no new Server or Console release has shipped them**. Server 20260903 and
Console v2.4.0 still use the prior password-permission mapping. Installing the
new mcli alone does not change server-side authorization. See the
[current component matrix](https://silo.pgsty.com/compatibility/versions/).

Earlier releases are preserved in the [GitHub release archive](https://github.com/pgsty/silo-pkg/releases).

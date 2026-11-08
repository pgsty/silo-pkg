# Changelog

## v3.16.0 — 2026-11-08

- Update go-jose/v4 from v4.1.6 to v4.1.9, including the five security fixes
  in [v4.1.7](https://github.com/go-jose/go-jose/releases/tag/v4.1.7) and the
  subsequent OpaqueSigner and unsupported-key error fixes. These are dependency
  fixes; their applicability depends on the consuming application's use of JOSE.
- Publish the networking dependency updates already on main: x/net v0.62.0 and
  x/crypto `v0.58.1-0.20261009181718-2f1f834330b2`, replacing v0.61.0 and
  v0.58.0 respectively.
- Update cpuid/v2 to v2.4.2 to correct XSAVE/OSXSAVE feature detection, and
  go.yaml.in/yaml/v3 to v3.0.6 to quote literal `<<` keys instead of emitting
  ambiguous YAML merge keys. Cover configuration save/load with a regression test.
- Clarify that the built-in infraAdmin policy includes batch-job administration.
  Public package APIs, policy formats and authorization decisions remain compatible.
- Retain Go 1.27.3, the existing upstream minio-go commit `320af050c445`,
  go-humanize v1.0.1, go-runewidth v0.0.29, and the go-systemd v22.6.0
  compatibility replacement. Consumers must retain their own required replacements.
- This is a source-only module release. Server, Console and mcli must select
  v3.16.0 and rebuild; this release does not publish their binaries or images.

## v3.15.2 — 2026-10-26

- Update upstream minio-go to `v7.4.2-0.20261025165926-320af050c445`.
  Removing bucket replication now accepts HTTP 204 as well as HTTP 200,
  including an empty configuration passed to SetBucketReplication. Other
  responses still return their original error. Upstream has not tagged v7.4.2.
- Keep Go 1.27.3, public APIs, policy semantics, and the verified output and
  NetBSD compatibility selections. This is a protocol-correctness update,
  with no newly claimed CVE fix.
- Consumers must adopt v3.15.2 and rebuild; this source-only module release
  does not update installed mcli, Console or SILO binaries.

## v3.15.1 — 2026-10-25

[GitHub release](https://github.com/pgsty/silo-pkg/releases/tag/v3.15.1) ·
[Changes since v3.15.0](https://github.com/pgsty/silo-pkg/compare/v3.15.0...v3.15.1)

- Update upstream minio-go from v7.4.1 to
  `v7.4.2-0.20261020194015-ab93db7ba435`. This exact upstream commit includes
  six fixes after v7.4.1; upstream has not published a v7.4.2 tag.
- Reject short streamed multipart uploads before sending the incomplete part,
  including parallel streams. Preserve sequential reads after `ReadAt` reaches
  EOF, reject negative `ReadAt` offsets before fetching metadata, and preserve
  read/seek bounds after `Stat`. Report each failed `RemoveObjects` object once.
- Retain Go 1.27.3, public package APIs, policy semantics and the v3.14.3 policy
  optimizations. Other declared dependencies are current except the deliberate
  go-humanize v1.0.1, go-runewidth v0.0.29 and go-systemd v22.6.0 compatibility
  selections. This release does not add a new CVE fix.
- This is a source-only module release. Consumers must adopt v3.15.1 and rebuild
  to receive the SDK fixes; SILO, Console and mcli publish independently.

## v3.15.0 — 2026-10-24

[GitHub release](https://github.com/pgsty/silo-pkg/releases/tag/v3.15.0) ·
[Changes since v3.14.3](https://github.com/pgsty/silo-pkg/compare/v3.14.3...v3.15.0)

- Require **Go 1.27.3** for library consumers and development, replacing the
  Go 1.26 library floor. Consumers must rebuild with the updated compiler to
  receive its certificate-trust fix for CVE-2026-111943 / GO-2026-6809.
- Update JWX to v3.4.0, all three etcd client/API modules to v3.7.3, gRPC to
  v1.84.0, grpc-gateway to v2.32.0, and goccy/go-json to v0.11.2. Refresh
  swag/conv, go-colorable, cpuid, fwd, and Google API/RPC generated modules.
- Preserve public package APIs, policy formats and authorization semantics,
  including the v3.14.3 policy fast paths. Keep go-humanize v1.0.1 and
  go-runewidth v0.0.29 for CLI output compatibility, and the go-systemd
  v22.6.0 replacement for NetBSD. Go does not inherit dependency replacements;
  consuming applications must retain the replacements they need.
- Pin development and CI to Go 1.27.3 and golangci-lint v2.15.0. This is a
  source-only module release; installed SILO, Console and mcli binaries are
  updated only when their own releases adopt the module and rebuild.
- GO-2026-5932 remains an upstream advisory without a fixed version in the
  deprecated OpenPGP packages. Those packages are not imported by this module.

## v3.14.3 — 2026-10-18

Published from `e9d585ba1e7a30ee5eb85b992d883820b83768f5`.
[GitHub release](https://github.com/pgsty/silo-pkg/releases/tag/v3.14.3) ·
[English notes](https://silo.pgsty.com/blog/release/pkg-3.14.3/) ·
[中文说明](https://silo.pgsty.com/zh/blog/release/pkg-3.14.3/)

- Add exact-lookup and wildcard-prefix fast paths to `AdminAction.IsValid`
  and `HasResource` (#11), preserving wildcard semantics and bucket-scoped
  resource checks. Add equivalence tests, fuzzing and policy benchmarks.
- Preserve APIs, policy formats, authorization decisions, dependencies and
  the Go 1.26 library floor. Existing mcli/Console artifacts remain on v3.14.2;
  consumers receive the optimization only after updating and rebuilding.
- Release validation includes Go 1.26/1.27 race tests, lint, LDAP, vulnerability
  scanning, equivalence fuzzing and independent policy comparisons. The reported
  82% reduction in one policy microbenchmark is not an S3 throughput claim.

Thanks to Denis Babchuk (@botdeniska) and original author Klaus Post (@klauspost).

## v3.14.2 — 2026-10-15

[GitHub release](https://github.com/pgsty/silo-pkg/releases/tag/v3.14.2) ·
[Changes since v3.14.1](https://github.com/pgsty/silo-pkg/compare/v3.14.1...v3.14.2)

- Update upstream minio-go to v7.4.1. Failed multipart uploads are cleaned up
  even after cancellation; retention timestamps are sent in UTC; STS duration
  requests are preserved; projected Kubernetes token files are supported.
  Parallel GET and ComposeObject remain opt-in.
- Update go-jose to v4.1.6 for the RSA JWK parsing denial of service fix
  [GHSA-mgcc-cww5-fp2q](https://github.com/go-jose/go-jose/security/advisories/GHSA-mgcc-cww5-fp2q).
  JWK parsing now limits RSA keys to 4096 bits by default. Consumers using
  larger keys must review their identity-provider configuration before upgrading.
- Update LDAP to v3.4.15, msgp to v1.6.5 and the etcd client modules to v3.7.2.
  Refresh x/crypto to v0.58.0, x/net to v0.61.0, x/sys to v0.49.0 and x/text
  to v0.43.0. The x/net update includes the October HTTP/2 security fixes.
- Use Go 1.27.2, golangci-lint v2.14.0 and govulncheck v1.9.0 for development.
  Retain the Go 1.26 library floor and go-systemd v22.6.0 NetBSD replacement.
  Public SILO package signatures and password-policy semantics are unchanged.

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

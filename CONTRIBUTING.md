# Contributing to SILO Shared Packages

Submit issues and pull requests to [pgsty/silo-pkg](https://github.com/pgsty/silo-pkg).
This repository maintains `github.com/pgsty/silo-pkg/v3` for the SILO server,
Console, and mcli client. Compatibility with unmodified upstream MinIO is best
effort; coordinate breaking changes with the maintained SILO components.

## Licensing of Contributions

Code contributions are accepted under [GNU AGPL v3.0 or later](LICENSE)
(`AGPL-3.0-or-later`), the same license as SILO Shared Packages.

- **No CLA.** Contributors retain copyright in their original work. No
  Contributor License Agreement, copyright assignment, or separate Apache-2.0
  license grant to SILO or upstream MinIO maintainers is required. Contributions
  are accepted inbound=outbound; maintainers receive no rights beyond the
  applicable project license.
- **DCO sign-off.** Sign each commit with `git commit -s` to certify the
  [Developer Certificate of Origin 1.1](https://developercertificate.org/).
  Preserve sign-off trailers when squashing commits.
- **Provenance and notices.** Preserve original authorship, copyright, and
  license notices when importing or modifying existing code. New original files
  name their actual copyright holders and use AGPL-3.0-or-later. Separately
  licensed third-party material keeps its existing license and attribution.

## Development and Validation

1. Fork this repository and create a topic branch from `main`.
2. Add focused tests for changes in behavior and format Go code with `gofmt`.
3. Run `make test`, which runs lint and package tests with the race detector.
   LDAP integration tests require a test server configured through
   `LDAP_TEST_SERVER`; use `make test-ldap` when changing LDAP behavior.
4. Commit with a concise message and DCO sign-off, then open a pull request
   against `pgsty/silo-pkg:main`.

Include the motivation, affected packages, validation results, and any impact
on the SILO server, Console, or mcli. Public documentation changes belong in
[pgsty/silo.pgsty.com](https://github.com/pgsty/silo.pgsty.com).

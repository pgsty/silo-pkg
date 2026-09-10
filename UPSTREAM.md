# Upstream review: 2026-09-10

Reviewed `minio/pkg` through
[`657d87e88168`](https://github.com/minio/pkg/commit/657d87e881680db68e3a9f9286569788651c968e)
and `minio/minio-go` through
[`78bfa91607c2`](https://github.com/minio/minio-go/commit/78bfa91607c2b9a7eb1ad5fe739de3c6c238d5ab).

## Adopted

- [pkg #262](https://github.com/minio/pkg/pull/262): report
  `admin:ChangeMyPassword` as an implicit self-service capability, subject to
  explicit Deny. `admin:CreateUser` requires an explicit Allow. Remove the
  legacy CreateUser deny from the built-in read-only policy.
- [pkg #233](https://github.com/minio/pkg/pull/233): add `consolereadonly`,
  with GetObject, GetBucketLocation and ListBucket. Apply #262 to this new
  policy too. The original `readonly` S3 permissions remain unchanged.
- Pin the upstream SDK to `v7.3.1-0.20260909183557-78bfa91607c2`. Relative to
  the previous `0e78d3f18efe` pin, it contains configurable upload limits
  ([#2299](https://github.com/minio/minio-go/pull/2299)), Content-Type in
  streaming SignedHeaders ([#2301](https://github.com/minio/minio-go/pull/2301))
  and caller TLS trust on RDMA ([#2302](https://github.com/minio/minio-go/pull/2302)).

## Breaking authorization compatibility

Adopting #262 changes existing policy semantics; it is independent of updating
the minio-go SDK. This must be disclosed as a breaking change in the release
that includes it. The Go signatures and Go compatibility floor are unchanged,
but the public `Policy.IsAllowedActions` method returns different capabilities:
ChangeMyPassword is implicit unless denied, and CreateUser requires an explicit
Allow. Consumers must use the matching capability for each operation.

With the matching Server change, a saved `Deny admin:CreateUser` no longer
prevents the caller from changing their own password. A saved
`Deny admin:ChangeMyPassword` now prevents it. To preserve a policy's old
combined restriction, add ChangeMyPassword to the same CreateUser Deny statement
before upgrading, preserving its other actions, scope and conditions. Saved
policy documents are not rewritten automatically.

The built-in `readonly` policy also drops its old CreateUser deny. It now allows
self-service password changes, and a separate CreateUser Allow can grant user
administration where the old built-in deny overrode it. Saved overrides retain
their existing deny; inspect the effective policy contents. `consolereadonly`
is new and follows the split. Neither read-only policy grants user
administration on its own.

Deploy password authorization with the matching SILO Server and Console.
During a mixed-version rollout or rollback, retain both denies if the old
combined restriction must hold: an old Server does not enforce a password-only
deny for this endpoint. See the Server's
[password-permission migration guide](https://github.com/pgsty/silo/blob/420340bc142b7dec00c26c28dd78102e3ed9d0f3/docs/iam/password-permissions.md)
for the before/after matrix, policy migration, read-only composition and
rollback limits.

## Already covered or deferred

| Upstream work | SILO decision |
| --- | --- |
| #265, x/crypto v0.56.0 | Already selected; no additional version bump. |
| #230, RNG subkey initialization | Already fixed locally. |
| #226, exact condition key lookup | Already fixed locally. |
| #242, xtime.Duration JSON marshaling | Already implemented locally. |
| #261, NotResource deduplication, literal-policy Deny detection and wildcard backtracking | Correctness fixes already present in SILO v3.13.3, including condition-value boundary tests; retain them. The remaining indexing and classification refactor is a separate optimization. |
| #249–252, ARN and Memory resource changes | Keep SILO's strict administration writes and compatible historical-policy parsing; do not replace them with the AIStor Memory model. |
| #245–246, typed action API | Breaking source changes without a current SILO requirement; defer. |
| Memory, S3Tables, annotation, compression and new closed-product admin actions | No matching maintained SILO endpoints in this update; defer. |
| #263, bounded deduplicating channel | New helper with no current SILO consumer; defer. |
| Certificate-cache accessors, bandwidth configuration and RNG assembly | Independent features or optimizations; defer until a consumer or measured need justifies them. |

[minio-go #2274](https://github.com/minio/minio-go/pull/2274), trimming bucket
location whitespace, is still open at this review. It is absent from the pinned
commit. Take it through a later upstream commit after merge; do not revive the
retired silo-go fork for this other-S3 compatibility issue. It does not block a
SILO release.

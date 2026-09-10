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

Password authorization must be deployed with the matching SILO Server and
Console changes. See the Server's `docs/iam/password-permissions.md` for old
policy behavior and migration. No public API or Go compatibility floor changes
are required by these package changes.

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

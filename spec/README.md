# Definitions and implementation evidence

The [library specification](../doc/diameterkit-specification.md) is the normative
behavior record. This directory contains versioned generator inputs, requirement
progress and reproducible acceptance evidence; it does not introduce a second set
of requirement IDs or an OpenSpec workflow.

Start with the English [support matrix](../doc/diameterkit-support-matrix.md),
[current remaining gaps](../doc/remaining-gaps.md),
[definition guide](../doc/diameterkit-application-definitions.md),
[version policy](../doc/version-install-policy.md) and
[publication record](../doc/releases/v0.1.0-rc.md).

- [progress.json](progress.json): current batch status, 24 requirements, source
  hashes, checks and remaining gates. A completed bounded batch is not complete v1.
- [history](history): frozen previous progress records; historical paths, language
  and measurements describe the state at that time.
- [evidence](evidence): detailed protocol, consumer, release and failure evidence.
- [public API](release/public-api.json): reviewed documentation/source snapshot,
  not automatic semantic compatibility certification.
- [support matrix](release/support-matrix.json): explicit profile/platform scope.
- [support matrix checks](../doc/support-matrix-checks.md): executable current-definition consistency gate.
- [version policy](release/version-policy.json): independent module/API rules.
- [performance baseline](release/performance-baseline.json): measured samples and
  environment; no inferred production throughput claim.

The selected implementation includes 11 generated definitions and one delegated
profile, with overlapping command bindings counted separately. Do not infer full
3GPP/RFC coverage from a supported Application-ID or profile name. Private template
mappings, target OCS/DRA/HSS acceptance and application-owned durable workflows
retain independent evidence gates.

Process notes `batch-*.md` stay under ignored `spec/local/batches/`; they are not
published. Build logs and caches stay under ignored `.cache/`. Public progress
uses JSON evidence rather than depending on local process notes. Historical
technical references retain their original language and are not silently rewritten.

## Benchmarks

Use `python3 spec/tools/release_check.py --suite bench`. See the
[methodology](../doc/load-checks.md) and final [README table](../README.md#benchmarks).

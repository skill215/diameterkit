# Protocol and interoperability evidence

This directory contains project-owned comparison/fixture tooling. It is development
infrastructure, not a production dependency or an automatically trusted external
Diameter endpoint. See [AI disclosure](../AI_DISCLOSURE.md).

The repository distinguishes wire/definition validation, paired local runtime
checks, public consumer examples and actual target interoperability. Passing one
does not silently establish the others. Recorded fixtures and old dictionary
comparisons have named inputs and limitations in [spec/evidence](../spec/evidence)
and [history](../spec/history). No new live DRA/OCS/HSS claim is made by publishing
the library or translating this guide.

For current users, start with the [public examples](../README.md#examples),
[support matrix](../doc/diameterkit-support-matrix.md) and
[version/install policy](../doc/version-install-policy.md). Detailed original
protocol, template and migration audits remain in the referenced historical
records and technical documents; their original language and measurements are
preserved. Only named selected bindings are supported, with explicit endpoint,
role, source, direction and recovery evidence.

Run the [dual-module acceptance entry point](../doc/release-checks.md) for generated
consistency, independent vectors, pure/race runtime execution and finite load
checks. The published-module gate separately verifies real authenticated downloads
into empty external caches with no workspace/replacement. Neither gate creates a
subscriber service, financial ledger, delivery engine or persistence boundary.

## Benchmarks

Measurements and repeatable commands are at the end of the
[README](../README.md#benchmarks). Loopback results do not establish external
network throughput or independent-stack interoperability.

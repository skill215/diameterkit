# Dual-module checkout acceptance

The [runner](../spec/tools/release_check.py) and [CI workflow](../.github/workflows/ci.yml)
validate the current checkout. Published-module installation is a separate
[gate](version-install-policy.md); the root test command does not include a nested module.

Use the exact [.go-version](../.go-version) toolchain, Python 3, a C compiler for
the Go race tool, and a Linux/amd64 kernel with native SCTP support.

```sh
python3 spec/tools/release_check.py --suite all
# Independently runnable suites:
python3 spec/tools/release_check.py --suite static
python3 spec/tools/release_check.py --suite pure
python3 spec/tools/release_check.py --suite race
python3 spec/tools/release_check.py --suite cross
python3 spec/tools/release_check.py --suite fuzz
python3 spec/tools/release_check.py --suite stress
python3 spec/tools/release_check.py --suite bench
```

| Suite | Acceptance |
| --- | --- |
| static | Tool negative controls, 11 generated outputs and determinism, Linux/Windows public API snapshots, both modules' vet, AI header checks |
| pure | Both modules' tests with CGO=0, count=1 and no skipped native SCTP tests |
| race | Both modules' tests with race instrumentation; production libraries remain cgo-independent |
| cross | CGO=0 builds for linux/arm64, linux/386 and windows/amd64; compilation only |
| fuzz | Twelve targets, 30 seconds and two workers per target |
| stress | Fixed TCP/SCTP workloads repeated 20 times in pure and race modes |
| bench | Eight workloads, 200 iterations per sample, five samples |

Results go to `.cache/release-check/<suite>/`; `--output` chooses another directory.
Failures return nonzero with command logs and hashes. Missing SCTP support is a
failure, not a skip. No-test packages do not count as skipped tests. Generated/API
negative controls must reject drift without rewriting their input.

The runner creates a temporary workspace containing both checkout modules. A
version-specific core replacement explicitly selects the current source; it does
not represent an installable dependency. The graph must contain exactly the two
checkout modules and pinned SCTP upstream, with no upstream replacement. Temporary
workspace files are removed; module manifests, sums and the API baseline are protected.

CI has seven Ubuntu jobs, read-only repository permissions, pinned Action commits,
a 20-minute limit per suite and 14-day success/failure artifacts. A new run on the
same ref may cancel an older run. Branch protection is configured separately.
Actual hosted outcomes are linked in the [release record](releases/v0.1.0-rc.md).

## Benchmarks

See [workloads, resource meanings and raw measurements](load-checks.md). Finite
benchmark samples are not production capacity, latency percentiles or a speedup claim.
